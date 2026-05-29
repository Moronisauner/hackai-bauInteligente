package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AccountContext é o resumo de uma conta-fonte que o assistente pode usar ao
// montar a proposta. Saldo formatado como decimal (ex: "1234.56").
type AccountContext struct {
	AccountID string
	BrandName string
	Type      string
	Number    string
	Balance   string
}

// ProposedAllocation é uma alocação sugerida pelo assistente para uma conta.
type ProposedAllocation struct {
	AccountID  string `json:"account_id"`
	Percentage int    `json:"percentage"`
	Reason     string `json:"reason"`
}

// Plan é uma variação de plano para o mesmo objetivo (mesmas contas): difere no
// valor-alvo e/ou no prazo, mudando o ritmo mensal. O assistente oferece de 1 a
// 3 opções entre as quais o cliente escolhe.
type Plan struct {
	Label          string `json:"label"`   // ex: "Equilibrado", "Mais rápido"
	Summary        string `json:"summary"` // frase curta com o trade-off
	TargetAmount   string `json:"target_amount"`
	DurationMonths int    `json:"duration_months"`
}

// Proposal agrega o objetivo proposto: nome, alocações (contas-fonte, comuns a
// todas as variações) e as opções de plano entre as quais o cliente escolhe.
type Proposal struct {
	Name        string               `json:"name"`
	StartDate   string               `json:"start_date"`
	Allocations []ProposedAllocation `json:"allocations"`
	Plans       []Plan               `json:"plans"`
}

// Turn é o resultado de um turno do assistente: a fala para o usuário e, quando
// a conversa reuniu informação suficiente, uma proposta pronta para confirmar.
type Turn struct {
	Reply    string
	Done     bool
	Proposal *Proposal
}

// defaultReply é a fala usada quando o modelo não devolve um reply utilizável.
const defaultReply = "Pode me contar um pouco mais?"

// maxTokens limita a geração. A resposta pode trazer 3 planos + alocações + a
// fala, então damos folga (o provedor é rápido).
const maxTokens = 900

// Plan roda um turno da conversa em UMA única chamada ao modelo (latência em CPU
// é cara — duas chamadas por turno é inviável). O modelo faz slot-filling: extrai
// o que o cliente já disse (nome, valor, prazo) e devolve a próxima pergunta.
//
// Quem DECIDE se o plano está pronto é o backend, deterministicamente (os três
// slots preenchidos) — não um flag do modelo, que oscila. Quando pronto, as
// alocações sugeridas pelo modelo são validadas; se ele não sugerir nenhuma
// válida, um fallback determinístico (contas de maior saldo) garante uma proposta
// sempre consistente. Contas vão ao modelo por ÍNDICE (1, 2, 3…), nunca pelo
// account_id real (hash de 64 chars): encolhe prompt/saída e evita corromper o id.
func (c *Client) Plan(ctx context.Context, accounts []AccountContext, refDate time.Time, history []Message) (Turn, error) {
	msgs := make([]Message, 0, len(history)+1)
	msgs = append(msgs, Message{Role: RoleSystem, Content: advisorPrompt(accounts)})
	msgs = append(msgs, history...)

	raw, err := c.complete(ctx, msgs, true, maxTokens)
	if err != nil {
		return Turn{}, err
	}

	var wire struct {
		Name        string          `json:"name"`
		Propose     json.RawMessage `json:"propose"`
		Reply       string          `json:"reply"`
		Allocations []struct {
			Account    json.RawMessage `json:"account"`
			Percentage json.RawMessage `json:"percentage"`
			Reason     string          `json:"reason"`
		} `json:"allocations"`
		Plans []struct {
			Label          string          `json:"label"`
			Summary        string          `json:"summary"`
			TargetAmount   json.RawMessage `json:"target_amount"`
			DurationMonths json.RawMessage `json:"duration_months"`
		} `json:"plans"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &wire); err != nil {
		return Turn{Reply: defaultReply}, nil
	}

	reply := strings.TrimSpace(wire.Reply)
	if reply == "" {
		reply = defaultReply
	}

	name := strings.TrimSpace(wire.Name)

	// Quem decide apresentar planos é a IA (propose), agindo como consultora: só
	// propõe quando reuniu o necessário E o cliente está de acordo com o ritmo.
	// Enquanto isso (ou se ainda não há nome), apenas conversa/aconselha.
	if !parseBoolLoose(wire.Propose) || name == "" {
		return Turn{Reply: reply}, nil
	}

	// Alocações da IA (índice→id), comuns a todas as variações; fallback se vazio.
	allocs := make([]ProposedAllocation, 0, len(wire.Allocations))
	for _, a := range wire.Allocations {
		id := resolveAccount(a.Account, accounts)
		if id == "" {
			continue
		}
		allocs = append(allocs, ProposedAllocation{AccountID: id, Percentage: parseIntLoose(a.Percentage), Reason: a.Reason})
	}
	if len(allocs) == 0 {
		allocs = fallbackAllocations(accounts)
	}

	// Opções de plano sugeridas pela IA.
	plans := make([]Plan, 0, len(wire.Plans))
	for _, pl := range wire.Plans {
		plans = append(plans, Plan{
			Label:          strings.TrimSpace(pl.Label),
			Summary:        strings.TrimSpace(pl.Summary),
			TargetAmount:   parseDecimalLoose(pl.TargetAmount),
			DurationMonths: parseIntLoose(pl.DurationMonths),
		})
	}

	p := sanitizeProposal(&Proposal{Name: name, Allocations: allocs, Plans: plans}, accounts, refDate)
	if p == nil {
		// Sem plano/alocação válidos (ex: cliente sem contas) — segue conversando.
		return Turn{Reply: reply}, nil
	}
	// A fala da IA acompanha os cards: explica as opções e convida a ajustar.
	return Turn{Reply: reply, Done: true, Proposal: p}, nil
}

// fallbackAllocations monta um plano padrão quando o modelo não sugere alocações
// válidas: as contas de maior saldo (a lista já chega ordenada desc), com
// percentuais decrescentes. Garante que uma proposta pronta nunca fique sem
// alocação por falha do modelo.
func fallbackAllocations(accounts []AccountContext) []ProposedAllocation {
	pcts := []int{40, 30, 20}
	out := make([]ProposedAllocation, 0, len(pcts))
	for i, a := range accounts {
		if i >= len(pcts) {
			break
		}
		out = append(out, ProposedAllocation{
			AccountID:  a.AccountID,
			Percentage: pcts[i],
			Reason:     "Conta com maior saldo disponível",
		})
	}
	return out
}

// resolveAccount converte o "account" do modelo no account_id real. Aceita o
// índice 1-based da lista (caminho normal) e, como tolerância, um account_id
// cru caso o modelo decida ecoá-lo.
func resolveAccount(raw json.RawMessage, accounts []AccountContext) string {
	if idx := parseIntLoose(raw); idx >= 1 && idx <= len(accounts) {
		return accounts[idx-1].AccountID
	}
	s := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	for _, a := range accounts {
		if a.AccountID == s {
			return s
		}
	}
	return ""
}

// advisorPrompt instrui o turno único. A IA age como CONSULTORA, não calculadora:
// faz perguntas relevantes (inclusive sobre o perfil financeiro), avalia a
// viabilidade do ritmo mensal e, ao propor, oferece de 2 a 3 OPÇÕES de plano.
// Contas vão por ÍNDICE numérico (nunca o account_id de 64 chars); se não sugerir
// alocações válidas, o backend cai no fallback determinístico.
func advisorPrompt(accounts []AccountContext) string {
	var b strings.Builder
	b.WriteString("Você é um consultor financeiro amigável que ajuda o cliente a montar um objetivo de poupança (uma \"conta baú\"). ")
	b.WriteString("Converse em português do Brasil, com frases curtas e tom acolhedor. Você NÃO é uma calculadora: faça perguntas relevantes, dê conselhos e explique as opções.\n\n")
	b.WriteString("Descubra ao longo da conversa: nome do objetivo (name), valor-alvo em reais (target_amount) e prazo em meses (duration_months). ")
	b.WriteString("A cada turno, preencha os que já souber (null no resto) e conduza a conversa em \"reply\".\n\n")
	b.WriteString("PERGUNTAS DE OTIMIZAÇÃO: faça NO MÁXIMO UMA pergunta que ajude a calibrar a meta/ritmo, por exemplo: ")
	b.WriteString("\"em que época do ano você fatura mais?\", \"qual é o seu tipo de negócio?\", \"qual é o seu maior receio financeiro?\". Nunca repita uma pergunta já respondida.\n\n")
	b.WriteString("VIABILIDADE: quando tiver valor e prazo, calcule o ritmo mensal = valor ÷ prazo. Se for alto/exigente, comente em reais quanto seria por mês e ofereça alternativas (esticar o prazo, reduzir a meta).\n\n")
	b.WriteString("QUANDO PROPOR (regra firme): assim que tiver o objetivo e o valor E (o prazo OU quanto o cliente consegue guardar por mês) — ou se o cliente pedir para ver as opções/planos — PROPONHA AGORA (\"propose\": true) com 2 a 3 planos. NÃO faça mais perguntas nesse momento. ")
	b.WriteString("Se o cliente disse quanto consegue por mês, derive o prazo de um dos planos (prazo ≈ valor ÷ mensal) e varie os outros em torno disso.\n\n")
	b.WriteString("CONTAS: você pode perguntar quais contas o cliente prefere usar (a lista está no fim). Se ele não disser, escolha as de maior saldo.\n\n")
	b.WriteString("PROPOR (\"propose\": true) só quando tiver objetivo, valor/prazo de base E o cliente estiver de acordo com o ritmo. Ao propor:\n")
	b.WriteString("- \"allocations\": 1 a 3 contas (percentual inteiro 1–100, maior nas de maior saldo). Vale para todas as opções.\n")
	b.WriteString("- \"plans\": 2 a 3 OPÇÕES do mesmo objetivo com trade-offs diferentes. Cada uma: {\"label\":\"...\",\"summary\":\"...\",\"target_amount\":\"...\",\"duration_months\":N}. ")
	b.WriteString("Ex: uma \"Equilibrado\", uma \"Mais rápido\" (mais por mês, menos meses) e uma \"Mais tranquilo\" (menos por mês, mais meses). No summary diga o ritmo mensal aproximado.\n")
	b.WriteString("- Em \"reply\", apresente as opções em uma frase e convide a escolher/ajustar.\n")
	b.WriteString("Enquanto não for propor: \"propose\": false, \"allocations\": [], \"plans\": [].\n\n")
	b.WriteString("Regras: valores são string decimal (\"60 mil\"→\"60000.00\"). prazos são inteiros (\"3 meses\"→3, \"2 anos\"→24). \"account\" é o NÚMERO da conta na lista.\n")
	b.WriteString("Responda SEMPRE só com um JSON com as chaves: name, target_amount, duration_months, propose, allocations, plans, reply.\n\n")
	b.WriteString("Exemplos:\n")
	b.WriteString("Cliente: \"quero comprar meu Celta de 60 mil em 3 meses\"\n")
	b.WriteString("→ {\"name\":\"Celta\",\"target_amount\":\"60000.00\",\"duration_months\":3,\"propose\":false,\"allocations\":[],\"plans\":[],\"reply\":\"Pra juntar R$60 mil em 3 meses dá ~R$20 mil/mês — é bastante! Me conta: em que época do ano você fatura mais?\"}\n")
	b.WriteString("Cliente: \"sou autônomo, faturo mais no fim do ano; consigo uns 3 mil/mês\"\n")
	b.WriteString("→ {\"name\":\"Celta\",\"target_amount\":\"60000.00\",\"duration_months\":20,\"propose\":true,\"allocations\":[{\"account\":1,\"percentage\":50,\"reason\":\"maior saldo\"},{\"account\":2,\"percentage\":30,\"reason\":\"reforço\"}],\"plans\":[{\"label\":\"Equilibrado\",\"summary\":\"~R$3 mil/mês\",\"target_amount\":\"60000.00\",\"duration_months\":20},{\"label\":\"Mais rápido\",\"summary\":\"~R$5 mil/mês, chega antes\",\"target_amount\":\"60000.00\",\"duration_months\":12},{\"label\":\"Mais tranquilo\",\"summary\":\"~R$2 mil/mês\",\"target_amount\":\"60000.00\",\"duration_months\":30}],\"reply\":\"Montei 3 opções pra você: equilibrada (~R$3 mil/mês em 20 meses), mais rápida ou mais tranquila. Qual combina mais?\"}\n\n")
	b.WriteString("Contas (use o número em \"account\"):\n")
	for i, a := range accounts {
		b.WriteString(fmt.Sprintf("%d) %s — saldo %s\n", i+1, a.BrandName, a.Balance))
	}
	return b.String()
}

// extractJSON recorta o primeiro objeto JSON de um texto que pode vir cercado
// por prosa ou cercas de código.
func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.IndexByte(s, '{'); i >= 0 {
		s = s[i:]
	}
	if j := strings.LastIndexByte(s, '}'); j >= 0 {
		s = s[:j+1]
	}
	return s
}

// sanitizeProposal valida e normaliza a proposta contra as contas reais:
// descarta alocações inválidas (conta inexistente, percentual fora de 1..100,
// duplicada) e planos inválidos (valor-alvo <= 0), aplica defaults e limita o
// prazo a 1..60 meses. Retorna nil se não sobrar alocação OU plano válido.
func sanitizeProposal(p *Proposal, accounts []AccountContext, refDate time.Time) *Proposal {
	valid := make(map[string]bool, len(accounts))
	for _, a := range accounts {
		valid[a.AccountID] = true
	}

	seen := make(map[string]bool)
	cleanAllocs := make([]ProposedAllocation, 0, len(p.Allocations))
	for _, a := range p.Allocations {
		if !valid[a.AccountID] || seen[a.AccountID] {
			continue
		}
		if a.Percentage < 1 || a.Percentage > 100 {
			continue
		}
		seen[a.AccountID] = true
		cleanAllocs = append(cleanAllocs, a)
	}
	if len(cleanAllocs) == 0 {
		return nil
	}
	p.Allocations = cleanAllocs

	cleanPlans := make([]Plan, 0, len(p.Plans))
	for _, pl := range p.Plans {
		// Sem valor-alvo válido (> 0) o plano não pode virar objetivo (RF-03).
		if f, err := strconv.ParseFloat(pl.TargetAmount, 64); err != nil || f <= 0 {
			continue
		}
		if pl.DurationMonths < 1 {
			pl.DurationMonths = 1
		} else if pl.DurationMonths > 60 {
			pl.DurationMonths = 60
		}
		if strings.TrimSpace(pl.Label) == "" {
			pl.Label = "Plano"
		}
		cleanPlans = append(cleanPlans, pl)
	}
	if len(cleanPlans) == 0 {
		return nil
	}
	p.Plans = cleanPlans

	if strings.TrimSpace(p.Name) == "" {
		p.Name = "Meu objetivo"
	}
	// O modelo não é confiável com datas — o início é sempre a data de referência.
	p.StartDate = refDate.Format("2006-01-02")
	return p
}

// parseIntLoose lê um inteiro de um JSON que pode ser número ("30") ou string ("30").
func parseIntLoose(raw json.RawMessage) int {
	s := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if s == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}
	return 0
}

// parseBoolLoose lê um bool de um JSON que pode ser true/false ou "true"/"false".
func parseBoolLoose(raw json.RawMessage) bool {
	s := strings.Trim(strings.TrimSpace(strings.ToLower(string(raw))), `"`)
	return s == "true"
}

// parseDecimalLoose lê um valor monetário de um JSON número ou string, devolvendo
// um decimal canônico com 2 casas (ex: "10000.00"). "" se não parsear.
func parseDecimalLoose(raw json.RawMessage) string {
	s := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if s == "" {
		return ""
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(f, 'f', 2, 64)
}

package assistant

import (
	"encoding/json"
	"testing"
	"time"
)

var refDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func testAccounts() []AccountContext {
	return []AccountContext{
		{AccountID: "acc-1", BrandName: "Banco A", Balance: "5000.00"},
		{AccountID: "acc-2", BrandName: "Banco B", Balance: "1200.00"},
	}
}

// resolveAccount mapeia o índice 1-based do modelo para o account_id real.
func TestResolveAccount(t *testing.T) {
	accs := testAccounts()
	if got := resolveAccount(json.RawMessage(`1`), accs); got != "acc-1" {
		t.Fatalf("índice 1 → %q", got)
	}
	if got := resolveAccount(json.RawMessage(`"2"`), accs); got != "acc-2" {
		t.Fatalf("índice string \"2\" → %q", got)
	}
	if got := resolveAccount(json.RawMessage(`9`), accs); got != "" {
		t.Fatalf("índice fora de faixa deveria ser vazio, got %q", got)
	}
	// Tolerância: account_id cru ecoado pelo modelo.
	if got := resolveAccount(json.RawMessage(`"acc-2"`), accs); got != "acc-2" {
		t.Fatalf("account_id cru → %q", got)
	}
}

// fallbackAllocations cobre o caso de o modelo não sugerir alocações válidas.
func TestFallbackAllocations(t *testing.T) {
	out := fallbackAllocations(testAccounts())
	if len(out) != 2 || out[0].AccountID != "acc-1" || out[0].Percentage != 40 {
		t.Fatalf("fallback inesperado: %+v", out)
	}
}

func TestSanitizeProposal_DropsInvalidAndDefaults(t *testing.T) {
	p := &Proposal{
		Name:      "Viagem",
		StartDate: "2099-12-31", // deve ser sobrescrito pela data de referência
		Allocations: []ProposedAllocation{
			{AccountID: "acc-1", Percentage: 70, Reason: "ok"},
			{AccountID: "acc-1", Percentage: 10, Reason: "duplicada"}, // dup
			{AccountID: "acc-2", Percentage: 200, Reason: "fora de faixa"},
			{AccountID: "ghost", Percentage: 50, Reason: "conta inexistente"},
		},
		Plans: []Plan{
			{Label: "Equilibrado", TargetAmount: "10000.00", DurationMonths: 12},
			{Label: "Inválido", TargetAmount: "0", DurationMonths: 6},  // target <= 0 → cai
			{Label: "Longo", TargetAmount: "10000.00", DurationMonths: 99}, // prazo limitado a 60
		},
	}
	out := sanitizeProposal(p, testAccounts(), refDate)
	if out == nil {
		t.Fatal("esperava proposta válida")
	}
	if len(out.Allocations) != 1 || out.Allocations[0].AccountID != "acc-1" {
		t.Fatalf("alocações não saneadas: %+v", out.Allocations)
	}
	if len(out.Plans) != 2 {
		t.Fatalf("esperava 2 planos válidos, got %d: %+v", len(out.Plans), out.Plans)
	}
	if out.Plans[1].DurationMonths != 60 {
		t.Fatalf("prazo deveria ser limitado a 60, got %d", out.Plans[1].DurationMonths)
	}
	if out.StartDate != "2025-01-01" {
		t.Fatalf("start_date deveria ser a data de referência, got %s", out.StartDate)
	}
}

func TestSanitizeProposal_NilWhenNoValidAllocation(t *testing.T) {
	p := &Proposal{
		Allocations: []ProposedAllocation{{AccountID: "ghost", Percentage: 50}},
		Plans:       []Plan{{Label: "x", TargetAmount: "1000.00", DurationMonths: 6}},
	}
	if sanitizeProposal(p, testAccounts(), refDate) != nil {
		t.Fatal("esperava nil sem alocação válida")
	}
}

func TestSanitizeProposal_NilWhenNoValidPlan(t *testing.T) {
	p := &Proposal{
		Allocations: []ProposedAllocation{{AccountID: "acc-1", Percentage: 50}},
		Plans:       []Plan{{Label: "x", TargetAmount: "0", DurationMonths: 6}},
	}
	if sanitizeProposal(p, testAccounts(), refDate) != nil {
		t.Fatal("esperava nil sem plano com valor-alvo > 0")
	}
}

func TestExtractJSON(t *testing.T) {
	// Recorta o objeto mesmo cercado por prosa.
	if got := extractJSON("Claro!\n{\"reply\":\"oi\"} fim"); got != `{"reply":"oi"}` {
		t.Fatalf("got %q", got)
	}
}

func TestParseDecimalLoose(t *testing.T) {
	cases := map[string]string{
		`10000`:      "10000.00",
		`"10000.50"`: "10000.50",
		`"abc"`:      "",
	}
	for in, want := range cases {
		if got := parseDecimalLoose(json.RawMessage(in)); got != want {
			t.Errorf("parseDecimalLoose(%s) = %q, want %q", in, got, want)
		}
	}
}

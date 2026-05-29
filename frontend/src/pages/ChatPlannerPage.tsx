import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { assistantPlan, createGoal, listAccounts } from '../api/client'
import type { Account, ChatMessage, PlanOption, PlanProposal } from '../api/types'
import { formatBRL, formatDateBR } from '../lib/format'
import { Button, ErrorBanner, PageShell, Spinner } from '../components/ui'

// Uma alocação escolhida pelo usuário (conta + percentual) pronta para a API.
type SelectedAllocation = { account_id: string; percentage: number }

// Saudação inicial fixa — evita uma ida ao LLM só pra abrir a conversa.
const GREETING =
  'Oi! Vou te ajudar a montar um objetivo de poupança. Para começar: o que você quer conquistar e quanto precisa juntar?'

export function ChatPlannerPage() {
  const { userID = '' } = useParams()
  const navigate = useNavigate()

  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'assistant', content: GREETING },
  ])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [proposal, setProposal] = useState<PlanProposal | null>(null)
  const [creating, setCreating] = useState(false)
  const [accounts, setAccounts] = useState<Account[]>([])

  // Carrega as contas do cliente para permitir editar a seleção na proposta.
  useEffect(() => {
    let active = true
    listAccounts(userID)
      .then((data) => active && setAccounts(data))
      .catch(() => {
        /* sem contas não impede a conversa; a proposta é que dependerá delas */
      })
    return () => {
      active = false
    }
  }, [userID])

  const endRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, proposal, sending])

  async function send() {
    const text = input.trim()
    if (!text || sending) return

    const history: ChatMessage[] = [...messages, { role: 'user', content: text }]
    setMessages(history)
    setInput('')
    setProposal(null)
    setError(null)
    setSending(true)
    try {
      const turn = await assistantPlan(userID, history)
      setMessages([...history, { role: 'assistant', content: turn.reply }])
      if (turn.done && turn.proposal) setProposal(turn.proposal)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
      // Devolve o texto pro input pra o usuário poder tentar de novo.
      setMessages(messages)
      setInput(text)
    } finally {
      setSending(false)
    }
  }

  // Cria o objetivo a partir da OPÇÃO de plano escolhida (target/prazo) + as
  // alocações que o usuário selecionou/ajustou no card.
  async function confirmPlan(plan: PlanOption, allocations: SelectedAllocation[]) {
    if (!proposal || creating) return
    setCreating(true)
    setError(null)
    try {
      const { goal_id } = await createGoal(userID, {
        name: proposal.name,
        target_amount: plan.target_amount,
        duration_months: plan.duration_months,
        start_date: proposal.start_date,
        allocations,
      })
      navigate(`/goals/${goal_id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
      setCreating(false)
    }
  }

  return (
    <PageShell
      title="Planejar com assistente"
      subtitle="Converse para montar seu objetivo — o assistente sugere quanto reservar de cada conta"
      right={
        <Button variant="secondary" onClick={() => navigate(`/users/${userID}`)}>
          ← Voltar
        </Button>
      }
    >
      {error && <ErrorBanner message={error} />}

      <div className="flex flex-col gap-3 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="flex max-h-[55vh] flex-col gap-3 overflow-y-auto pr-1">
          {messages.map((m, i) => (
            <div
              key={i}
              className={m.role === 'user' ? 'flex justify-end' : 'flex justify-start'}
            >
              <div
                className={
                  'max-w-[80%] whitespace-pre-wrap rounded-2xl px-4 py-2 text-sm ' +
                  (m.role === 'user'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-slate-100 text-slate-800')
                }
              >
                {m.content}
              </div>
            </div>
          ))}

          {sending && (
            <div className="flex justify-start">
              <div className="rounded-2xl bg-slate-100 px-4 py-2">
                <Spinner label="Pensando…" />
              </div>
            </div>
          )}

          {proposal && (
            <ProposalCard
              proposal={proposal}
              accounts={accounts}
              creating={creating}
              onConfirm={confirmPlan}
            />
          )}

          <div ref={endRef} />
        </div>

        <form
          className="flex items-center gap-2 border-t border-slate-100 pt-3"
          onSubmit={(e) => {
            e.preventDefault()
            send()
          }}
        >
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={sending || creating}
            placeholder="Escreva sua resposta…"
            className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:bg-slate-100"
          />
          <Button type="submit" disabled={!input.trim() || sending || creating}>
            Enviar
          </Button>
        </form>
      </div>
    </PageShell>
  )
}

type AllocRow = { selected: boolean; percentage: string; reason?: string }

function ProposalCard({
  proposal,
  accounts,
  creating,
  onConfirm,
}: {
  proposal: PlanProposal
  accounts: Account[]
  creating: boolean
  onConfirm: (plan: PlanOption, allocations: SelectedAllocation[]) => void
}) {
  // Linhas editáveis de alocação, semeadas com a sugestão da IA: as contas que
  // ela escolheu vêm marcadas com o percentual dela; o usuário pode ajustar,
  // desmarcar ou marcar outras contas suas.
  const [rows, setRows] = useState<Record<string, AllocRow>>({})

  // Contas oferecidas para seleção: as com saldo > 0 mais as que a IA escolheu
  // (mesmo que zeradas). Evita listar dezenas de contas vazias.
  const selectable = useMemo(() => {
    const picked = new Set(proposal.allocations.map((a) => a.account_id))
    return accounts.filter((a) => Number(a.balance) > 0 || picked.has(a.account_id))
  }, [accounts, proposal])

  useEffect(() => {
    const byId = new Map(proposal.allocations.map((a) => [a.account_id, a]))
    const seeded: Record<string, AllocRow> = {}
    for (const a of selectable) {
      const sug = byId.get(a.account_id)
      seeded[a.account_id] = {
        selected: !!sug,
        percentage: sug ? String(sug.percentage) : '',
        reason: sug?.reason,
      }
    }
    setRows(seeded)
  }, [selectable, proposal])

  function toggle(id: string, selected: boolean) {
    setRows((prev) => ({
      ...prev,
      [id]: { ...prev[id], selected, percentage: selected ? prev[id].percentage : '' },
    }))
  }
  function setPct(id: string, value: string) {
    setRows((prev) => ({ ...prev, [id]: { ...prev[id], percentage: value } }))
  }

  // Alocações válidas selecionadas (percentual inteiro 1–100).
  const selected: SelectedAllocation[] = []
  let allValid = true
  for (const a of selectable) {
    const row = rows[a.account_id]
    if (!row?.selected) continue
    const pct = Number(row.percentage)
    if (!Number.isInteger(pct) || pct < 1 || pct > 100) {
      allValid = false
      continue
    }
    selected.push({ account_id: a.account_id, percentage: pct })
  }
  const canConfirm = selected.length > 0 && allValid && !creating

  return (
    <div className="rounded-lg border border-indigo-200 bg-indigo-50/60 p-4">
      <h3 className="text-sm font-semibold text-slate-900">
        Plano para <span className="text-indigo-700">{proposal.name}</span>
      </h3>

      {/* Contas-fonte editáveis: comuns a todas as opções de plano. */}
      <p className="mt-2 text-xs font-medium uppercase tracking-wide text-slate-500">
        De onde sai · marque as contas e a fatia (%) da evolução mensal de cada uma
      </p>
      <div className="mt-1 max-h-56 overflow-y-auto rounded-md border border-slate-200 bg-white">
        <table className="min-w-full divide-y divide-slate-200 text-sm">
          <thead className="sticky top-0 bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-3 py-2 w-10"></th>
              <th className="px-3 py-2">Conta</th>
              <th className="px-3 py-2 text-right">Saldo</th>
              <th className="px-3 py-2 text-right w-24">% evolução</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {selectable.map((a) => {
              const row = rows[a.account_id]
              if (!row) return null
              return (
                <tr key={a.account_id} className={row.selected ? 'bg-indigo-50/50' : ''}>
                  <td className="px-3 py-2">
                    <input
                      type="checkbox"
                      checked={row.selected}
                      onChange={(e) => toggle(a.account_id, e.target.checked)}
                      className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
                    />
                  </td>
                  <td className="px-3 py-2">
                    <div className="font-medium text-slate-800">{a.brand_name}</div>
                    <div className="font-mono text-xs text-slate-500">{a.number}</div>
                    {row.selected && row.reason && (
                      <div className="text-xs text-indigo-600">{row.reason}</div>
                    )}
                  </td>
                  <td className="px-3 py-2 text-right text-slate-600">{formatBRL(a.balance)}</td>
                  <td className="px-3 py-2 text-right">
                    <input
                      type="number"
                      min={1}
                      max={100}
                      disabled={!row.selected}
                      value={row.percentage}
                      onChange={(e) => setPct(a.account_id, e.target.value)}
                      className="w-16 rounded-md border border-slate-300 px-2 py-1 text-right text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:bg-slate-100 disabled:text-slate-400"
                    />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {!canConfirm && selected.length === 0 && (
        <p className="mt-1 text-xs text-amber-600">Selecione ao menos uma conta para simular.</p>
      )}

      {/* Opções de plano: o cliente escolhe um ritmo para criar e simular. */}
      <p className="mt-4 text-xs font-medium uppercase tracking-wide text-slate-500">
        Escolha um ritmo
      </p>
      <div className="mt-1 grid gap-2 sm:grid-cols-3">
        {proposal.plans.map((plan, i) => {
          const months = plan.duration_months || 1
          const monthlyPace = Number(plan.target_amount) / months
          return (
            <div
              key={i}
              className="flex flex-col justify-between rounded-md border border-slate-200 bg-white p-3"
            >
              <div>
                <div className="text-sm font-semibold text-slate-900">{plan.label}</div>
                <div className="mt-1 text-lg font-semibold text-indigo-700">
                  {formatBRL(monthlyPace)}
                  <span className="text-xs font-normal text-slate-500">/mês*</span>
                </div>
                <div className="mt-1 text-xs text-slate-600">
                  {formatBRL(plan.target_amount)} em {plan.duration_months}{' '}
                  {plan.duration_months === 1 ? 'mês' : 'meses'}
                </div>
                {plan.summary && (
                  <div className="mt-1 text-xs text-slate-500">{plan.summary}</div>
                )}
              </div>
              <Button
                className="mt-3 w-full"
                onClick={() => onConfirm(plan, selected)}
                disabled={!canConfirm}
              >
                {creating ? 'Criando…' : 'Escolher e simular'}
              </Button>
            </div>
          )
        })}
      </div>

      <p className="mt-2 text-xs text-slate-500">
        *Ritmo de referência para bater a meta no prazo (meta ÷ prazo). O valor efetivamente
        reservado a cada mês depende de quanto cada conta evoluir — confira na simulação. Início em{' '}
        {formatDateBR(proposal.start_date)}.
      </p>
    </div>
  )
}

import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  HttpError,
  addGoalAllocation,
  getBacktest,
  getGoal,
  listAccounts,
  removeGoalAllocation,
  runBacktest,
} from '../api/client'
import type { Account, BacktestMovementStatus, BacktestResult, Goal } from '../api/types'
import { formatBRL, formatMonthBR } from '../lib/format'
import { Button, EmptyState, ErrorBanner, Modal, PageShell, Spinner } from '../components/ui'

export function BacktestResultsPage() {
  const { goalID = '' } = useParams()
  const navigate = useNavigate()
  const [goal, setGoal] = useState<Goal | null>(null)
  const [result, setResult] = useState<BacktestResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [rerunning, setRerunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [manageOpen, setManageOpen] = useState(false)

  // Carrega o backtest: tenta GET; se 404, dispara o POST automaticamente.
  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [g, r] = await Promise.all([
        getGoal(goalID),
        getBacktest(goalID).catch((e) => {
          if (e instanceof HttpError && e.status === 404) return runBacktest(goalID)
          throw e
        }),
      ])
      setGoal(g)
      setResult(r)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }, [goalID])

  useEffect(() => {
    void load()
  }, [load])

  async function handleRerun() {
    setRerunning(true)
    setError(null)
    try {
      const r = await runBacktest(goalID)
      setResult(r)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setRerunning(false)
    }
  }

  // Recarrega goal + re-executa o backtest após mexer nas alocações, pra que
  // KPIs, gráfico e tabela reflitam as contas atuais.
  const refreshAfterChange = useCallback(async () => {
    const [g, r] = await Promise.all([getGoal(goalID), runBacktest(goalID)])
    setGoal(g)
    setResult(r)
  }, [goalID])

  // Mapa account_id → rótulo amigável, a partir das alocações do goal.
  const accountLabel = useMemo(() => {
    const map: Record<string, string> = {}
    for (const a of goal?.allocations ?? []) {
      map[a.account_id] = [a.brand_name, a.number].filter(Boolean).join(' · ') || a.account_id
    }
    return (id: string) => map[id] ?? id
  }, [goal])

  // Mapa account_id → percentual da evolução da conta que vai pra meta.
  const accountPct = useMemo(() => {
    const map: Record<string, number> = {}
    for (const a of goal?.allocations ?? []) {
      map[a.account_id] = a.percentage
    }
    return (id: string): number | undefined => map[id]
  }, [goal])

  return (
    <PageShell
      title={goal ? goal.name : 'Resultado do backtest'}
      subtitle={goal ? `Meta de ${formatBRL(goal.target_amount)} em ${goal.duration_months} meses` : undefined}
      right={
        <>
          {goal && (
            <Button variant="secondary" onClick={() => navigate(`/users/${goal.user_id}`)}>
              ← Voltar pra contas
            </Button>
          )}
          {goal && (
            <Button variant="secondary" onClick={() => setManageOpen(true)}>
              Gerenciar contas
            </Button>
          )}
          <Button onClick={handleRerun} disabled={rerunning || loading}>
            {rerunning ? 'Re-executando…' : 'Re-executar backtest'}
          </Button>
        </>
      }
    >
      {error && <ErrorBanner message={error} />}

      {loading ? (
        <Spinner label="Carregando backtest…" />
      ) : result ? (
        <div className="space-y-8">
          <KpiCards result={result} />
          <VaultChart result={result} />
          <ContributionChart result={result} accountLabel={accountLabel} />
          <MonthlyTable result={result} accountLabel={accountLabel} accountPct={accountPct} />
        </div>
      ) : null}

      {manageOpen && goal && (
        <ManageAccountsModal
          goal={goal}
          onClose={() => setManageOpen(false)}
          onChanged={refreshAfterChange}
        />
      )}
    </PageShell>
  )
}

// ManageAccountsModal permite adicionar uma nova conta-fonte ao objetivo ou
// remover uma existente. Cada mutação chama onChanged, que recarrega o goal e
// re-executa o backtest na página.
function ManageAccountsModal({
  goal,
  onClose,
  onChanged,
}: {
  goal: Goal
  onClose: () => void
  onChanged: () => Promise<void>
}) {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  // account_id em mutação no momento (desabilita seus botões).
  const [busyID, setBusyID] = useState<string | null>(null)
  // Percentual digitado por conta disponível, antes de adicionar.
  const [pct, setPct] = useState<Record<string, string>>({})

  useEffect(() => {
    let active = true
    setLoading(true)
    listAccounts(goal.user_id)
      .then((data) => active && setAccounts(data))
      .catch((e) => active && setError(e instanceof Error ? e.message : String(e)))
      .finally(() => active && setLoading(false))
    return () => {
      active = false
    }
  }, [goal.user_id])

  const allocatedIDs = useMemo(
    () => new Set(goal.allocations.map((a) => a.account_id)),
    [goal],
  )
  const available = useMemo(
    () => accounts.filter((a) => !allocatedIDs.has(a.account_id)),
    [accounts, allocatedIDs],
  )

  // Executa uma mutação (add/remove) e, no sucesso, propaga onChanged. Os erros
  // do backend (ex: última alocação) viram banner sem fechar o modal.
  async function run(accountID: string, fn: () => Promise<void>) {
    setBusyID(accountID)
    setError(null)
    try {
      await fn()
      await onChanged()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusyID(null)
    }
  }

  async function handleAdd(accountID: string) {
    const value = Number(pct[accountID])
    if (!Number.isInteger(value) || value < 1 || value > 100) {
      setError('Informe um percentual inteiro entre 1 e 100.')
      return
    }
    await run(accountID, () =>
      addGoalAllocation(goal.goal_id, { account_id: accountID, percentage: value }),
    )
    setPct((prev) => ({ ...prev, [accountID]: '' }))
  }

  function handleRemove(accountID: string) {
    return run(accountID, () => removeGoalAllocation(goal.goal_id, accountID))
  }

  return (
    <Modal title="Gerenciar contas do objetivo" onClose={onClose}>
      {error && <ErrorBanner message={error} />}

      <section className="mb-6">
        <h3 className="mb-2 text-sm font-semibold text-slate-700">Contas alocadas</h3>
        {goal.allocations.length === 0 ? (
          <EmptyState title="Nenhuma conta alocada" />
        ) : (
          <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200">
            {goal.allocations.map((a) => (
              <li
                key={a.account_id}
                className="flex items-center justify-between gap-3 px-4 py-3"
              >
                <div className="min-w-0">
                  <div className="truncate font-medium text-slate-800">
                    {[a.brand_name, a.number].filter(Boolean).join(' · ') || a.account_id}
                  </div>
                  <div className="text-xs text-slate-500">{a.percentage}% da evolução mensal</div>
                </div>
                <Button
                  variant="secondary"
                  disabled={busyID != null || goal.allocations.length <= 1}
                  onClick={() => handleRemove(a.account_id)}
                  title={
                    goal.allocations.length <= 1
                      ? 'O objetivo precisa de ao menos uma conta'
                      : undefined
                  }
                >
                  {busyID === a.account_id ? 'Removendo…' : 'Remover'}
                </Button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h3 className="mb-2 text-sm font-semibold text-slate-700">Adicionar conta</h3>
        {loading ? (
          <Spinner label="Carregando contas…" />
        ) : available.length === 0 ? (
          <EmptyState title="Todas as contas do cliente já estão alocadas" />
        ) : (
          <ul className="divide-y divide-slate-100 rounded-lg border border-slate-200">
            {available.map((a) => (
              <li
                key={a.account_id}
                className="flex items-center justify-between gap-3 px-4 py-3"
              >
                <div className="min-w-0">
                  <div className="truncate font-medium text-slate-800">{a.brand_name}</div>
                  <div className="font-mono text-xs text-slate-500">
                    {a.branch_code} / {a.number}-{a.check_digit}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    min={1}
                    max={100}
                    placeholder="%"
                    disabled={busyID != null}
                    value={pct[a.account_id] ?? ''}
                    onChange={(e) =>
                      setPct((prev) => ({ ...prev, [a.account_id]: e.target.value }))
                    }
                    className="w-20 rounded-md border border-slate-300 px-2 py-1 text-right text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:bg-slate-100"
                  />
                  <Button
                    disabled={busyID != null}
                    onClick={() => handleAdd(a.account_id)}
                  >
                    {busyID === a.account_id ? 'Adicionando…' : 'Adicionar'}
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </Modal>
  )
}

function KpiCards({ result }: { result: BacktestResult }) {
  const s = result.summary
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm lg:col-span-1">
        <div className="text-xs font-medium uppercase tracking-wide text-slate-500">
          Saldo do baú
        </div>
        <div className="mt-1 text-2xl font-bold text-slate-900">
          {formatBRL(s.vault_balance)}
        </div>
      </div>
      <Kpi label="Meta" value={formatBRL(s.target_amount)} />
      <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="text-xs font-medium uppercase tracking-wide text-slate-500">
          Meta atingida?
        </div>
        <span
          className={[
            'mt-2 inline-flex rounded-full px-3 py-1 text-sm font-semibold',
            s.goal_reached ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700',
          ].join(' ')}
        >
          {s.goal_reached ? 'Sim' : 'Não'}
        </span>
      </div>
    </div>
  )
}

function Kpi({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="text-xs font-medium uppercase tracking-wide text-slate-500">{label}</div>
      <div className="mt-1 text-lg font-semibold text-slate-900">{value}</div>
    </div>
  )
}

function ExpandIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7" />
    </svg>
  )
}

// Painel com botão de foco: renderiza o conteúdo inline e, ao focar, reabre o
// mesmo conteúdo dentro de um modal amplo. O render-prop recebe `focused` pra
// o conteúdo poder crescer (ex.: altura do gráfico) na visão expandida.
function FocusPanel({
  title,
  maxWidthClass = 'max-w-6xl',
  children,
}: {
  title: string
  maxWidthClass?: string
  children: (focused: boolean) => ReactNode
}) {
  const [focused, setFocused] = useState(false)

  return (
    <>
      <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-700">{title}</h2>
          <button
            type="button"
            onClick={() => setFocused(true)}
            aria-label={`Focar em ${title}`}
            title="Focar"
            className="rounded p-1 text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
          >
            <ExpandIcon />
          </button>
        </div>
        {children(false)}
      </section>

      {focused && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4"
          onClick={() => setFocused(false)}
          role="dialog"
          aria-modal="true"
        >
          <div
            className={`flex max-h-[95vh] w-full ${maxWidthClass} flex-col overflow-hidden rounded-lg bg-white shadow-xl`}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="sticky top-0 flex items-center justify-between border-b border-slate-200 bg-white px-5 py-4">
              <h2 className="text-lg font-semibold text-slate-900">{title}</h2>
              <button
                type="button"
                onClick={() => setFocused(false)}
                aria-label="Fechar"
                className="rounded p-1 text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
              >
                ✕
              </button>
            </div>
            <div className="overflow-auto px-5 py-4">{children(true)}</div>
          </div>
        </div>
      )}
    </>
  )
}

function VaultChart({ result }: { result: BacktestResult }) {
  const target = Number(result.summary.target_amount)
  const data = result.vault_evolution.map((v) => ({
    month: formatMonthBR(v.month),
    balance: Number(v.balance),
  }))
  return (
    <FocusPanel title="Evolução do baú">
      {(focused) => (
        <div className={focused ? 'h-[75vh] w-full' : 'h-72 w-full'}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
            <XAxis dataKey="month" tick={{ fontSize: 12 }} />
            <YAxis
              tick={{ fontSize: 12 }}
              tickFormatter={(v) => formatBRL(v)}
              width={90}
            />
            <Tooltip
              formatter={(v) => formatBRL(v as number)}
              labelClassName="text-xs"
            />
            <ReferenceLine
              y={target}
              stroke="#ef4444"
              strokeDasharray="4 4"
              label={{ value: 'Meta', position: 'insideTopRight', fontSize: 12, fill: '#ef4444' }}
            />
            <Line
              type="monotone"
              dataKey="balance"
              stroke="#4f46e5"
              strokeWidth={2}
              dot={{ r: 3 }}
              name="Saldo do baú"
            />
          </LineChart>
        </ResponsiveContainer>
        </div>
      )}
    </FocusPanel>
  )
}

// Paleta para fatias do gráfico de contribuição por conta.
const PIE_COLORS = [
  '#4f46e5',
  '#0ea5e9',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
]

// Gráfico de pizza: quanto cada conta contribuiu pro baú no período, somando
// os valores poupados (movements.amount) por conta.
function ContributionChart({
  result,
  accountLabel,
}: {
  result: BacktestResult
  accountLabel: (id: string) => string
}) {
  const data = useMemo(() => {
    const totals = new Map<string, number>()
    for (const m of result.movements) {
      const amount = Number(m.amount)
      if (!amount) continue
      totals.set(m.account_id, (totals.get(m.account_id) ?? 0) + amount)
    }
    return Array.from(totals.entries())
      .map(([id, value]) => ({ id, name: accountLabel(id), value }))
      .sort((a, b) => b.value - a.value)
  }, [result, accountLabel])

  const total = data.reduce((sum, d) => sum + d.value, 0)

  if (data.length === 0) {
    return null
  }

  return (
    <FocusPanel title="Contribuição por conta">
      {(focused) => (
        <div className={focused ? 'h-[75vh] w-full' : 'h-72 w-full'}>
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
            <Pie
              data={data}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              outerRadius="80%"
              label={(entry) =>
                total > 0 ? `${Math.round((entry.value / total) * 100)}%` : ''
              }
            >
              {data.map((entry, i) => (
                <Cell key={entry.id} fill={PIE_COLORS[i % PIE_COLORS.length]} />
              ))}
            </Pie>
            <Tooltip formatter={(v) => formatBRL(v as number)} labelClassName="text-xs" />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
        </div>
      )}
    </FocusPanel>
  )
}

// Estilo de cada status na tabela mês a mês (RF-06).
const STATUS_STYLE: Record<
  BacktestMovementStatus,
  { icon: string; color: string; label: string }
> = {
  COMPLETED: { icon: '✅', color: 'text-green-600', label: 'Reserva cheia' },
  SKIPPED_NO_GROWTH: { icon: '➖', color: 'text-slate-400', label: 'Sem evolução' },
}

function MonthlyTable({
  result,
  accountLabel,
  accountPct,
}: {
  result: BacktestResult
  accountLabel: (id: string) => string
  accountPct: (id: string) => number | undefined
}) {
  // Pivot: linhas = meses, colunas = contas. Célula = status + valor.
  const months = useMemo(
    () => Array.from(new Set(result.movements.map((m) => m.reference_month))),
    [result],
  )
  const accountIDs = useMemo(
    () => Array.from(new Set(result.movements.map((m) => m.account_id))),
    [result],
  )
  const cell = useMemo(() => {
    const map = new Map<
      string,
      { status: string; amount: string; entradas: string; saidas: string }
    >()
    for (const m of result.movements) {
      map.set(`${m.reference_month}|${m.account_id}`, {
        status: m.status,
        amount: m.amount,
        entradas: m.entradas,
        saidas: m.saidas,
      })
    }
    return map
  }, [result])

  return (
    <FocusPanel title="Movimentos mês a mês" maxWidthClass="max-w-[90rem]">
      {() => (
        <>
      <div className="mb-4 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
        {(Object.keys(STATUS_STYLE) as BacktestMovementStatus[]).map((k) => (
          <span key={k} className="inline-flex items-center gap-1">
            <span className={STATUS_STYLE[k].color}>{STATUS_STYLE[k].icon}</span>
            {STATUS_STYLE[k].label}
          </span>
        ))}
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 text-sm">
          <thead className="text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-3 py-2">Mês</th>
              {accountIDs.map((id) => {
                const pct = accountPct(id)
                return (
                  <th key={id} className="px-3 py-2 whitespace-nowrap">
                    {accountLabel(id)}
                    {pct != null && <span className="text-slate-400"> ({pct}%)</span>}
                  </th>
                )
              })}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {months.map((month) => (
              <tr key={month}>
                <td className="px-3 py-2 whitespace-nowrap font-medium text-slate-700">
                  {formatMonthBR(month)}
                </td>
                {accountIDs.map((id) => {
                  const c = cell.get(`${month}|${id}`)
                  if (!c) return <td key={id} className="px-3 py-2 text-slate-300">—</td>
                  const s = STATUS_STYLE[c.status as BacktestMovementStatus] ?? STATUS_STYLE.SKIPPED_NO_GROWTH
                  // Tooltip no hover (RF-06): além do poupado, os totais de
                  // entradas e saídas da conta naquele mês.
                  const tooltip = [
                    s.label,
                    `Poupado: ${formatBRL(c.amount)}`,
                    `Entradas: ${formatBRL(c.entradas)}`,
                    `Saídas: ${formatBRL(c.saidas)}`,
                  ].join('\n')
                  return (
                    <td key={id} className="px-3 py-2 whitespace-nowrap">
                      <span className={`${s.color} cursor-help`} title={tooltip}>
                        {s.icon} {formatBRL(c.amount)}
                      </span>
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
        </>
      )}
    </FocusPanel>
  )
}

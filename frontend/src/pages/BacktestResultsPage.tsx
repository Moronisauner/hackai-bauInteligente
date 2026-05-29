import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  CartesianGrid,
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { HttpError, getBacktest, getGoal, runBacktest } from '../api/client'
import type { BacktestMovementStatus, BacktestResult, Goal } from '../api/types'
import { formatBRL, formatMonthBR } from '../lib/format'
import { Button, ErrorBanner, PageShell, Spinner } from '../components/ui'

export function BacktestResultsPage() {
  const { goalID = '' } = useParams()
  const navigate = useNavigate()
  const [goal, setGoal] = useState<Goal | null>(null)
  const [result, setResult] = useState<BacktestResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [rerunning, setRerunning] = useState(false)
  const [error, setError] = useState<string | null>(null)

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
          <KpiCards result={result} accountLabel={accountLabel} />
          <VaultChart result={result} />
          <MonthlyTable result={result} accountLabel={accountLabel} accountPct={accountPct} />
        </div>
      ) : null}
    </PageShell>
  )
}

function KpiCards({
  result,
  accountLabel,
}: {
  result: BacktestResult
  accountLabel: (id: string) => string
}) {
  const s = result.summary
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
      <Kpi label="Meses cumpridos" value={`${Math.round(s.completed_months_pct * 100)}%`} />
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
      <Kpi
        label="Conta com pior aproveitamento"
        value={s.worst_account_id ? accountLabel(s.worst_account_id) : '—'}
      />
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

function VaultChart({ result }: { result: BacktestResult }) {
  const target = Number(result.summary.target_amount)
  const data = result.vault_evolution.map((v) => ({
    month: formatMonthBR(v.month),
    balance: Number(v.balance),
  }))
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <h2 className="mb-4 text-sm font-semibold text-slate-700">Evolução do baú</h2>
      <div className="h-72 w-full">
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
    </section>
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
    <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <h2 className="mb-3 text-sm font-semibold text-slate-700">Movimentos mês a mês</h2>
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
    </section>
  )
}

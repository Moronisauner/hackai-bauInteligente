import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { createGoal, listAccounts } from '../api/client'
import type { Account, GoalDraft } from '../api/types'
import { formatBRL } from '../lib/format'
import {
  Button,
  EmptyState,
  ErrorBanner,
  PageShell,
  Spinner,
} from '../components/ui'

type Row = { selected: boolean; percentage: string }

export function AllocationFormPage() {
  const { userID = '' } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const draft = location.state as GoalDraft | null

  const [accounts, setAccounts] = useState<Account[]>([])
  const [rows, setRows] = useState<Record<string, Row>>({})
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // Sem dados da etapa 1 (ex: acesso direto à URL) → volta pro formulário.
  useEffect(() => {
    if (!draft) navigate(`/users/${userID}/goals/new`, { replace: true })
  }, [draft, navigate, userID])

  useEffect(() => {
    let active = true
    setLoading(true)
    setLoadError(null)
    listAccounts(userID)
      .then((data) => {
        if (!active) return
        setAccounts(data)
        setRows(
          Object.fromEntries(
            data.map((a) => [a.account_id, { selected: false, percentage: '' }]),
          ),
        )
      })
      .catch((e) => active && setLoadError(e instanceof Error ? e.message : String(e)))
      .finally(() => active && setLoading(false))
    return () => {
      active = false
    }
  }, [userID])

  const selectedRows = useMemo(
    () => Object.values(rows).filter((r) => r.selected),
    [rows],
  )
  const selectedCount = selectedRows.length
  // Cada conta marcada precisa de um percentual válido (1–100). As alocações são
  // independentes e NÃO precisam somar 100%.
  const allPctValid = selectedRows.every((r) => {
    const pct = Number(r.percentage)
    return Number.isInteger(pct) && pct >= 1 && pct <= 100
  })
  const canSubmit = selectedCount > 0 && allPctValid && !submitting

  function toggle(accountID: string, selected: boolean) {
    setRows((prev) => ({
      ...prev,
      [accountID]: {
        selected,
        percentage: selected ? prev[accountID].percentage : '',
      },
    }))
  }

  function setPct(accountID: string, value: string) {
    setRows((prev) => ({ ...prev, [accountID]: { ...prev[accountID], percentage: value } }))
  }

  async function handleSubmit() {
    if (!draft || !canSubmit) return
    setSubmitting(true)
    setSubmitError(null)
    const allocations = accounts
      .filter((a) => rows[a.account_id]?.selected)
      .map((a) => ({
        account_id: a.account_id,
        percentage: Number(rows[a.account_id].percentage),
      }))
    try {
      const { goal_id } = await createGoal(userID, { ...draft, allocations })
      navigate(`/goals/${goal_id}`)
    } catch (e) {
      setSubmitError(e instanceof Error ? e.message : String(e))
      setSubmitting(false)
    }
  }

  if (!draft) return null

  return (
    <PageShell
      title="Novo objetivo"
      subtitle={
        <>
          Etapa 2 de 2 · meta {formatBRL(draft.target_amount)} em {draft.duration_months} meses ·
          o % é a fatia do lucro mensal de cada conta (quanto do que ela lucrou no mês vai pra
          meta) — cada conta contribui de forma independente e se não houver lucro não vai
          contribuir com a meta.
        </>
      }
    >
      {submitError && <ErrorBanner message={submitError} />}
      {loadError && <ErrorBanner message={loadError} />}

      {loading ? (
        <Spinner label="Carregando contas…" />
      ) : accounts.length === 0 ? (
        <EmptyState title="Sem contas disponíveis para alocar" />
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-slate-200 text-sm">
              <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-4 py-3 w-12"></th>
                  <th className="px-4 py-3">Conta</th>
                  <th className="px-4 py-3 text-right">Saldo</th>
                  <th className="px-4 py-3 w-40 text-right">% do Saldo positivo</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {accounts.map((a) => {
                  const row = rows[a.account_id]
                  return (
                    <tr key={a.account_id} className={row.selected ? 'bg-indigo-50/40' : ''}>
                      <td className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={row.selected}
                          onChange={(e) => toggle(a.account_id, e.target.checked)}
                          className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-slate-800">{a.brand_name}</div>
                        <div className="font-mono text-xs text-slate-500">
                          {a.branch_code} / {a.number}-{a.check_digit}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-right text-slate-600">
                        {formatBRL(a.balance)}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <input
                          type="number"
                          min={1}
                          max={100}
                          disabled={!row.selected}
                          value={row.percentage}
                          onChange={(e) => setPct(a.account_id, e.target.value)}
                          className="w-20 rounded-md border border-slate-300 px-2 py-1 text-right text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:bg-slate-100 disabled:text-slate-400"
                        />
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          <div className="sticky bottom-0 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
            <span className="text-sm font-medium text-slate-600">
              {selectedCount === 0
                ? 'Nenhuma conta selecionada'
                : `${selectedCount} ${selectedCount === 1 ? 'conta selecionada' : 'contas selecionadas'}`}
            </span>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                onClick={() =>
                  navigate(`/users/${userID}/goals/new`, { state: draft })
                }
              >
                ← Voltar
              </Button>
              <Button disabled={!canSubmit} onClick={handleSubmit}>
                {submitting ? 'Criando…' : 'Criar objetivo'}
              </Button>
            </div>
          </div>
        </>
      )}
    </PageShell>
  )
}

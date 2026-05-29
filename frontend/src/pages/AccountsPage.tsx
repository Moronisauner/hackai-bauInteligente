import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { listAccounts, listGoals, listUsers } from '../api/client'
import type { Account, GoalSummary, User } from '../api/types'
import { formatBRL, formatDateBR } from '../lib/format'
import {
  Button,
  EmptyState,
  ErrorBanner,
  PageShell,
  Spinner,
} from '../components/ui'

const ACCOUNT_TYPE_LABEL: Record<string, string> = {
  CONTA_DEPOSITO_A_VISTA: 'Conta corrente',
  CONTA_POUPANCA: 'Poupança',
  CONTA_PAGAMENTO_PRE_PAGA: 'Conta de pagamento',
}

function accountTypeLabel(type: string): string {
  return ACCOUNT_TYPE_LABEL[type] ?? type
}

export function AccountsPage() {
  const { userID = '' } = useParams()
  const navigate = useNavigate()
  const [accounts, setAccounts] = useState<Account[]>([])
  const [goals, setGoals] = useState<GoalSummary[]>([])
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)
    Promise.all([listAccounts(userID), listGoals(userID), listUsers()])
      .then(([accountsData, goalsData, usersData]) => {
        if (!active) return
        setAccounts(accountsData)
        setGoals(goalsData)
        setUser(usersData.find((u) => u.user_id === userID) ?? null)
      })
      .catch((e) => active && setError(e instanceof Error ? e.message : String(e)))
      .finally(() => active && setLoading(false))
    return () => {
      active = false
    }
  }, [userID])

  const total = accounts.reduce((sum, a) => sum + Number(a.balance), 0)
  const refDate = accounts[0]?.balance_reference_date

  return (
    <PageShell
      title={user?.nome ?? 'Contas do cliente'}
      subtitle={
        <>
          <span className="font-mono">{userID}</span>
          {refDate && <span className="ml-2">· Saldo em {formatDateBR(refDate)}</span>}
        </>
      }
      right={
        <>
          <Button variant="secondary" onClick={() => navigate('/')}>
            ← Voltar
          </Button>
          {accounts.length > 0 && (
            <Button onClick={() => navigate(`/users/${userID}/goals/new`)}>
              Criar objetivo
            </Button>
          )}
        </>
      }
    >
      {error && <ErrorBanner message={error} />}

      {loading ? (
        <Spinner label="Carregando contas…" />
      ) : accounts.length === 0 ? (
        <EmptyState title="Este cliente não tem contas disponíveis">
          <button
            className="text-indigo-600 underline hover:text-indigo-700"
            onClick={() => navigate('/')}
          >
            Voltar para a seleção de clientes
          </button>
        </EmptyState>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
          <table className="min-w-full divide-y divide-slate-200 text-sm">
            <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-3">Instituição</th>
                <th className="px-4 py-3">Tipo</th>
                <th className="px-4 py-3">Agência / Número</th>
                <th className="px-4 py-3 text-right">Saldo</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {accounts.map((a) => (
                <tr key={a.account_id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium text-slate-800">{a.brand_name}</td>
                  <td className="px-4 py-3 text-slate-600">{accountTypeLabel(a.type)}</td>
                  <td className="px-4 py-3 font-mono text-slate-600">
                    {a.branch_code} / {a.number}-{a.check_digit}
                  </td>
                  <td className="px-4 py-3 text-right font-medium text-slate-800">
                    {formatBRL(a.balance)}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot className="bg-slate-50 font-semibold text-slate-800">
              <tr>
                <td className="px-4 py-3" colSpan={3}>
                  Total
                </td>
                <td className="px-4 py-3 text-right">{formatBRL(total)}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      )}

      {!loading && accounts.length > 0 && (
        <section className="mt-8">
          <h2 className="mb-3 text-lg font-semibold tracking-tight text-slate-900">
            Objetivos cadastrados
          </h2>

          {goals.length === 0 ? (
            <EmptyState title="Este cliente ainda não tem objetivos cadastrados">
              Crie o primeiro objetivo usando o botão acima.
            </EmptyState>
          ) : (
            <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
              <table className="min-w-full divide-y divide-slate-200 text-sm">
                <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                  <tr>
                    <th className="px-4 py-3">Objetivo</th>
                    <th className="px-4 py-3">Início</th>
                    <th className="px-4 py-3 text-right">Prazo</th>
                    <th className="px-4 py-3 text-right">Meta</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {goals.map((g) => (
                    <tr
                      key={g.goal_id}
                      className="cursor-pointer hover:bg-slate-50"
                      onClick={() => navigate(`/goals/${g.goal_id}`)}
                    >
                      <td className="px-4 py-3 font-medium text-slate-800">{g.name}</td>
                      <td className="px-4 py-3 text-slate-600">
                        {formatDateBR(g.start_date)}
                      </td>
                      <td className="px-4 py-3 text-right text-slate-600">
                        {g.duration_months} {g.duration_months === 1 ? 'mês' : 'meses'}
                      </td>
                      <td className="px-4 py-3 text-right font-medium text-slate-800">
                        {formatBRL(g.target_amount)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </PageShell>
  )
}

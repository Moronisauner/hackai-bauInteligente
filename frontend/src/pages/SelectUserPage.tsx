import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listUsers } from '../api/client'
import type { User } from '../api/types'
import { useConfig } from '../lib/useConfig'
import { formatDateBR } from '../lib/format'
import {
  EmptyState,
  ErrorBanner,
  PageShell,
  ReferenceDateBadge,
  Spinner,
} from '../components/ui'

export function SelectUserPage() {
  const navigate = useNavigate()
  const config = useConfig()
  const [query, setQuery] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Debounce de 300ms: refaz a busca só depois que o usuário para de digitar.
  useEffect(() => {
    const handle = setTimeout(() => {
      let active = true
      setLoading(true)
      setError(null)
      listUsers(query.trim() || undefined)
        .then((data) => active && setUsers(data))
        .catch((e) => active && setError(e instanceof Error ? e.message : String(e)))
        .finally(() => active && setLoading(false))
      return () => {
        active = false
      }
    }, 300)
    return () => clearTimeout(handle)
  }, [query])

  return (
    <PageShell
      title="Selecione um cliente"
      subtitle="Clientes com contas disponíveis, ordenados por nome."
      right={
        config && <ReferenceDateBadge date={formatDateBR(config.poc_reference_date)} />
      }
    >
      <div className="mb-4">
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Buscar por nome…"
          className="w-full rounded-md border border-slate-300 px-4 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
      </div>

      {error && <ErrorBanner message={error} />}

      {loading ? (
        <Spinner label="Carregando clientes…" />
      ) : users.length === 0 ? (
        <EmptyState title="Nenhum cliente encontrado">
          {query ? 'Tente outro termo de busca.' : 'Não há clientes com contas disponíveis.'}
        </EmptyState>
      ) : (
        <ul className="divide-y divide-slate-200 overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          {users.map((u) => (
            <li key={u.user_id}>
              <button
                onClick={() => navigate(`/users/${u.user_id}`)}
                className="flex w-full items-center justify-between px-4 py-3 text-left transition hover:bg-slate-50"
              >
                <span className="text-sm font-medium text-slate-800">{u.nome}</span>
                <span className="rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-600">
                  {u.accounts_count} {u.accounts_count === 1 ? 'conta' : 'contas'}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </PageShell>
  )
}

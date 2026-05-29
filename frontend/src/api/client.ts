// Cliente HTTP tipado. Toda chamada ao backend passa por aqui — nada de fetch
// solto nos componentes. As rotas são relativas a /api (proxy do Vite → :8080).
import type {
  Account,
  BacktestResult,
  Config,
  CreateGoalInput,
  CreateGoalResponse,
  Goal,
  GoalSummary,
  User,
} from './types'

const BASE = '/api'

// HttpError carrega o status pra que as páginas possam tratar 404 etc.
export class HttpError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'HttpError'
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.text()
    // O backend devolve {"error": "..."} — tenta extrair a mensagem.
    let detail = body
    try {
      const parsed = JSON.parse(body)
      if (parsed && typeof parsed.error === 'string') detail = parsed.error
    } catch {
      // corpo não-JSON; usa o texto cru
    }
    throw new HttpError(res.status, `${res.status} ${detail}`)
  }
  // 204 ou corpo vazio → undefined
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export function getConfig(): Promise<Config> {
  return request<Config>('/config')
}

export function listUsers(q?: string): Promise<User[]> {
  const query = q ? `?q=${encodeURIComponent(q)}` : ''
  return request<User[]>(`/users${query}`)
}

export function listAccounts(userID: string): Promise<Account[]> {
  return request<Account[]>(`/users/${encodeURIComponent(userID)}/accounts`)
}

export function listGoals(userID: string): Promise<GoalSummary[]> {
  return request<GoalSummary[]>(
    `/users/${encodeURIComponent(userID)}/goals`,
  )
}

export function createGoal(
  userID: string,
  body: CreateGoalInput,
): Promise<CreateGoalResponse> {
  return request<CreateGoalResponse>(
    `/users/${encodeURIComponent(userID)}/goals`,
    { method: 'POST', body: JSON.stringify(body) },
  )
}

export function getGoal(goalID: string): Promise<Goal> {
  return request<Goal>(`/goals/${encodeURIComponent(goalID)}`)
}

export function runBacktest(goalID: string): Promise<BacktestResult> {
  return request<BacktestResult>(
    `/goals/${encodeURIComponent(goalID)}/backtest`,
    { method: 'POST' },
  )
}

export function getBacktest(goalID: string): Promise<BacktestResult> {
  return request<BacktestResult>(`/goals/${encodeURIComponent(goalID)}/backtest`)
}

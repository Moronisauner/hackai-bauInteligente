// Tipos espelhando os DTOs do backend (backend/internal/httpapi/*).
// Valores monetários trafegam como string decimal — manter como string e
// formatar só na UI (vide src/lib/format.ts).

export type Config = {
  poc_reference_date: string // YYYY-MM-DD
}

export type User = {
  user_id: string
  nome: string
  accounts_count: number
}

export type Account = {
  account_id: string
  brand_name: string
  type: string
  branch_code: string
  number: string
  check_digit: string
  compe_code: string
  balance: string // string decimal
  balance_reference_date: string // YYYY-MM-DD
}

// Alocação como volta no detalhe do objetivo (GoalAllocationDTO), enriquecida
// com dados da conta-fonte. O percentual é a fatia da evolução mensal da conta.
export type Allocation = {
  account_id: string
  percentage: number
  brand_name?: string
  number?: string
  type?: string
}

// Resumo de um objetivo na listagem do cliente (GoalSummaryDTO).
export type GoalSummary = {
  goal_id: string
  name: string
  target_amount: string // string decimal
  duration_months: number
  start_date: string // YYYY-MM-DD
}

// Detalhe de um objetivo (GoalDTO). Observação: o backend usa `goal_id`.
export type Goal = {
  goal_id: string
  user_id: string
  name: string
  target_amount: string
  duration_months: number
  start_date: string
  vault_id: string
  allocations: Allocation[]
}

// Corpo de criação de objetivo (CreateGoalRequest). Sem dados da conta — só
// account_id + percentage por alocação.
export type CreateGoalInput = {
  name: string
  target_amount: string
  duration_months: number
  start_date: string
  allocations: { account_id: string; percentage: number }[]
}

export type CreateGoalResponse = {
  goal_id: string
  vault_id: string
}

// Estado parcial do objetivo coletado no step 1 do wizard (GoalFormPage),
// passado ao step 2 (AllocationFormPage) via location.state.
export type GoalDraft = {
  name: string
  target_amount: string // já normalizado (decimal "1234.56")
  duration_months: number
  start_date: string
}

// --- Assistente de planejamento (chat) ---

export type ChatRole = 'user' | 'assistant'

export type ChatMessage = {
  role: ChatRole
  content: string
}

// Alocação sugerida pelo assistente, enriquecida com dados da conta.
export type ProposedAllocation = {
  account_id: string
  percentage: number
  reason: string
  brand_name: string
  number: string
}

// Uma opção de plano (variação de valor/prazo) do mesmo objetivo.
export type PlanOption = {
  label: string
  summary: string
  target_amount: string
  duration_months: number
}

// Objetivo proposto pelo assistente: nome + alocações (comuns) + opções de plano.
// O cliente escolhe uma opção, que vira um CreateGoalInput.
export type PlanProposal = {
  name: string
  start_date: string
  allocations: ProposedAllocation[]
  plans: PlanOption[]
}

// Resposta de um turno do assistente (AssistantPlanResponse).
export type AssistantTurn = {
  reply: string
  done: boolean
  proposal: PlanProposal | null
}

export type BacktestMovementStatus =
  | 'COMPLETED'
  | 'SKIPPED_NO_GROWTH'

export type BacktestResult = {
  summary: {
    completed_months_pct: number
    completed_count: number
    failed_count: number
    vault_balance: string
    target_amount: string
    goal_reached: boolean
    worst_account_id: string // "" quando não há
  }
  movements: {
    reference_month: string
    account_id: string
    status: BacktestMovementStatus
    amount: string
    entradas: string // total de créditos efetivados da conta no mês
    saidas: string // total de débitos efetivados da conta no mês
  }[]
  vault_evolution: { month: string; balance: string }[]
}

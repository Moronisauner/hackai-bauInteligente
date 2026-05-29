import { createBrowserRouter } from 'react-router-dom'
import { SelectUserPage } from './pages/SelectUserPage'
import { AccountsPage } from './pages/AccountsPage'
import { GoalFormPage } from './pages/GoalFormPage'
import { AllocationFormPage } from './pages/AllocationFormPage'
import { BacktestResultsPage } from './pages/BacktestResultsPage'
import { ChatPlannerPage } from './pages/ChatPlannerPage'

export const router = createBrowserRouter([
  { path: '/', element: <SelectUserPage /> },
  { path: '/users/:userID', element: <AccountsPage /> },
  { path: '/users/:userID/goals/new', element: <GoalFormPage /> },
  { path: '/users/:userID/goals/new/allocations', element: <AllocationFormPage /> },
  { path: '/users/:userID/plan/chat', element: <ChatPlannerPage /> },
  { path: '/goals/:goalID', element: <BacktestResultsPage /> },
])

// Pequenos componentes de UI reaproveitados pelas páginas.
import type { ReactNode } from 'react'

export function PageShell({
  title,
  subtitle,
  right,
  children,
}: {
  title: string
  subtitle?: ReactNode
  right?: ReactNode
  children: ReactNode
}) {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <header className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
            {title}
          </h1>
          {subtitle && <div className="mt-1 text-sm text-slate-500">{subtitle}</div>}
        </div>
        {right && <div className="flex items-center gap-2">{right}</div>}
      </header>
      {children}
    </div>
  )
}

export function ReferenceDateBadge({ date }: { date: string }) {
  return (
    <span className="inline-flex items-center rounded-full bg-indigo-50 px-3 py-1 text-sm font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200">
      Data de referência: {date}
    </span>
  )
}

export function ErrorBanner({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
    >
      {message}
    </div>
  )
}

export function EmptyState({
  title,
  children,
}: {
  title: string
  children?: ReactNode
}) {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 bg-white px-6 py-12 text-center">
      <p className="text-sm font-medium text-slate-700">{title}</p>
      {children && <div className="mt-2 text-sm text-slate-500">{children}</div>}
    </div>
  )
}

export function Spinner({ label }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 py-12 text-slate-500" role="status">
      <span className="h-5 w-5 animate-spin rounded-full border-2 border-slate-300 border-t-indigo-600" />
      {label && <span className="text-sm">{label}</span>}
    </div>
  )
}

export function Button({
  children,
  variant = 'primary',
  className = '',
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary'
}) {
  const base =
    'inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50'
  const styles =
    variant === 'primary'
      ? 'bg-indigo-600 text-white hover:bg-indigo-700'
      : 'bg-white text-slate-700 ring-1 ring-inset ring-slate-300 hover:bg-slate-50'
  return (
    <button className={`${base} ${styles} ${className}`} {...props}>
      {children}
    </button>
  )
}

import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { GoalDraft } from '../api/types'
import { useConfig } from '../lib/useConfig'
import { parseBRLInput } from '../lib/format'
import { Button, PageShell } from '../components/ui'

type Field = 'name' | 'targetAmount' | 'durationMonths' | 'startDate'

export function GoalFormPage() {
  const { userID = '' } = useParams()
  const navigate = useNavigate()
  const config = useConfig()

  const [name, setName] = useState('')
  const [targetAmount, setTargetAmount] = useState('')
  const [durationMonths, setDurationMonths] = useState('12')
  const [startDate, setStartDate] = useState('')
  const [touched, setTouched] = useState<Record<Field, boolean>>({
    name: false,
    targetAmount: false,
    durationMonths: false,
    startDate: false,
  })

  // Default de start_date = POC_REFERENCE_DATE assim que a config carrega.
  const effectiveStartDate = startDate || config?.poc_reference_date || ''

  const errors = useMemo(() => {
    const e: Partial<Record<Field, string>> = {}
    if (!name.trim()) e.name = 'Informe um nome.'
    else if (name.trim().length > 255) e.name = 'Máximo de 255 caracteres.'

    const target = Number(parseBRLInput(targetAmount))
    if (!targetAmount.trim() || target <= 0) e.targetAmount = 'Valor deve ser maior que zero.'

    const dur = Number(durationMonths)
    if (!Number.isInteger(dur) || dur < 1 || dur > 60)
      e.durationMonths = 'Entre 1 e 60 meses.'

    if (!effectiveStartDate) e.startDate = 'Informe a data de início.'

    return e
  }, [name, targetAmount, durationMonths, effectiveStartDate])

  const isValid = Object.keys(errors).length === 0

  function markTouched(field: Field) {
    setTouched((t) => ({ ...t, [field]: true }))
  }

  function handleNext() {
    if (!isValid) {
      setTouched({
        name: true,
        targetAmount: true,
        durationMonths: true,
        startDate: true,
      })
      return
    }
    const draft: GoalDraft = {
      name: name.trim(),
      target_amount: parseBRLInput(targetAmount),
      duration_months: Number(durationMonths),
      start_date: effectiveStartDate,
    }
    navigate(`/users/${userID}/goals/new/allocations`, { state: draft })
  }

  return (
    <PageShell
      title="Novo objetivo"
      subtitle="Etapa 1 de 2 · Dados do objetivo"
    >
      <form
        className="space-y-5 rounded-lg border border-slate-200 bg-white p-6 shadow-sm"
        onSubmit={(e) => {
          e.preventDefault()
          handleNext()
        }}
      >
        <FormField label="Nome do objetivo" error={touched.name ? errors.name : undefined}>
          <input
            type="text"
            value={name}
            maxLength={255}
            onChange={(e) => setName(e.target.value)}
            onBlur={() => markTouched('name')}
            placeholder="Ex: Viagem de férias"
            className={inputClass(touched.name && !!errors.name)}
          />
        </FormField>

        <FormField
          label="Valor da meta (R$)"
          error={touched.targetAmount ? errors.targetAmount : undefined}
        >
          <input
            type="text"
            inputMode="decimal"
            value={targetAmount}
            onChange={(e) => setTargetAmount(e.target.value)}
            onBlur={() => markTouched('targetAmount')}
            placeholder="Ex: 10.000,00"
            className={inputClass(touched.targetAmount && !!errors.targetAmount)}
          />
        </FormField>

        <FormField
          label="Duração (meses)"
          error={touched.durationMonths ? errors.durationMonths : undefined}
        >
          <input
            type="number"
            min={1}
            max={60}
            value={durationMonths}
            onChange={(e) => setDurationMonths(e.target.value)}
            onBlur={() => markTouched('durationMonths')}
            className={inputClass(touched.durationMonths && !!errors.durationMonths)}
          />
        </FormField>

        <FormField
          label="Data de início"
          error={touched.startDate ? errors.startDate : undefined}
        >
          <input
            type="date"
            value={effectiveStartDate}
            onChange={(e) => setStartDate(e.target.value)}
            onBlur={() => markTouched('startDate')}
            className={inputClass(touched.startDate && !!errors.startDate)}
          />
        </FormField>

        <div className="flex justify-between border-t border-slate-100 pt-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => navigate(`/users/${userID}`)}
          >
            Voltar
          </Button>
          <Button type="submit" disabled={!isValid}>
            Próximo →
          </Button>
        </div>
      </form>
    </PageShell>
  )
}

function FormField({
  label,
  error,
  children,
}: {
  label: string
  error?: string
  children: React.ReactNode
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-slate-700">{label}</span>
      {children}
      {error && <span className="mt-1 block text-xs text-red-600">{error}</span>}
    </label>
  )
}

function inputClass(hasError: boolean): string {
  return [
    'w-full rounded-md border px-3 py-2 text-sm shadow-sm focus:outline-none focus:ring-1',
    hasError
      ? 'border-red-400 focus:border-red-500 focus:ring-red-500'
      : 'border-slate-300 focus:border-indigo-500 focus:ring-indigo-500',
  ].join(' ')
}

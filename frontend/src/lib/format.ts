// Helpers de formatação monetária e de data para a UI (pt-BR).

const BRL = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

// formatBRL: "1234.56" → "R$ 1.234,56". Aceita string decimal ou número.
export function formatBRL(value: string | number): string {
  const n = typeof value === 'number' ? value : Number(value)
  if (Number.isNaN(n)) return 'R$ 0,00'
  return BRL.format(n)
}

// parseBRLInput: input do usuário (ex: "1.234,56" ou "1234,56") → string
// decimal "1234.56" pronta pra API. Remove separador de milhar e troca a
// vírgula decimal por ponto.
export function parseBRLInput(value: string): string {
  const cleaned = value
    .replace(/\s/g, '')
    .replace(/R\$/gi, '')
    .replace(/\./g, '')
    .replace(',', '.')
    .replace(/[^0-9.]/g, '')
  if (cleaned === '' || cleaned === '.') return '0'
  const n = Number(cleaned)
  if (Number.isNaN(n)) return '0'
  return n.toFixed(2)
}

// formatDateBR: "2025-01-01" → "01/01/2025".
export function formatDateBR(isoDate: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(isoDate)
  if (!m) return isoDate
  return `${m[3]}/${m[2]}/${m[1]}`
}

// formatMonthBR: "2025-06-01" → "06/2025".
export function formatMonthBR(isoDate: string): string {
  const m = /^(\d{4})-(\d{2})/.exec(isoDate)
  if (!m) return isoDate
  return `${m[2]}/${m[1]}`
}

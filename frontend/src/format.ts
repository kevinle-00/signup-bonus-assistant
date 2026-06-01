export function dollarsToCents(value: string): number {
  const numeric = Number(value)
  if (!Number.isFinite(numeric) || numeric <= 0) {
    return 0
  }
  return Math.round(numeric * 100)
}

export function formatCents(cents: number): string {
  return new Intl.NumberFormat('en-AU', {
    style: 'currency',
    currency: 'AUD',
    maximumFractionDigits: 0,
  }).format(cents / 100)
}

export function formatDate(value?: string): string | null {
  if (!value) {
    return null
  }
  return new Intl.DateTimeFormat('en-AU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(new Date(value))
}

export function formatRewardType(value: string): string {
  return value.replaceAll('_', ' ')
}

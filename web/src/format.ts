export const money = (cents: number): string =>
  ((cents || 0) / 100).toLocaleString('en-US', { style: 'currency', currency: 'USD' })

export const dollarsToCents = (s: string): number => Math.round((parseFloat(s) || 0) * 100)

export const centsToDollars = (cents: number): string => (cents ? (cents / 100).toFixed(2) : '')

export const titleCase = (s: string): string => (s ? s.charAt(0).toUpperCase() + s.slice(1) : '')

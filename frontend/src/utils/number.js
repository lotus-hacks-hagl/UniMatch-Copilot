export function formatFloat2(value, fallback = '0.00') {
  const n = typeof value === 'string' ? Number(value) : value
  if (n === null || n === undefined) return fallback
  if (typeof n !== 'number' || !Number.isFinite(n)) return fallback
  return n.toFixed(2)
}

export function formatPercent2(value, fallback = '—') {
  const n = typeof value === 'string' ? Number(value) : value
  if (n === null || n === undefined) return fallback
  if (typeof n !== 'number' || !Number.isFinite(n)) return fallback
  return n.toFixed(2) + '%'
}


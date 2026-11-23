export function parseVersionParts(v: string): number[] {
  if (!v || v === '未查询') return []
  const clean = v.trim().toLowerCase().replace(/^v/, '')
  const match = clean.match(/^(\d+(?:\.\d+)*)/)
  if (!match) return []
  return match[1].split('.').map(num => parseInt(num, 10))
}

export function compareVersions(v1: string, v2: string): number {
  const p1 = parseVersionParts(v1)
  const p2 = parseVersionParts(v2)
  if (p1.length === 0 || p2.length === 0) {
    const clean1 = v1.trim().toLowerCase().replace(/^v/, '')
    const clean2 = v2.trim().toLowerCase().replace(/^v/, '')
    if (clean1 === clean2) return 0
    return clean1 > clean2 ? 1 : -1
  }
  const maxLen = Math.max(p1.length, p2.length)
  for (let i = 0; i < maxLen; i++) {
    const num1 = p1[i] || 0
    const num2 = p2[i] || 0
    if (num1 > num2) return 1
    if (num1 < num2) return -1
  }
  return 0
}

export function isVersionMatch(actual: string, target: string): boolean {
  if (!actual || actual === '未查询') return false
  const a = actual.trim().toLowerCase().replace(/^v/, '')
  const t = target.trim().toLowerCase().replace(/^v/, '')
  if (a === t) return true
  return compareVersions(actual, target) === 0
}

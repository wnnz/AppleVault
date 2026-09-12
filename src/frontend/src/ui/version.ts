export const bundledAppVersion = '1.3'

export function formatAppVersion(value: unknown): string {
  const version = typeof value === 'string' ? value.trim() : ''
  return (version || bundledAppVersion).replace(/\.0$/, '')
}

export const bundledAppVersion =
  (typeof __APP_VERSION__ === 'string' ? __APP_VERSION__ : '').trim().replace(/\.0$/, '') || '0.0'

export function formatAppVersion(value: unknown): string {
  const version = typeof value === 'string' ? value.trim() : ''
  return (version || bundledAppVersion).replace(/\.0$/, '')
}

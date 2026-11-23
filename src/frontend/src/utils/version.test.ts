import { describe, expect, it } from 'vitest'
import { compareVersions, isVersionMatch, parseVersionParts } from './version'

describe('parseVersionParts', () => {
  it('parses standard dotted versions and optional v prefix', () => {
    expect(parseVersionParts('8.0.50')).toEqual([8, 0, 50])
    expect(parseVersionParts('v2.1')).toEqual([2, 1])
    expect(parseVersionParts('  V10.2.3  ')).toEqual([10, 2, 3])
  })

  it('parses leading numeric segments when suffixes are present', () => {
    expect(parseVersionParts('1.2.3-beta.1')).toEqual([1, 2, 3])
  })

  it('returns empty array for empty, unqueried, or non-numeric strings', () => {
    expect(parseVersionParts('')).toEqual([])
    expect(parseVersionParts('未查询')).toEqual([])
    expect(parseVersionParts('abc')).toEqual([])
  })
})

describe('compareVersions', () => {
  it('treats trailing zeros as equal', () => {
    expect(compareVersions('8.0', '8.0.0')).toBe(0)
    expect(compareVersions('v1.2', '1.2.0.0')).toBe(0)
  })

  it('compares numeric segments properly rather than lexicographically', () => {
    expect(compareVersions('8.0.10', '8.0.2')).toBe(1)
    expect(compareVersions('8.0.2', '8.0.10')).toBe(-1)
    expect(compareVersions('v1.2', '1.3')).toBe(-1)
  })

  it('falls back to normalized string comparison when non-numeric', () => {
    expect(compareVersions('alpha', 'beta')).toBe(-1)
    expect(compareVersions('beta', 'alpha')).toBe(1)
    expect(compareVersions('vAlpha', 'alpha')).toBe(0)
  })
})

describe('isVersionMatch', () => {
  it('matches exact and equivalent semantic versions', () => {
    expect(isVersionMatch('8.0.50', '8.0.50')).toBe(true)
    expect(isVersionMatch('v8.0', '8.0.0')).toBe(true)
    expect(isVersionMatch('8.0.1', '8.0.2')).toBe(false)
  })

  it('returns false for unqueried or empty actual versions', () => {
    expect(isVersionMatch('未查询', '1.0.0')).toBe(false)
    expect(isVersionMatch('未查询', '未查询')).toBe(false)
    expect(isVersionMatch('', '')).toBe(false)
    expect(isVersionMatch('', '1.0')).toBe(false)
  })
})

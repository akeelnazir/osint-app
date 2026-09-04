import { describe, it, expect } from 'vitest'
import { extractMentions } from '@/lib/mentions'

describe('extractMentions', () => {
  it('extracts @username mentions', () => {
    const mentions = extractMentions('Hello @alice and @bob, check this out!')
    expect(mentions).toEqual(['alice', 'bob'])
  })

  it('returns empty for no mentions', () => {
    const mentions = extractMentions('No mentions here')
    expect(mentions).toEqual([])
  })

  it('handles mentions at start and end', () => {
    const mentions = extractMentions('@user1 says hi to @user2')
    expect(mentions).toEqual(['user1', 'user2'])
  })
})

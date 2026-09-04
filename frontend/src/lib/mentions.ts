// extractMentions finds @username mentions in a text body.
export function extractMentions(body: string): string[] {
  const mentions: string[] = []
  const re = /@([A-Za-z0-9_]+)/g
  let match
  while ((match = re.exec(body)) !== null) {
    mentions.push(match[1])
  }
  return mentions
}

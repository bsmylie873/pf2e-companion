import type { JSONContent } from '@tiptap/react'

export function extractPreviewText(content: JSONContent | null | undefined, maxLength = 120): string {
  if (!content || !content.content) return 'No content yet'
  const firstParagraph = content.content.find((node) => node.type === 'paragraph')
  if (!firstParagraph || !firstParagraph.content) return 'No content yet'
  const text = firstParagraph.content
    .filter((node) => node.type === 'text' && node.text)
    .map((node) => node.text)
    .join('')
  if (!text.trim()) return 'No content yet'
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength) + '\u2026'
}

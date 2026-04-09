import { extractPreviewText } from './contentPreview'

describe('extractPreviewText', () => {
  it('should return "No content yet" for null input', () => {
    expect(extractPreviewText(null)).toBe('No content yet')
  })

  it('should return "No content yet" for undefined input', () => {
    expect(extractPreviewText(undefined)).toBe('No content yet')
  })

  it('should return "No content yet" when content array is absent', () => {
    expect(extractPreviewText({})).toBe('No content yet')
  })

  it('should return "No content yet" when content array is empty', () => {
    expect(extractPreviewText({ type: 'doc', content: [] })).toBe('No content yet')
  })

  it('should return "No content yet" when no paragraph node exists', () => {
    const content = {
      type: 'doc',
      content: [
        { type: 'heading', content: [{ type: 'text', text: 'A Heading' }] },
      ],
    }
    expect(extractPreviewText(content)).toBe('No content yet')
  })

  it('should return "No content yet" for a paragraph with no content array', () => {
    const content = {
      type: 'doc',
      content: [{ type: 'paragraph' }],
    }
    expect(extractPreviewText(content)).toBe('No content yet')
  })

  it('should return "No content yet" for a paragraph with only whitespace text', () => {
    const content = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: '   ' }] }],
    }
    expect(extractPreviewText(content)).toBe('No content yet')
  })

  it('should return plain text from the first paragraph', () => {
    const content = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: 'Hello world' }] },
      ],
    }
    expect(extractPreviewText(content)).toBe('Hello world')
  })

  it('should concatenate multiple text nodes within a paragraph', () => {
    const content = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'Hello' },
            { type: 'text', text: ', ' },
            { type: 'text', text: 'world' },
          ],
        },
      ],
    }
    expect(extractPreviewText(content)).toBe('Hello, world')
  })

  it('should ignore non-text nodes within a paragraph', () => {
    const content = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'Caption' },
            { type: 'image', attrs: { src: 'photo.png' } },
          ],
        },
      ],
    }
    expect(extractPreviewText(content)).toBe('Caption')
  })

  it('should truncate text longer than the default 120 characters', () => {
    const longText = 'a'.repeat(130)
    const content = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: longText }] }],
    }
    const result = extractPreviewText(content)
    expect(result).toBe('a'.repeat(120) + '\u2026')
  })

  it('should not truncate text exactly at the default limit', () => {
    const exactText = 'b'.repeat(120)
    const content = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: exactText }] }],
    }
    expect(extractPreviewText(content)).toBe(exactText)
  })

  it('should respect a custom maxLength', () => {
    const content = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: 'Hello world' }] }],
    }
    expect(extractPreviewText(content, 5)).toBe('Hello\u2026')
  })

  it('should use only the first paragraph when multiple exist', () => {
    const content = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: 'First paragraph' }] },
        { type: 'paragraph', content: [{ type: 'text', text: 'Second paragraph' }] },
      ],
    }
    expect(extractPreviewText(content)).toBe('First paragraph')
  })
})

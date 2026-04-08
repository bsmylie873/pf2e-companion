import { Extension } from '@tiptap/core'

export const LINE_HEIGHT_PRESETS = [
  { label: 'Compact', value: '1.3' },
  { label: 'Normal', value: '1.5' },
  { label: 'Relaxed', value: '1.7' },
  { label: 'Spacious', value: '2.0' },
] as const

export type LineHeightValue = (typeof LINE_HEIGHT_PRESETS)[number]['value']

const VALID_VALUES = new Set<string>(LINE_HEIGHT_PRESETS.map((p) => p.value))

export const BlockLineHeight = Extension.create({
  name: 'lineHeight',

  addGlobalAttributes() {
    return [
      {
        types: ['paragraph', 'heading', 'blockquote', 'listItem', 'taskItem'],
        attributes: {
          lineHeight: {
            default: null,
            parseHTML: (element: HTMLElement) => {
              const val = element.style.lineHeight || null
              return val && VALID_VALUES.has(val) ? val : null
            },
            renderHTML: (attributes: { lineHeight?: string | null }) => {
              if (!attributes.lineHeight) return {}
              return { style: `line-height: ${attributes.lineHeight}` }
            },
          },
        },
      },
    ]
  },

  addCommands() {
    return {
      setLineHeight:
        (lineHeight: string) =>
        ({ commands }) =>
          ['paragraph', 'heading', 'blockquote', 'listItem', 'taskItem'].every((type) =>
            commands.updateAttributes(type, { lineHeight }),
          ),
      unsetLineHeight:
        () =>
        ({ commands }) =>
          ['paragraph', 'heading', 'blockquote', 'listItem', 'taskItem'].every((type) =>
            commands.updateAttributes(type, { lineHeight: null }),
          ),
    }
  },
})

import { describe, it, expect, vi } from 'vitest'

// Use vi.hoisted so the variable is available when vi.mock factory runs (vi.mock is hoisted)
const mockExtensionCreate = vi.hoisted(() =>
  vi.fn((config: Record<string, unknown>) => ({ ...config, _isMocked: true }))
)

vi.mock('@tiptap/core', () => ({
  Extension: {
    create: mockExtensionCreate,
  },
}))

import {
  LINE_HEIGHT_PRESETS,
  BlockLineHeight,
} from './blockLineHeight'
import type { LineHeightValue } from './blockLineHeight'

describe('LINE_HEIGHT_PRESETS', () => {
  it('is a readonly array', () => {
    expect(Array.isArray(LINE_HEIGHT_PRESETS)).toBe(true)
  })

  it('contains exactly 4 presets', () => {
    expect(LINE_HEIGHT_PRESETS).toHaveLength(4)
  })

  it('each preset has a label and value', () => {
    for (const preset of LINE_HEIGHT_PRESETS) {
      expect(preset).toHaveProperty('label')
      expect(preset).toHaveProperty('value')
      expect(typeof preset.label).toBe('string')
      expect(typeof preset.value).toBe('string')
    }
  })

  it('contains Compact preset with value 1.3', () => {
    const compact = LINE_HEIGHT_PRESETS.find((p) => p.label === 'Compact')
    expect(compact).toBeDefined()
    expect(compact?.value).toBe('1.3')
  })

  it('contains Normal preset with value 1.5', () => {
    const normal = LINE_HEIGHT_PRESETS.find((p) => p.label === 'Normal')
    expect(normal).toBeDefined()
    expect(normal?.value).toBe('1.5')
  })

  it('contains Relaxed preset with value 1.7', () => {
    const relaxed = LINE_HEIGHT_PRESETS.find((p) => p.label === 'Relaxed')
    expect(relaxed).toBeDefined()
    expect(relaxed?.value).toBe('1.7')
  })

  it('contains Spacious preset with value 2.0', () => {
    const spacious = LINE_HEIGHT_PRESETS.find((p) => p.label === 'Spacious')
    expect(spacious).toBeDefined()
    expect(spacious?.value).toBe('2.0')
  })

  it('all preset values are valid LineHeightValue strings', () => {
    const values = LINE_HEIGHT_PRESETS.map((p) => p.value)
    expect(values).toContain('1.3')
    expect(values).toContain('1.5')
    expect(values).toContain('1.7')
    expect(values).toContain('2.0')
  })
})

describe('BlockLineHeight extension', () => {
  it('is defined', () => {
    expect(BlockLineHeight).toBeDefined()
  })

  it('was created via Extension.create', () => {
    expect(mockExtensionCreate).toHaveBeenCalled()
  })

  it('has the name "lineHeight"', () => {
    const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
    expect(config.name).toBe('lineHeight')
  })

  it('has addGlobalAttributes method in config', () => {
    const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
    expect(typeof config.addGlobalAttributes).toBe('function')
  })

  it('has addCommands method in config', () => {
    const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
    expect(typeof config.addCommands).toBe('function')
  })

  describe('addGlobalAttributes()', () => {
    it('returns an array of attribute definitions', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const result = (config.addGlobalAttributes as () => unknown[])()
      expect(Array.isArray(result)).toBe(true)
      expect(result).toHaveLength(1)
    })

    it('targets the expected node types', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{types: string[]}>)()
      expect(attrDef.types).toContain('paragraph')
      expect(attrDef.types).toContain('heading')
      expect(attrDef.types).toContain('blockquote')
      expect(attrDef.types).toContain('listItem')
      expect(attrDef.types).toContain('taskItem')
    })

    it('lineHeight attribute defaults to null', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, unknown>) => unknown } }
      }>)()
      expect(attrDef.attributes.lineHeight.default).toBeNull()
    })

    it('parseHTML returns null for elements without line-height style', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, string | null>) => unknown } }
      }>)()
      const el = document.createElement('p')
      expect(attrDef.attributes.lineHeight.parseHTML(el)).toBeNull()
    })

    it('parseHTML returns valid line-height value from element style', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, string | null>) => unknown } }
      }>)()
      const el = document.createElement('p')
      el.style.lineHeight = '1.5'
      expect(attrDef.attributes.lineHeight.parseHTML(el)).toBe('1.5')
    })

    it('parseHTML returns null for invalid line-height values', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, string | null>) => unknown } }
      }>)()
      const el = document.createElement('p')
      el.style.lineHeight = '99'
      expect(attrDef.attributes.lineHeight.parseHTML(el)).toBeNull()
    })

    it('renderHTML returns empty object when lineHeight is null/undefined', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, string | null>) => unknown } }
      }>)()
      expect(attrDef.attributes.lineHeight.renderHTML({ lineHeight: null })).toEqual({})
    })

    it('renderHTML returns style attribute with line-height when lineHeight is set', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const [attrDef] = (config.addGlobalAttributes as () => Array<{
        types: string[]
        attributes: { lineHeight: { default: unknown; parseHTML: (el: HTMLElement) => unknown; renderHTML: (attrs: Record<string, string | null>) => unknown } }
      }>)()
      expect(attrDef.attributes.lineHeight.renderHTML({ lineHeight: '1.7' })).toEqual({
        style: 'line-height: 1.7',
      })
    })
  })

  describe('addCommands()', () => {
    it('returns setLineHeight and unsetLineHeight commands', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const commands = (config.addCommands as () => Record<string, unknown>)()
      expect(commands).toHaveProperty('setLineHeight')
      expect(commands).toHaveProperty('unsetLineHeight')
      expect(typeof commands.setLineHeight).toBe('function')
      expect(typeof commands.unsetLineHeight).toBe('function')
    })

    it('setLineHeight returns a command function that calls updateAttributes on all types', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const commands = (config.addCommands as () => Record<string, unknown>)()
      const setLineHeight = commands.setLineHeight as (lh: string) => (ctx: unknown) => boolean
      const commandFn = setLineHeight('1.5')
      const mockUpdateAttributes = vi.fn(() => true)
      const result = commandFn({ commands: { updateAttributes: mockUpdateAttributes } })
      expect(mockUpdateAttributes).toHaveBeenCalledTimes(5)
      expect(mockUpdateAttributes).toHaveBeenCalledWith('paragraph', { lineHeight: '1.5' })
      expect(mockUpdateAttributes).toHaveBeenCalledWith('heading', { lineHeight: '1.5' })
      expect(result).toBe(true)
    })

    it('unsetLineHeight returns a command function that clears lineHeight on all types', () => {
      const config = mockExtensionCreate.mock.calls[0][0] as Record<string, unknown>
      const commands = (config.addCommands as () => Record<string, unknown>)()
      const unsetLineHeight = commands.unsetLineHeight as () => (ctx: unknown) => boolean
      const commandFn = unsetLineHeight()
      const mockUpdateAttributes = vi.fn(() => true)
      commandFn({ commands: { updateAttributes: mockUpdateAttributes } })
      expect(mockUpdateAttributes).toHaveBeenCalledTimes(5)
      expect(mockUpdateAttributes).toHaveBeenCalledWith('paragraph', { lineHeight: null })
      expect(mockUpdateAttributes).toHaveBeenCalledWith('heading', { lineHeight: null })
    })
  })
})

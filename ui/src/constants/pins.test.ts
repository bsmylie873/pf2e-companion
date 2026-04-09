import { describe, it, expect } from 'vitest'
import {
  PIN_COLOURS,
  PIN_ICONS,
  COLOUR_MAP,
  PIN_ICON_COMPONENTS,
  PIN_ICON_LABELS,
} from './pins'
import type { PinColour, PinIcon } from './pins'

describe('PIN_COLOURS', () => {
  it('is a readonly array', () => {
    expect(Array.isArray(PIN_COLOURS)).toBe(true)
  })

  it('contains exactly 8 colours', () => {
    expect(PIN_COLOURS).toHaveLength(8)
  })

  it('contains expected colour names', () => {
    expect(PIN_COLOURS).toContain('grey')
    expect(PIN_COLOURS).toContain('red')
    expect(PIN_COLOURS).toContain('orange')
    expect(PIN_COLOURS).toContain('gold')
    expect(PIN_COLOURS).toContain('green')
    expect(PIN_COLOURS).toContain('blue')
    expect(PIN_COLOURS).toContain('purple')
    expect(PIN_COLOURS).toContain('brown')
  })
})

describe('PIN_ICONS', () => {
  it('is a readonly array', () => {
    expect(Array.isArray(PIN_ICONS)).toBe(true)
  })

  it('contains exactly 20 icon identifiers', () => {
    expect(PIN_ICONS).toHaveLength(20)
  })

  it('contains expected icon names', () => {
    expect(PIN_ICONS).toContain('position-marker')
    expect(PIN_ICONS).toContain('castle')
    expect(PIN_ICONS).toContain('crossed-swords')
    expect(PIN_ICONS).toContain('skull')
    expect(PIN_ICONS).toContain('treasure-map')
    expect(PIN_ICONS).toContain('campfire')
    expect(PIN_ICONS).toContain('forest-camp')
    expect(PIN_ICONS).toContain('mountain-cave')
    expect(PIN_ICONS).toContain('village')
    expect(PIN_ICONS).toContain('temple-gate')
    expect(PIN_ICONS).toContain('sailboat')
    expect(PIN_ICONS).toContain('crown')
    expect(PIN_ICONS).toContain('dragon-head')
    expect(PIN_ICONS).toContain('tombstone')
    expect(PIN_ICONS).toContain('bridge')
    expect(PIN_ICONS).toContain('mine-entrance')
    expect(PIN_ICONS).toContain('tower-flag')
    expect(PIN_ICONS).toContain('cauldron')
    expect(PIN_ICONS).toContain('wood-cabin')
    expect(PIN_ICONS).toContain('portal')
  })
})

describe('COLOUR_MAP', () => {
  it('is an object', () => {
    expect(typeof COLOUR_MAP).toBe('object')
    expect(COLOUR_MAP).not.toBeNull()
  })

  it('has an entry for every PIN_COLOUR', () => {
    for (const colour of PIN_COLOURS) {
      expect(COLOUR_MAP).toHaveProperty(colour)
    }
  })

  it('maps each colour to a hex string', () => {
    const hexPattern = /^#[0-9a-fA-F]{6}$/
    for (const colour of PIN_COLOURS) {
      expect(COLOUR_MAP[colour as PinColour]).toMatch(hexPattern)
    }
  })

  it('has correct hex values', () => {
    expect(COLOUR_MAP.grey).toBe('#8b8b8b')
    expect(COLOUR_MAP.red).toBe('#c94c4c')
    expect(COLOUR_MAP.orange).toBe('#d4783a')
    expect(COLOUR_MAP.gold).toBe('#c4a035')
    expect(COLOUR_MAP.green).toBe('#4a8c5c')
    expect(COLOUR_MAP.blue).toBe('#4a6fa5')
    expect(COLOUR_MAP.purple).toBe('#7b5ea7')
    expect(COLOUR_MAP.brown).toBe('#8b6b4a')
  })
})

describe('PIN_ICON_COMPONENTS', () => {
  it('is an object', () => {
    expect(typeof PIN_ICON_COMPONENTS).toBe('object')
    expect(PIN_ICON_COMPONENTS).not.toBeNull()
  })

  it('has an entry for every PIN_ICON', () => {
    for (const icon of PIN_ICONS) {
      expect(PIN_ICON_COMPONENTS).toHaveProperty(icon)
    }
  })

  it('maps each icon key to a React component (function)', () => {
    for (const icon of PIN_ICONS) {
      expect(typeof PIN_ICON_COMPONENTS[icon as PinIcon]).toBe('function')
    }
  })
})

describe('PIN_ICON_LABELS', () => {
  it('is an object', () => {
    expect(typeof PIN_ICON_LABELS).toBe('object')
    expect(PIN_ICON_LABELS).not.toBeNull()
  })

  it('has an entry for every PIN_ICON', () => {
    for (const icon of PIN_ICONS) {
      expect(PIN_ICON_LABELS).toHaveProperty(icon)
    }
  })

  it('maps each icon key to a non-empty string label', () => {
    for (const icon of PIN_ICONS) {
      expect(typeof PIN_ICON_LABELS[icon as PinIcon]).toBe('string')
      expect(PIN_ICON_LABELS[icon as PinIcon].length).toBeGreaterThan(0)
    }
  })

  it('has correct labels', () => {
    expect(PIN_ICON_LABELS['position-marker']).toBe('Position Marker')
    expect(PIN_ICON_LABELS['castle']).toBe('Castle')
    expect(PIN_ICON_LABELS['crossed-swords']).toBe('Crossed Swords')
    expect(PIN_ICON_LABELS['skull']).toBe('Skull')
    expect(PIN_ICON_LABELS['treasure-map']).toBe('Treasure Map')
    expect(PIN_ICON_LABELS['dragon-head']).toBe('Dragon Head')
    expect(PIN_ICON_LABELS['mine-entrance']).toBe('Mine Entrance')
    expect(PIN_ICON_LABELS['portal']).toBe('Portal')
  })

  it('PIN_ICONS and PIN_ICON_LABELS have the same keys', () => {
    const labelKeys = Object.keys(PIN_ICON_LABELS).sort()
    const iconKeys = [...PIN_ICONS].sort()
    expect(labelKeys).toEqual(iconKeys)
  })

  it('PIN_ICONS and PIN_ICON_COMPONENTS have the same keys', () => {
    const componentKeys = Object.keys(PIN_ICON_COMPONENTS).sort()
    const iconKeys = [...PIN_ICONS].sort()
    expect(componentKeys).toEqual(iconKeys)
  })
})

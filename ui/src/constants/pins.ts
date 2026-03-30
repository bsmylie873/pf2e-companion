import { GiPositionMarker, GiCastle, GiCrossedSwords, GiDeathSkull, GiTreasureMap, GiCampfire, GiForestCamp, GiMountainCave, GiVillage, GiTempleGate, GiSailboat, GiCrown, GiDragonHead, GiTombstone, GiBridge, GiGoldMine, GiTowerFlag, GiCauldron, GiWoodCabin, GiPortal } from 'react-icons/gi'
import type React from 'react'

export const PIN_COLOURS = ['grey', 'red', 'orange', 'gold', 'green', 'blue', 'purple', 'brown'] as const
export const PIN_ICONS = ['position-marker', 'castle', 'crossed-swords', 'skull', 'treasure-map', 'campfire', 'forest-camp', 'mountain-cave', 'village', 'temple-gate', 'sailboat', 'crown', 'dragon-head', 'tombstone', 'bridge', 'mine-entrance', 'tower-flag', 'cauldron', 'wood-cabin', 'portal'] as const
export type PinColour = typeof PIN_COLOURS[number]
export type PinIcon = typeof PIN_ICONS[number]

export const COLOUR_MAP: Record<PinColour, string> = {
  grey:   '#8b8b8b',
  red:    '#c94c4c',
  orange: '#d4783a',
  gold:   '#c4a035',
  green:  '#4a8c5c',
  blue:   '#4a6fa5',
  purple: '#7b5ea7',
  brown:  '#8b6b4a',
}

export const PIN_ICON_COMPONENTS: Record<string, React.ComponentType<{ size?: number }>> = {
  'position-marker': GiPositionMarker,
  'castle': GiCastle,
  'crossed-swords': GiCrossedSwords,
  'skull': GiDeathSkull,
  'treasure-map': GiTreasureMap,
  'campfire': GiCampfire,
  'forest-camp': GiForestCamp,
  'mountain-cave': GiMountainCave,
  'village': GiVillage,
  'temple-gate': GiTempleGate,
  'sailboat': GiSailboat,
  'crown': GiCrown,
  'dragon-head': GiDragonHead,
  'tombstone': GiTombstone,
  'bridge': GiBridge,
  'mine-entrance': GiGoldMine,
  'tower-flag': GiTowerFlag,
  'cauldron': GiCauldron,
  'wood-cabin': GiWoodCabin,
  'portal': GiPortal,
}

export const PIN_ICON_LABELS: Record<string, string> = {
  'position-marker': 'Position Marker',
  'castle': 'Castle',
  'crossed-swords': 'Crossed Swords',
  'skull': 'Skull',
  'treasure-map': 'Treasure Map',
  'campfire': 'Campfire',
  'forest-camp': 'Forest Camp',
  'mountain-cave': 'Mountain Cave',
  'village': 'Village',
  'temple-gate': 'Temple Gate',
  'sailboat': 'Sailboat',
  'crown': 'Crown',
  'dragon-head': 'Dragon Head',
  'tombstone': 'Tombstone',
  'bridge': 'Bridge',
  'mine-entrance': 'Mine Entrance',
  'tower-flag': 'Tower Flag',
  'cauldron': 'Cauldron',
  'wood-cabin': 'Wood Cabin',
  'portal': 'Portal',
}

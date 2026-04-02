import { createContext, useContext, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
import type { GameMap } from '../types/map'

interface MapNavState {
  gameId: string
  gameTitle: string
  maps: GameMap[]
  archivedMaps: GameMap[]
  activeMapId: string | null
  isGM: boolean
  onSelectMap: (mapId: string) => void
  onCreateMap: (name: string) => void
  onRenameMap: (mapId: string, name: string) => void
  onArchiveMap: (mapId: string) => void
  onUnarchiveMap: (mapId: string) => void
  onReorderMaps: (ids: string[]) => void
}

interface MapNavContextValue {
  state: MapNavState | null
  register: (state: MapNavState) => void
  unregister: () => void
}

const MapNavContext = createContext<MapNavContextValue>({
  state: null,
  register: () => {},
  unregister: () => {},
})

export function MapNavProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<MapNavState | null>(null)

  const register = useCallback((s: MapNavState) => setState(s), [])
  const unregister = useCallback(() => setState(null), [])

  return (
    <MapNavContext.Provider value={{ state, register, unregister }}>
      {children}
    </MapNavContext.Provider>
  )
}

export function useMapNav() {
  return useContext(MapNavContext)
}

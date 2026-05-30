import { create } from 'zustand'

import { GetAppConfig, UpdateAppConfig } from '../../wailsjs/go/main/App'
import { mockGames } from '@/data/mock-games'
import type { AppConfig, GameSettings, ImportedGame } from '@/types/game'

type GameState = {
  games: ImportedGame[]
  activeGameId?: string
  settings: GameSettings
  config?: AppConfig
  draftConfig?: AppConfig
  addGame: (game: ImportedGame) => void
  setActiveGame: (gameId: string) => void
  updateSettings: (settings: Partial<GameSettings>) => void
  fetchConfig: () => Promise<void>
  setDraftConfig: (config: AppConfig) => void
  resetDraftConfig: () => void
  saveDraftConfig: () => Promise<void>
}

const initialSettings: GameSettings = {
  textSpeed: 42,
  autoAdvance: false,
  showMap: true,
  voiceVolume: 64,
  uiScale: 100,
}

export const useGameStore = create<GameState>((set) => ({
  games: mockGames,
  settings: initialSettings,
  addGame: (game) =>
    set((state) => ({
      games: [game, ...state.games.filter((item) => item.id !== game.id)],
      activeGameId: game.id,
    })),
  setActiveGame: (gameId) => set({ activeGameId: gameId }),
  updateSettings: (settings) =>
    set((state) => ({
      settings: {
        ...state.settings,
        ...settings,
      },
    })),
  fetchConfig: async () => {
    const config = await GetAppConfig() as AppConfig
    set({
      config,
      draftConfig: { ...config },
    })
  },
  setDraftConfig: (config) => set({ draftConfig: config }),
  resetDraftConfig: () =>
    set((state) => ({
      draftConfig: state.config ? { ...state.config } : undefined,
    })),
  saveDraftConfig: async () => {
    const draftConfig = useGameStore.getState().draftConfig
    if (!draftConfig) {
      return
    }
    await UpdateAppConfig(draftConfig)
    set({
      config: { ...draftConfig },
      draftConfig: { ...draftConfig },
    })
  },
}))

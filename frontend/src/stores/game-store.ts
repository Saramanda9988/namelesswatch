import { create } from 'zustand'

import { mockGames } from '@/data/mock-games'
import type { GameSettings, ImportedGame } from '@/types/game'

type GameState = {
  games: ImportedGame[]
  activeGameId?: string
  settings: GameSettings
  addGame: (game: ImportedGame) => void
  setActiveGame: (gameId: string) => void
  updateSettings: (settings: Partial<GameSettings>) => void
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
}))

import { create } from 'zustand'

import { CreateGame, DeleteGame, GetAppConfig, GetGame, GetGames, ImportGamePack, UpdateAppConfig, UpdateGame } from '../../wailsjs/go/main/App'
import type { appconf, roleplay } from '../../wailsjs/go/models'

type GameSettings = {
  textSpeed: number
  autoAdvance: boolean
  showMap: boolean
  voiceVolume: number
  uiScale: number
}

type GameState = {
  games: roleplay.LibraryGame[]
  activeGameId?: string
  settings: GameSettings
  config?: appconf.AppConfig
  draftConfig?: appconf.AppConfig
  fetchGames: () => Promise<void>
  getGame: (gameId: string) => Promise<roleplay.LibraryGame>
  createGame: (game: roleplay.LibraryGame) => Promise<roleplay.LibraryGame>
  updateGame: (gameId: string, game: roleplay.LibraryGame) => Promise<roleplay.LibraryGame>
  deleteGame: (gameId: string) => Promise<void>
  importGameFiles: (files: Record<string, string>) => Promise<roleplay.ImportGameResult>
  setActiveGame: (gameId: string) => void
  updateSettings: (settings: Partial<GameSettings>) => void
  fetchConfig: () => Promise<void>
  setDraftConfig: (config: appconf.AppConfig) => void
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
  games: [],
  settings: initialSettings,
  fetchGames: async () => {
    const games = await GetGames()
    set({ games })
  },
  getGame: (gameId) => GetGame(gameId),
  createGame: async (game) => {
    const createdGame = await CreateGame(game)
    set((state) => ({
      games: [createdGame, ...state.games.filter((item) => item.id !== createdGame.id)],
      activeGameId: createdGame.id,
    }))
    return createdGame
  },
  updateGame: async (gameId, game) => {
    const updatedGame = await UpdateGame(gameId, game)
    set((state) => ({
      games: state.games.map((item) => (item.id === gameId ? updatedGame : item)),
    }))
    return updatedGame
  },
  deleteGame: async (gameId) => {
    await DeleteGame(gameId)
    set((state) => ({
      games: state.games.filter((game) => game.id !== gameId),
      activeGameId: state.activeGameId === gameId ? undefined : state.activeGameId,
    }))
  },
  importGameFiles: async (files) => {
    const result = await ImportGamePack(files)
    const games = await GetGames()
    set({ games })
    return result
  },
  setActiveGame: (gameId) => set({ activeGameId: gameId }),
  updateSettings: (settings) =>
    set((state) => ({
      settings: {
        ...state.settings,
        ...settings,
      },
    })),
  fetchConfig: async () => {
    const config = await GetAppConfig()
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

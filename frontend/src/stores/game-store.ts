import { create } from 'zustand'

import { CreateGame, CreateStoryTemplate, DeleteGame, DeleteSession, GetAppConfig, GetGame, GetGames, ImportGamePack, ListSessions, ListUnlockedAchievements, ResumeSession, SaveSnapshot, SelectStoryTemplateDirectory, SubmitChoice, SubmitCustomChoice, UpdateAppConfig, UpdateGame } from '../../wailsjs/go/main/App'
import type { appconf, main, roleplay, service } from '../../wailsjs/go/models'

type GameSettings = {
  textSpeed: number
  autoAdvance: boolean
  showMap: boolean
  voiceVolume: number
  bgmEnabled: boolean
  bgmVolume: number
  uiScale: number
}

type GameState = {
  games: roleplay.LibraryGame[]
  activeGameId?: string
  pendingResumeSessionId?: string
  settings: GameSettings
  config?: appconf.AppConfig
  draftConfig?: appconf.AppConfig
  fetchGames: () => Promise<void>
  getGame: (gameId: string) => Promise<roleplay.LibraryGame>
  createGame: (game: roleplay.LibraryGame) => Promise<roleplay.LibraryGame>
  updateGame: (gameId: string, game: roleplay.LibraryGame) => Promise<roleplay.LibraryGame>
  deleteGame: (gameId: string) => Promise<void>
  importGameFiles: (files: Record<string, string>) => Promise<roleplay.ImportGameResult>
  selectStoryTemplateDirectory: () => Promise<string>
  createStoryTemplate: (parentPath: string, folderName: string, title: string, initialScene: string, force: boolean) => Promise<main.StoryTemplateResult>
  setActiveGame: (gameId: string) => void
  setPendingResumeSession: (sessionId?: string) => void
  listSessions: (gameId: string) => Promise<service.SessionSummary[]>
  listUnlockedAchievements: (gameId: string) => Promise<roleplay.AchievementUnlock[]>
  resumeSession: (sessionId: string) => Promise<roleplay.GameTurnResult>
  submitChoice: (sessionId: string, choiceId: string) => Promise<roleplay.GameTurnResult>
  submitCustomChoice: (sessionId: string, reply: string) => Promise<roleplay.GameTurnResult>
  saveSnapshot: (sessionId: string, label: string) => Promise<service.SessionSummary>
  deleteSession: (sessionId: string) => Promise<void>
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
  bgmEnabled: true,
  bgmVolume: 64,
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
  selectStoryTemplateDirectory: () => SelectStoryTemplateDirectory(),
  createStoryTemplate: (parentPath, folderName, title, initialScene, force) => (
    CreateStoryTemplate(parentPath, folderName, title, initialScene, force)
  ),
  setActiveGame: (gameId) => set({ activeGameId: gameId }),
  setPendingResumeSession: (sessionId) => set({ pendingResumeSessionId: sessionId }),
  listSessions: (gameId) => ListSessions(gameId),
  listUnlockedAchievements: (gameId) => ListUnlockedAchievements(gameId),
  resumeSession: (sessionId) => ResumeSession(sessionId),
  submitChoice: (sessionId, choiceId) => SubmitChoice(sessionId, choiceId),
  submitCustomChoice: (sessionId, reply) => SubmitCustomChoice(sessionId, reply),
  saveSnapshot: (sessionId, label) => SaveSnapshot(sessionId, label),
  deleteSession: (sessionId) => DeleteSession(sessionId),
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

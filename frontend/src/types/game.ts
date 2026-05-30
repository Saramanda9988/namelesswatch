export type ScriptLine = {
  id: string
  speaker: string
  text: string
  backgroundUrl?: string
}

export type ImportedGame = {
  id: string
  title: string
  importedAt: string
  files: Record<string, string>
  photoUrls: string[]
  mapUrls: string[]
  script: ScriptLine[]
}

export type ImportReport = {
  game?: ImportedGame
  missing: string[]
  warnings: string[]
}

export type GameSettings = {
  textSpeed: number
  autoAdvance: boolean
  showMap: boolean
  voiceVolume: number
  uiScale: number
}

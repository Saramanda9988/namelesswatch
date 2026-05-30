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
  validFiles: string[]
}

export type SessionState = 'idle' | 'playing' | 'ended'

export type ChoiceOption = {
  id: string
  label: string
}

export type ChoiceTool = {
  type: 'choice'
  id: string
  prompt?: string
  options: ChoiceOption[]
}

export type EndingResult = {
  id: string
  title: string
  kind: 'good' | 'bad' | 'loop' | 'neutral'
}

export type GameTurn = {
  id: string
  role: 'ai' | 'user'
  payload: string[]
  selectedChoiceId?: string
  selectedChoiceLabel?: string
  tools?: ChoiceTool[]
  ending?: EndingResult
  createdAt: string
}

export type GameTurnResult = {
  sessionId: string
  gameId: string
  state: SessionState
  payload: string[]
  tools: ChoiceTool[]
  ending?: EndingResult
  turn: GameTurn
}

export type GameSession = {
  id: string
  gameId: string
  state: SessionState
  workspacePath: string
  memoryPath: string
  turns: GameTurn[]
  createdAt: string
  updatedAt: string
}

export type AgentTerminalRequest = {
  type: 'agent_terminal'
  reason: string
  commands: Array<{ command: string }>
}

export type GameSettings = {
  textSpeed: number
  autoAdvance: boolean
  showMap: boolean
  voiceVolume: number
  uiScale: number
}

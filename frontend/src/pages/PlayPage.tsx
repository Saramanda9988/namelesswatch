import { Link, useNavigate, useParams } from '@tanstack/react-router'
import {
  ArrowLeft,
  ChevronDown,
  FolderOpen,
  Loader2,
  Map,
  Moon,
  RefreshCcw,
  Save,
  Settings,
  SunMedium,
  Trash2,
  Volume2,
  VolumeX,
} from 'lucide-react'
import * as React from 'react'

import { RegisterGamePack, StartGame, SubmitChoice } from '../../wailsjs/go/main/App'
import { LogError, LogInfo } from '../../wailsjs/runtime/runtime'
import { PlayerBriefingPanel } from '@/components/player-briefing-panel'
import { PlaySidebar, type PlaySidebarHistoryItem, type PlaySidebarSceneMarker } from '@/components/play-sidebar'
import { PlaySettingsModal } from '@/components/play-settings-modal'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useBgmPlayer } from '@/hooks/use-bgm-player'
import { useNarrativeReveal } from '@/hooks/use-narrative-reveal'
import { parsePlayerBriefing } from '@/lib/player-briefing'
import { cn } from '@/lib/utils'
import { useGameStore } from '@/stores/game-store'
import type { roleplay, service } from '../../wailsjs/go/models'

function isRenderableTurn(turn: roleplay.GameTurn) {
  return turn.role === 'ai' && turn.payload.length > 0
}

function choiceToolFrom(result?: roleplay.GameTurnResult): roleplay.ChoiceTool | undefined {
  if (!result || result.state === 'ended') {
    return undefined
  }
  return result.tools?.find((tool) => tool.type === 'choice')
}

function logRuntimeInfo(message: string) {
  try {
    LogInfo(message)
  }
  catch {
    console.info(message)
  }
}

function logRuntimeError(message: string) {
  try {
    LogError(message)
  }
  catch {
    console.error(message)
  }
}

const startGameRequests = new globalThis.Map<string, Promise<roleplay.GameTurnResult>>()

function startGameOnce(game: roleplay.LibraryGame) {
  const existingRequest = startGameRequests.get(game.id)
  if (existingRequest) {
    logRuntimeInfo(`[play] start attached game=${game.id} title=${game.title}`)
    return existingRequest
  }

  logRuntimeInfo(`[play] start requested game=${game.id} title=${game.title}`)
  const request = RegisterGamePack(game.id, game.files).then(() => StartGame(game.id))

  // Only dedupe concurrent in-flight starts (e.g. React StrictMode's double effect on a
  // single mount). Once the request settles, drop it so a later visit starts a fresh
  // session instead of replaying the cached opening turn.
  const release = () => {
    if (startGameRequests.get(game.id) === request) {
      startGameRequests.delete(game.id)
    }
  }
  request.then(release, release)

  startGameRequests.set(game.id, request)
  return request
}

export function PlayPage() {
  const { gameId } = useParams({ from: '/play/$gameId' })
  const navigate = useNavigate()
  const games = useGameStore((state) => state.games)
  const setActiveGame = useGameStore((state) => state.setActiveGame)
  const textSpeed = useGameStore((state) => state.settings.textSpeed)
  const autoAdvance = useGameStore((state) => state.settings.autoAdvance)
  const bgmEnabled = useGameStore((state) => state.settings.bgmEnabled)
  const bgmVolume = useGameStore((state) => state.settings.bgmVolume)
  const updateSettings = useGameStore((state) => state.updateSettings)
  const resumeSession = useGameStore((state) => state.resumeSession)
  const saveSnapshot = useGameStore((state) => state.saveSnapshot)
  const listSessions = useGameStore((state) => state.listSessions)
  const deleteSession = useGameStore((state) => state.deleteSession)
  const game = games.find((item) => item.id === gameId)
  const playerBriefing = React.useMemo(() => parsePlayerBriefing(game?.files), [game?.files])
  const [sessionId, setSessionId] = React.useState<string>()
  const [latestResult, setLatestResult] = React.useState<roleplay.GameTurnResult>()
  const [turns, setTurns] = React.useState<roleplay.GameTurn[]>([])
  // Player choices are persisted server-side as user turns, but the live result only
  // carries the latest AI turn, so we track the labels here to interleave them into history.
  const [choiceLabels, setChoiceLabels] = React.useState<string[]>([])
  const [error, setError] = React.useState<string>()
  const [isStarting, setIsStarting] = React.useState(true)
  const [isBriefingConfirmed, setIsBriefingConfirmed] = React.useState(false)
  const [pendingChoiceId, setPendingChoiceId] = React.useState<string>()
  const [activeSceneId, setActiveSceneId] = React.useState<string>()
  const [snapshotBusy, setSnapshotBusy] = React.useState(false)
  const [snapshotHint, setSnapshotHint] = React.useState<string>()
  const [isLoadOpen, setIsLoadOpen] = React.useState(false)
  const [snapshots, setSnapshots] = React.useState<service.SessionSummary[]>()
  const [loadBusyId, setLoadBusyId] = React.useState<string>()
  const [isSidebarOpen, setIsSidebarOpen] = React.useState(false)
  const [isSettingsOpen, setIsSettingsOpen] = React.useState(false)
  const [restartToken, setRestartToken] = React.useState(0)

  // Capture the pending resume target exactly once. Reading/clearing it inside the
  // start effect breaks under React StrictMode (the effect runs twice; the first run
  // clears it so the second run falls back to starting a brand-new game).
  const resumeIdRef = React.useRef<string | undefined>(undefined)
  const resumeCapturedRef = React.useRef(false)
  if (!resumeCapturedRef.current) {
    resumeCapturedRef.current = true
    resumeIdRef.current = useGameStore.getState().pendingResumeSessionId
    useGameStore.getState().setPendingResumeSession(undefined)
  }

  React.useEffect(() => {
    if (!game) {
      return
    }
    setActiveGame(game.id)
  }, [game, setActiveGame])

  React.useEffect(() => {
    if (!game) {
      return
    }

    let cancelled = false
    const currentGame = game
    const resumeId = resumeIdRef.current
    setIsStarting(true)
    setError(undefined)
    setTurns([])
    setChoiceLabels([])
    setLatestResult(undefined)
    setSessionId(undefined)
    setActiveSceneId(currentGame.scenes?.[0]?.id)
    setIsBriefingConfirmed(Boolean(resumeId || !playerBriefing))

    async function start() {
      try {
        let result: roleplay.GameTurnResult
        if (resumeId) {
          logRuntimeInfo(`[play] resume requested game=${currentGame.id} session=${resumeId}`)
          await RegisterGamePack(currentGame.id, currentGame.files)
          result = await resumeSession(resumeId)
        }
        else {
          result = await startGameOnce(currentGame)
        }
        if (cancelled) {
          return
        }
        logRuntimeInfo(`[play] start result game=${currentGame.id} session=${result.sessionId} state=${result.state} tools=${result.tools?.length ?? 0} ending=${Boolean(result.ending)}`)
        setSessionId(result.sessionId)
        setLatestResult(result)
        setTurns([result.turn])
        if (result.scene?.id) {
          setActiveSceneId(result.scene.id)
        }
      }
      catch (cause) {
        if (!cancelled) {
          logRuntimeError(`[play] start failed game=${currentGame.id} error=${cause instanceof Error ? cause.message : String(cause)}`)
          setError(cause instanceof Error ? cause.message : String(cause))
        }
      }
      finally {
        if (!cancelled) {
          setIsStarting(false)
        }
      }
    }

    void start()

    return () => {
      cancelled = true
    }
  }, [game, playerBriefing, restartToken, resumeSession])

  async function submitChoice(choiceId: string) {
    if (!sessionId || pendingChoiceId || latestResult?.state === 'ended') {
      return
    }

    const chosenLabel
      = choiceToolFrom(latestResult)?.options.find((option) => option.id === choiceId)?.label
        ?? choiceId

    setPendingChoiceId(choiceId)
    setError(undefined)
    try {
      logRuntimeInfo(`[play] submit choice session=${sessionId} choice=${choiceId}`)
      const result = await SubmitChoice(sessionId, choiceId)
      logRuntimeInfo(`[play] choice result session=${sessionId} state=${result.state} tools=${result.tools?.length ?? 0} ending=${Boolean(result.ending)}`)
      setLatestResult(result)
      setChoiceLabels((current) => [...current, chosenLabel])
      setTurns((currentTurns) => [...currentTurns, result.turn])
      if (result.scene?.id) {
        setActiveSceneId(result.scene.id)
      }
    }
    catch (cause) {
      logRuntimeError(`[play] submit choice failed session=${sessionId} choice=${choiceId} error=${cause instanceof Error ? cause.message : String(cause)}`)
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setPendingChoiceId(undefined)
    }
  }

  async function handleSaveSnapshot() {
    if (!sessionId || snapshotBusy) {
      return
    }
    const label = window.prompt('存档名称（可留空）', `存档 ${new Date().toLocaleString()}`)
    if (label === null) {
      return
    }
    setSnapshotBusy(true)
    setSnapshotHint(undefined)
    try {
      const summary = await saveSnapshot(sessionId, label.trim())
      logRuntimeInfo(`[play] snapshot saved session=${sessionId} snapshot=${summary.id}`)
      setSnapshotHint('已保存快照')
      window.setTimeout(() => setSnapshotHint(undefined), 2400)
    }
    catch (cause) {
      logRuntimeError(`[play] snapshot failed session=${sessionId} error=${cause instanceof Error ? cause.message : String(cause)}`)
      setSnapshotHint('保存失败')
    }
    finally {
      setSnapshotBusy(false)
    }
  }

  async function refreshSnapshots() {
    if (!game) {
      return
    }
    try {
      const sessions = await listSessions(game.id)
      setSnapshots(sessions.filter((session) => session.isSnapshot))
    }
    catch (cause) {
      logRuntimeError(`[play] list sessions failed game=${game.id} error=${cause instanceof Error ? cause.message : String(cause)}`)
      setSnapshots([])
    }
  }

  function openLoadDialog() {
    setSnapshots(undefined)
    setIsLoadOpen(true)
    void refreshSnapshots()
  }

  async function handleLoadSnapshot(snapshotId: string) {
    if (loadBusyId) {
      return
    }
    setLoadBusyId(snapshotId)
    setError(undefined)
    try {
      const result = await resumeSession(snapshotId)
      logRuntimeInfo(`[play] loaded snapshot=${snapshotId} session=${result.sessionId}`)
      setSessionId(result.sessionId)
      setLatestResult(result)
      setChoiceLabels([])
      setTurns(result.turn ? [result.turn] : [])
      setIsLoadOpen(false)
    }
    catch (cause) {
      logRuntimeError(`[play] load snapshot failed snapshot=${snapshotId} error=${cause instanceof Error ? cause.message : String(cause)}`)
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setLoadBusyId(undefined)
    }
  }

  async function handleDeleteSnapshot(snapshotId: string) {
    if (loadBusyId) {
      return
    }
    setLoadBusyId(snapshotId)
    try {
      await deleteSession(snapshotId)
      await refreshSnapshots()
    }
    catch (cause) {
      logRuntimeError(`[play] delete snapshot failed snapshot=${snapshotId} error=${cause instanceof Error ? cause.message : String(cause)}`)
    }
    finally {
      setLoadBusyId(undefined)
    }
  }

  const renderableTurns = React.useMemo(() => turns.filter(isRenderableTurn), [turns])
  const choiceTool = choiceToolFrom(latestResult)
  const isEnded = latestResult?.state === 'ended'
  const currentTurn = renderableTurns.at(-1)
  const currentLines = currentTurn?.payload ?? []
  const activeScene = React.useMemo(() => {
    if (!game) {
      return undefined
    }
    return game.scenes?.find((scene) => scene.id === activeSceneId) ?? game.scenes?.[0]
  }, [activeSceneId, game])
  const sceneImage = activeScene?.url || game?.photoUrls?.[0]
  const gameOverImage = React.useMemo(() => {
    return game ? findGameOverImage(game) : undefined
  }, [game])
  const mapImage = game?.mapUrls?.[0]
  const currentBgmId = latestResult?.currentBgmId
  const currentBgm = React.useMemo(() => {
    if (!currentBgmId) {
      return undefined
    }
    return game?.bgms?.find((bgm) => bgm.id === currentBgmId && bgm.url)
  }, [currentBgmId, game?.bgms])
  const bgmPlayer = useBgmPlayer({
    track: currentBgm ? { id: currentBgm.id, url: currentBgm.url } : undefined,
    enabled: bgmEnabled,
    volume: bgmVolume,
  })
  const sceneMarkers = React.useMemo<PlaySidebarSceneMarker[]>(() => {
    return (game?.scenes ?? [])
      .filter((scene) => scene.hasPosition)
      .map((scene) => ({
        id: scene.id,
        name: scene.name || scene.id,
        x: scene.x,
        y: scene.y,
        active: scene.id === (activeScene?.id ?? activeSceneId),
      }))
  }, [game?.scenes, activeScene?.id, activeSceneId])

  React.useEffect(() => {
    if (!game) {
      return
    }
    logRuntimeInfo(`[play] scene resolved game=${game.id} scenes=${game.scenes?.length ?? 0} active=${activeScene?.id ?? activeSceneId ?? ''} image=${sceneImage ?? ''}`)
  }, [activeScene?.id, activeSceneId, game, sceneImage])

  React.useEffect(() => {
    if (!game || !currentBgmId || currentBgm) {
      return
    }
    logRuntimeError(`[play] bgm missing game=${game.id} bgm=${currentBgmId}`)
  }, [currentBgm, currentBgmId, game])

  const narrativeLines = isBriefingConfirmed ? currentLines : []
  const { revealedLines, activeIndex, phase, isComplete, advance } = useNarrativeReveal({
    lines: narrativeLines,
    resetKey: `${currentTurn?.id ?? 'none'}:${isBriefingConfirmed ? 'ready' : 'briefing'}`,
    textSpeed,
    autoAdvance,
  })

  const narrativeScrollRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    const node = narrativeScrollRef.current
    if (node) {
      node.scrollTop = node.scrollHeight
    }
  }, [revealedLines.length, activeIndex])

  // Chat history interleaves the player's own choices with the AI narrative turns.
  // The currently-typing AI turn stays exclusively in the left box and only joins the
  // history once it is fully revealed; the player's message appears as soon as it is sent.
  const historyItems = React.useMemo(() => {
    const items: PlaySidebarHistoryItem[] = []
    renderableTurns.forEach((turn, index) => {
      const isLatest = index === renderableTurns.length - 1
      if (!isLatest || isComplete) {
        items.push({ key: turn.id, role: 'ai', text: turn.payload.join('\n') })
      }
      const choice = choiceLabels[index]
      if (choice) {
        items.push({ key: `${turn.id}-choice-${index}`, role: 'user', text: choice })
      }
    })
    return items
  }, [renderableTurns, choiceLabels, isComplete])

  const showGameOver = Boolean(isBriefingConfirmed && isEnded && latestResult?.ending && isComplete)
  const showBriefingPanel = Boolean(playerBriefing && !isBriefingConfirmed && !error && !showGameOver)
  const canAdvance = !showBriefingPanel && !isStarting && !error && currentLines.length > 0 && !isComplete
  const showChoicePanel = Boolean(!showBriefingPanel && choiceTool && !isEnded && isComplete)
  const visibleNarrativeLine = showChoicePanel && choiceTool?.prompt
    ? choiceTool.prompt
    : revealedLines.at(-1) ?? ''
  const stageImage = showGameOver ? gameOverImage || sceneImage : sceneImage

  const isBgmAudible = bgmEnabled && bgmVolume > 0

  function handleReplay() {
    resumeIdRef.current = undefined
    setSessionId(undefined)
    setLatestResult(undefined)
    setTurns([])
    setChoiceLabels([])
    setError(undefined)
    setPendingChoiceId(undefined)
    setActiveSceneId(game?.scenes?.[0]?.id)
    setSnapshotHint(undefined)
    setIsBriefingConfirmed(!playerBriefing)
    setIsStarting(true)
    setRestartToken((current) => current + 1)
  }

  function handleBgmControl() {
    if (!isBgmAudible) {
      updateSettings({
        bgmEnabled: true,
        bgmVolume: bgmVolume > 0 ? bgmVolume : 64,
      })
      return
    }
    if (bgmPlayer.isBlocked) {
      void bgmPlayer.retry()
      return
    }
    updateSettings({ bgmEnabled: false })
  }

  const bgmButtonLabel = !isBgmAudible
    ? '启用背景音乐'
    : bgmPlayer.isBlocked
      ? '启用音乐播放'
      : '静音背景音乐'

  React.useEffect(() => {
    if (!canAdvance) {
      return
    }
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === ' ' || event.key === 'Enter') {
        event.preventDefault()
        advance()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [canAdvance, advance])

  if (!game) {
    return (
      <div className="dark grid min-h-screen place-items-center bg-background p-6 text-foreground">
        <Card className="w-full max-w-sm text-center">
          <CardHeader>
            <CardTitle>未找到游戏</CardTitle>
            <CardDescription>这个剧情包不存在或已经被移除。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link to="/">返回主页</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="dark h-screen overflow-hidden bg-background text-foreground">
      <main className={cn('grid h-full grid-cols-1', !showGameOver && 'lg:grid-cols-[minmax(0,1fr)_360px]')}>
        <section className="relative min-h-0 overflow-hidden bg-card">
          {stageImage ? (
            <img src={stageImage} alt="" className="absolute inset-0 size-full object-cover" />
          ) : (
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_42%,hsl(var(--muted))_0,transparent_34%),linear-gradient(90deg,hsl(var(--card))_0%,hsl(var(--muted))_52%,hsl(var(--background))_100%)]" />
          )}
          <div className="absolute inset-0 bg-gradient-to-r from-background/55 via-background/18 to-background/72" />
          <div className="absolute inset-0 bg-gradient-to-t from-background/92 via-transparent to-background/42" />

          <nav className="absolute left-4 top-4 z-20 flex flex-col gap-3">
            <Button asChild variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="返回">
              <Link to="/">
                <ArrowLeft data-icon />
              </Link>
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="bg-background/55 backdrop-blur-md"
              aria-label="设置"
              title="设置"
              onClick={() => setIsSettingsOpen(true)}
            >
              <Settings data-icon />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="bg-background/55 backdrop-blur-md"
              aria-label="保存快照"
              title="保存快照"
              disabled={!sessionId || snapshotBusy}
              onClick={() => void handleSaveSnapshot()}
            >
              {snapshotBusy ? <Loader2 className="animate-spin" data-icon /> : <Save data-icon />}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="bg-background/55 backdrop-blur-md"
              aria-label={bgmButtonLabel}
              aria-pressed={isBgmAudible}
              title={bgmButtonLabel}
              onClick={handleBgmControl}
            >
              {isBgmAudible && !bgmPlayer.isBlocked ? <Volume2 data-icon /> : <VolumeX data-icon />}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="bg-background/55 backdrop-blur-md"
              aria-label="读档"
              title="读档"
              onClick={openLoadDialog}
            >
              <FolderOpen data-icon />
            </Button>
            <Button type="button" variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="昼夜">
              <SunMedium data-icon />
            </Button>
            <Button type="button" variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="夜间">
              <Moon data-icon />
            </Button>
            {snapshotHint ? (
              <span className="rounded-md bg-background/70 px-2 py-1 text-center text-xs text-muted-foreground backdrop-blur-md">
                {snapshotHint}
              </span>
            ) : null}
          </nav>

          {!showGameOver ? (
            <div className="absolute right-4 top-4 z-20 flex items-center gap-2 lg:hidden">
              <Button
                type="button"
                variant="outline"
                size="icon-lg"
                className="bg-background/55 backdrop-blur-md"
                aria-label="打开侧栏"
                title="打开侧栏"
                onClick={() => setIsSidebarOpen(true)}
              >
                <Map data-icon />
              </Button>
            </div>
          ) : null}

          {showBriefingPanel && playerBriefing ? (
            <div className="absolute inset-x-0 top-[12vh] z-20 px-4 md:top-[12%]">
              <PlayerBriefingPanel
                briefing={playerBriefing}
                isPreparing={isStarting}
                onConfirm={() => setIsBriefingConfirmed(true)}
              />
            </div>
          ) : null}

          {showGameOver && latestResult?.ending ? (
            <GameOverScreen
              endingTitle={latestResult.ending.title}
              showTitle={!gameOverImage}
              onReplay={handleReplay}
            />
          ) : (
            <>
              {!showBriefingPanel ? (
                <div className="absolute inset-x-0 bottom-0 z-20 px-4 pb-5 md:px-8 md:pb-8">
                  <section
                    role="button"
                    tabIndex={canAdvance ? 0 : -1}
                    aria-label="点击继续"
                    onClick={canAdvance ? () => advance() : undefined}
                    className={`mx-auto flex max-w-4xl flex-col gap-3 rounded-lg border bg-background/78 p-4 shadow-2xl backdrop-blur-md md:p-5 ${canAdvance ? 'cursor-pointer' : ''}`}
                  >
                    <div className="flex min-h-[112px] flex-col gap-3">
                      {isStarting ? (
                        <div className="flex items-center gap-3 text-muted-foreground">
                          <Loader2 className="animate-spin" data-icon />
                          <span>正在准备游戏会话...</span>
                        </div>
                      ) : null}

                      {error ? (
                        <div className="flex flex-col gap-3">
                          <div className="flex items-center gap-2 text-destructive">
                            <RefreshCcw className="size-4" />
                            <span className="font-medium">会话暂不可用</span>
                          </div>
                          <p className="break-words text-sm leading-6 text-muted-foreground">{error}</p>
                          <Button type="button" variant="outline" className="w-fit" onClick={() => void navigate({ to: '/' })}>
                            返回游戏库
                          </Button>
                        </div>
                      ) : null}

                      {!isStarting && !error && currentLines.length > 0 ? (
                        <div
                          ref={narrativeScrollRef}
                          className="max-h-[25vh] overflow-y-auto pr-3"
                        >
                          <div className="flex flex-col gap-2 text-base leading-7 text-foreground md:text-lg">
                            <p key={`${currentTurn?.id ?? 'turn'}-${activeIndex}`} className="text-pretty">
                              {visibleNarrativeLine}
                              {phase === 'typing' && !showChoicePanel ? (
                                <span className="ml-0.5 inline-block h-[1.05em] w-[2px] animate-pulse bg-foreground align-[-0.1em]" />
                              ) : null}
                            </p>
                          </div>
                        </div>
                      ) : null}

                      {!isStarting && !error && currentLines.length === 0 ? (
                        <p className="text-base leading-7 text-muted-foreground">等待 AI 主持生成开场叙事...</p>
                      ) : null}
                    </div>

                    {!isStarting && !error && phase === 'waiting' && !showChoicePanel ? (
                      <ChevronDown className="absolute bottom-4 right-5 size-5 animate-bounce text-muted-foreground" />
                    ) : null}
                  </section>
                </div>
              ) : null}

              {showChoicePanel && choiceTool ? (
                <section
                  aria-label="行动选项"
                  className="absolute inset-x-4 top-1/2 z-20 mx-auto flex max-w-2xl -translate-y-1/2 flex-col gap-2 md:inset-x-10"
                >
                  {choiceTool.options.map((option) => (
                    <Button
                      key={option.id}
                      type="button"
                      variant="outline"
                      className="h-auto min-h-11 justify-center whitespace-normal bg-card/72 px-4 py-2.5 text-center text-base leading-6 shadow-lg shadow-black/25 backdrop-blur-md transition-[background,border-color,transform] hover:-translate-y-0.5 hover:bg-card/90"
                      disabled={Boolean(pendingChoiceId)}
                      onClick={() => void submitChoice(option.id)}
                    >
                      {pendingChoiceId === option.id ? <Loader2 className="animate-spin" data-icon="inline-start" /> : null}
                      {option.label}
                    </Button>
                  ))}
                </section>
              ) : null}
            </>
          )}
        </section>

        {!showGameOver ? (
          <PlaySidebar
            gameTitle={game.title}
            mapImage={mapImage}
            sceneMarkers={sceneMarkers}
            activeSceneLabel={activeScene?.name || activeSceneId || '未选择'}
            historyItems={historyItems}
            open={isSidebarOpen}
            onOpenChange={setIsSidebarOpen}
          />
        ) : null}
      </main>

      <Dialog open={isLoadOpen} onOpenChange={setIsLoadOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>读取存档</DialogTitle>
            <DialogDescription>载入快照会从该存档点新开一条进度，原快照保持不变。</DialogDescription>
          </DialogHeader>
          <div className="max-h-[50vh] overflow-y-auto">
            {snapshots === undefined ? (
              <div className="flex items-center gap-2 py-6 text-muted-foreground">
                <Loader2 className="animate-spin size-4" />
                <span>加载存档...</span>
              </div>
            ) : snapshots.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">还没有快照存档，先点左侧「保存快照」。</p>
            ) : (
              <div className="flex flex-col gap-2">
                {snapshots.map((snapshot) => (
                  <div key={snapshot.id} className="flex items-start gap-2 rounded-md border bg-card/60 p-3">
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{snapshot.label || '未命名存档'}</p>
                      <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">{snapshot.preview || '（无预览）'}</p>
                      <p className="mt-1 text-[11px] text-muted-foreground">
                        {snapshot.turnCount} 回合 · {formatSaveTime(snapshot.updatedAt)}
                        {snapshot.state === 'ended' ? ' · 已结束' : ''}
                      </p>
                    </div>
                    <div className="flex shrink-0 flex-col gap-1.5">
                      <Button
                        type="button"
                        size="sm"
                        className="h-8"
                        disabled={Boolean(loadBusyId)}
                        onClick={() => void handleLoadSnapshot(snapshot.id)}
                      >
                        {loadBusyId === snapshot.id ? <Loader2 className="animate-spin" data-icon="inline-start" /> : null}
                        载入
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        className="h-8"
                        disabled={Boolean(loadBusyId)}
                        onClick={() => void handleDeleteSnapshot(snapshot.id)}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
      <PlaySettingsModal open={isSettingsOpen} onOpenChange={setIsSettingsOpen} />
    </div>
  )
}

function formatSaveTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

type GameOverScreenProps = {
  endingTitle: string
  showTitle: boolean
  onReplay: () => void
}

function GameOverScreen({ endingTitle, showTitle, onReplay }: GameOverScreenProps) {
  return (
    <section
      aria-label="游戏结束"
      className="absolute inset-0 z-10 flex items-end justify-center px-6 pb-10 pt-24 text-center text-white md:pb-14"
    >
      <div className="absolute inset-0 bg-black/10" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_46%,transparent_0,transparent_42%,rgba(0,0,0,0.58)_92%)]" />
      <div className="relative flex w-full max-w-md flex-col items-center gap-3 rounded-md border border-white/18 bg-black/36 px-5 py-4 shadow-2xl shadow-black/40 backdrop-blur-sm">
        {showTitle ? (
          <h2 className="text-4xl font-semibold leading-none text-white drop-shadow-[0_0_18px_rgba(255,255,255,0.28)] md:text-6xl">
            GAME OVER
          </h2>
        ) : (
          <h2 className="sr-only">GAME OVER</h2>
        )}
        <div className="w-full">
          <p className="text-xs text-white/58">结局名称</p>
          <p className="mt-1 break-words text-lg font-medium leading-7 text-white md:text-xl">{endingTitle}</p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="lg"
          className="mt-1 border-white/40 bg-black/35 px-6 text-white shadow-xl shadow-black/35 backdrop-blur-md hover:bg-white/12 hover:text-white"
          onClick={onReplay}
        >
          <RefreshCcw data-icon="inline-start" />
          重玩
        </Button>
      </div>
    </section>
  )
}

function findGameOverImage(game: roleplay.LibraryGame) {
  const gameOverScene = game.scenes?.find((scene) => normalizeGameOverKey(scene.id) === 'gameover')
  if (gameOverScene?.url) {
    return gameOverScene.url
  }

  const metadataURL = findGameOverImageFromMetadata(game.files)
  if (metadataURL) {
    return metadataURL
  }

  return Object.entries(game.files ?? {}).find(([name, url]) => {
    const key = normalizeAssetPath(name)
    if (!key.startsWith('photo/') || key.endsWith('/metadata.json') || key.endsWith('metadata.json')) {
      return false
    }
    const fileName = key.slice('photo/'.length).replace(/\.[^.]+$/, '')
    return normalizeGameOverKey(fileName) === 'gameover' && isImageURL(url)
  })?.[1]
}

function findGameOverImageFromMetadata(files: Record<string, string>) {
  const raw = files[normalizeAssetPath('photo/metadata.json')]
  if (!raw) {
    return undefined
  }

  try {
    const mapping = JSON.parse(raw) as Record<string, string>
    const mappedFile = Object.entries(mapping).find(([id]) => normalizeGameOverKey(id) === 'gameover')?.[1]
    if (!mappedFile) {
      return undefined
    }
    const url = files[normalizeAssetPath(`photo/${mappedFile}`)]
    return isImageURL(url) ? url : undefined
  }
  catch {
    return undefined
  }
}

function normalizeAssetPath(value: string) {
  return value.trim().replaceAll('\\', '/').replace(/^\.?\//, '').toLowerCase()
}

function normalizeGameOverKey(value: string) {
  return value.trim().toLowerCase().replace(/[-_\s]/g, '')
}

function isImageURL(value?: string) {
  return Boolean(value && (value.startsWith('/local/') || value.startsWith('data:image/') || /^https?:\/\//.test(value)))
}

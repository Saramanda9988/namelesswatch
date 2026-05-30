import { Link, useNavigate, useParams } from '@tanstack/react-router'
import {
  ArrowLeft,
  Bot,
  CheckCircle2,
  ChevronDown,
  Clock3,
  Loader2,
  Map,
  Moon,
  RefreshCcw,
  Settings,
  Sparkles,
  SunMedium,
  Triangle,
} from 'lucide-react'
import * as React from 'react'

import { RegisterGamePack, StartGame, SubmitChoice } from '../../wailsjs/go/main/App'
import { LogError, LogInfo } from '../../wailsjs/runtime/runtime'
import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useGameStore } from '@/stores/game-store'
import type { roleplay } from '../../wailsjs/go/models'

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

const eventMemoryPlaceholders = [
  'ai引导对话和主人公选择语句',
  'ai引导对话和主人公选择语句',
  'ai引导对话和主人公选择语句',
  'ai引导对话和主人公选择语句',
  'ai引导对话和主人公选择语句',
  'ai引导对话和主人公选择语句',
]

export function PlayPage() {
  const { gameId } = useParams({ from: '/play/$gameId' })
  const navigate = useNavigate()
  const games = useGameStore((state) => state.games)
  const setActiveGame = useGameStore((state) => state.setActiveGame)
  const game = games.find((item) => item.id === gameId)
  const [sessionId, setSessionId] = React.useState<string>()
  const [latestResult, setLatestResult] = React.useState<roleplay.GameTurnResult>()
  const [turns, setTurns] = React.useState<roleplay.GameTurn[]>([])
  const [error, setError] = React.useState<string>()
  const [isStarting, setIsStarting] = React.useState(true)
  const [pendingChoiceId, setPendingChoiceId] = React.useState<string>()
  const startedGameIdRef = React.useRef<string | undefined>(undefined)

  React.useEffect(() => {
    if (!game) {
      return
    }
    setActiveGame(game.id)
  }, [game, setActiveGame])

  React.useEffect(() => {
    if (!game || startedGameIdRef.current === game.id) {
      return
    }

    let cancelled = false
    const currentGame = game
    startedGameIdRef.current = game.id
    setIsStarting(true)
    setError(undefined)
    setTurns([])
    setLatestResult(undefined)
    setSessionId(undefined)

    async function start() {
      try {
        logRuntimeInfo(`[play] start requested game=${currentGame.id} title=${currentGame.title}`)
        await RegisterGamePack(currentGame.id, currentGame.files)
        const result = await StartGame(currentGame.id)
        if (cancelled) {
          return
        }
        logRuntimeInfo(`[play] start result game=${currentGame.id} session=${result.sessionId} state=${result.state} tools=${result.tools?.length ?? 0} ending=${Boolean(result.ending)}`)
        setSessionId(result.sessionId)
        setLatestResult(result)
        setTurns([result.turn])
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
      if (startedGameIdRef.current === currentGame.id) {
        startedGameIdRef.current = undefined
      }
    }
  }, [game])

  async function submitChoice(choiceId: string) {
    if (!sessionId || pendingChoiceId || latestResult?.state === 'ended') {
      return
    }

    setPendingChoiceId(choiceId)
    setError(undefined)
    try {
      logRuntimeInfo(`[play] submit choice session=${sessionId} choice=${choiceId}`)
      const result = await SubmitChoice(sessionId, choiceId)
      logRuntimeInfo(`[play] choice result session=${sessionId} state=${result.state} tools=${result.tools?.length ?? 0} ending=${Boolean(result.ending)}`)
      setLatestResult(result)
      setTurns((currentTurns) => [...currentTurns, result.turn])
    }
    catch (cause) {
      logRuntimeError(`[play] submit choice failed session=${sessionId} choice=${choiceId} error=${cause instanceof Error ? cause.message : String(cause)}`)
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setPendingChoiceId(undefined)
    }
  }

  const renderableTurns = React.useMemo(() => turns.filter(isRenderableTurn), [turns])
  const choiceTool = choiceToolFrom(latestResult)
  const isEnded = latestResult?.state === 'ended'
  const currentTurn = renderableTurns.at(-1)
  const currentLines = currentTurn?.payload ?? []
  const sceneImage = game?.photoUrls?.[0]
  const mapImage = game?.mapUrls?.[0]

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
      <main className="grid h-full grid-cols-1 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section className="relative min-h-0 overflow-hidden bg-card">
          {sceneImage ? (
            <img src={sceneImage} alt="" className="absolute inset-0 size-full object-cover" />
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
            <Button type="button" variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="昼夜">
              <SunMedium data-icon />
            </Button>
            <Button type="button" variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="夜间">
              <Moon data-icon />
            </Button>
          </nav>

          <div className="absolute right-4 top-4 z-20 flex items-center gap-2 lg:hidden">
            <Badge variant={isEnded ? 'default' : 'secondary'}>{isEnded ? '已结束' : 'AI 主持'}</Badge>
            <Button asChild variant="outline" size="icon-lg" className="bg-background/55 backdrop-blur-md" aria-label="设置">
              <Link to="/settings">
                <Settings data-icon />
              </Link>
            </Button>
          </div>

          <div className="absolute left-1/2 top-4 z-10 hidden -translate-x-1/2 items-center gap-2 rounded-md border bg-background/55 px-3 py-2 text-sm text-muted-foreground backdrop-blur-md md:flex">
            <Bot className="size-4" />
            <span className="max-w-[44vw] truncate">{game.title}</span>
            <Separator orientation="vertical" className="h-4" />
            <span>{sessionId ? `Session ${sessionId.slice(-6)}` : '准备会话'}</span>
          </div>

          <div className="absolute inset-x-0 bottom-0 z-20 px-4 pb-5 md:px-8 md:pb-8">
            <section className="mx-auto flex max-w-4xl flex-col gap-3 rounded-lg border bg-background/78 p-4 shadow-2xl backdrop-blur-md md:p-5">
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
                  <ScrollArea className="max-h-[25vh] pr-3">
                    <div className="flex flex-col gap-2 text-base leading-7 text-foreground md:text-lg">
                      {currentLines.map((line, lineIndex) => (
                        <p key={`${currentTurn?.id ?? 'turn'}-${lineIndex}`} className="text-pretty">
                          {line}
                        </p>
                      ))}
                    </div>
                  </ScrollArea>
                ) : null}

                {!isStarting && !error && currentLines.length === 0 ? (
                  <p className="text-base leading-7 text-muted-foreground">等待 AI 主持生成开场叙事...</p>
                ) : null}
              </div>

              {choiceTool && !isEnded ? (
                <>
                  <Separator />
                  <div className="flex flex-col gap-3">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Sparkles className="size-4" />
                      <span>{choiceTool.prompt || '你要怎么做？'}</span>
                    </div>
                    <div className="grid gap-2 md:grid-cols-2">
                      {choiceTool.options.map((option) => (
                        <Button
                          key={option.id}
                          type="button"
                          variant="outline"
                          className="h-auto min-h-10 justify-start whitespace-normal bg-card/70 px-3 py-2 text-left"
                          disabled={Boolean(pendingChoiceId)}
                          onClick={() => void submitChoice(option.id)}
                        >
                          {pendingChoiceId === option.id ? <Loader2 className="animate-spin" data-icon="inline-start" /> : null}
                          {option.label}
                        </Button>
                      ))}
                    </div>
                  </div>
                </>
              ) : null}

              {isEnded && latestResult?.ending ? (
                <>
                  <Separator />
                  <div className="flex items-start gap-3">
                    <Badge>
                      <CheckCircle2 data-icon />
                      结局
                    </Badge>
                    <div className="min-w-0">
                      <p className="truncate font-medium">{latestResult.ending.title}</p>
                      <p className="mt-1 text-sm text-muted-foreground">普通推进已停止。</p>
                    </div>
                  </div>
                </>
              ) : null}

              <ChevronDown className="absolute bottom-4 right-5 size-5 text-muted-foreground" />
            </section>
          </div>
        </section>

        <aside className="hidden min-h-0 flex-col border-l bg-background lg:flex">
          <div className="flex h-full min-h-0 flex-col gap-5 px-4 py-5">
            <div className="flex items-center justify-between gap-3">
              <div className="flex min-w-0 items-center gap-2">
                <Badge variant="outline" className="size-7 justify-center rounded-md p-0">
                  <Map data-icon />
                </Badge>
                <div className="min-w-0">
                  <h2 className="truncate text-xl font-semibold">地图简览</h2>
                  <p className="truncate text-xs text-muted-foreground">{game.title}</p>
                </div>
              </div>
              <Button asChild variant="ghost" size="icon-lg" aria-label="设置">
                <Link to="/settings">
                  <Settings data-icon />
                </Link>
              </Button>
            </div>

            <AspectRatio ratio={1.54} className="overflow-hidden rounded-md border bg-card">
              {mapImage ? (
                <img src={mapImage} alt="" className="size-full object-cover" />
              ) : (
                <MiniMapPlaceholder />
              )}
            </AspectRatio>

            <Separator />

            <section className="flex min-h-0 flex-1 flex-col gap-4">
              <div className="flex items-center gap-2">
                <Triangle className="size-4 fill-current text-muted-foreground" />
                <h2 className="text-xl font-semibold">事件回忆</h2>
              </div>
              <div className="rounded-md border bg-card/50 px-3 py-3 text-sm font-medium text-muted-foreground">
                【场景画外音.................................】
              </div>
              <ScrollArea className="min-h-0 flex-1 pr-3">
                <div className="flex flex-col gap-4 pb-4">
                  {eventMemoryPlaceholders.map((item, index) => (
                    <div key={`${item}-${index}`} className="flex items-start gap-2 text-sm leading-6 text-muted-foreground">
                      <Clock3 className="mt-1 size-3.5 shrink-0" />
                      <p>{item}</p>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </section>
          </div>
        </aside>
      </main>
    </div>
  )
}

function MiniMapPlaceholder() {
  return (
    <div className="grid size-full grid-cols-[0.7fr_1.05fr_0.7fr] grid-rows-[0.62fr_0.76fr_0.72fr] gap-1 bg-background p-1">
      <div className="border bg-card" />
      <div className="border bg-muted" />
      <div className="relative border bg-card">
        <div className="absolute right-1 top-1 h-5 w-8 bg-muted" />
      </div>
      <div className="relative border bg-card">
        <div className="absolute left-0 top-2 h-6 w-14 bg-muted" />
        <div className="absolute bottom-1 left-1 h-5 w-5 bg-muted" />
      </div>
      <div className="relative row-span-2 border bg-card">
        <div className="absolute left-0 top-2 h-14 w-6 bg-muted" />
        <div className="absolute left-0 top-1 h-5 w-16 bg-muted" />
      </div>
      <div className="border bg-card" />
      <div className="relative border bg-card">
        <div className="absolute left-0 top-2 h-7 w-16 bg-muted" />
      </div>
      <div className="relative border bg-card">
        <div className="absolute right-0 top-3 h-8 w-1 bg-muted" />
      </div>
    </div>
  )
}

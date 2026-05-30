import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { ArrowLeft, Bot, CheckCircle2, Loader2, RefreshCcw, Settings, Sparkles } from 'lucide-react'
import * as React from 'react'

import { RegisterGamePack, StartGame, SubmitChoice } from '../../wailsjs/go/main/App'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useGameStore } from '@/stores/game-store'
import type { ChoiceTool, GameTurn, GameTurnResult } from '@/types/game'

function isRenderableTurn(turn: GameTurn) {
  return turn.role === 'ai' && turn.payload.length > 0
}

function choiceToolFrom(result?: GameTurnResult): ChoiceTool | undefined {
  if (!result || result.state === 'ended') {
    return undefined
  }
  return result.tools?.find((tool) => tool.type === 'choice')
}

export function PlayPage() {
  const { gameId } = useParams({ from: '/play/$gameId' })
  const navigate = useNavigate()
  const games = useGameStore((state) => state.games)
  const setActiveGame = useGameStore((state) => state.setActiveGame)
  const game = games.find((item) => item.id === gameId)
  const [sessionId, setSessionId] = React.useState<string>()
  const [latestResult, setLatestResult] = React.useState<GameTurnResult>()
  const [turns, setTurns] = React.useState<GameTurn[]>([])
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
        await RegisterGamePack(currentGame.id, currentGame.files)
        const result = await StartGame(currentGame.id) as GameTurnResult
        if (cancelled) {
          return
        }
        setSessionId(result.sessionId)
        setLatestResult(result)
        setTurns([result.turn])
      }
      catch (cause) {
        if (!cancelled) {
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
  }, [game])

  async function submitChoice(choiceId: string) {
    if (!sessionId || pendingChoiceId || latestResult?.state === 'ended') {
      return
    }

    setPendingChoiceId(choiceId)
    setError(undefined)
    try {
      const result = await SubmitChoice(sessionId, choiceId) as GameTurnResult
      setLatestResult(result)
      setTurns((currentTurns) => [...currentTurns, result.turn])
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setPendingChoiceId(undefined)
    }
  }

  const renderableTurns = React.useMemo(() => turns.filter(isRenderableTurn), [turns])
  const choiceTool = choiceToolFrom(latestResult)
  const isEnded = latestResult?.state === 'ended'

  if (!game) {
    return (
      <div className="grid min-h-screen place-items-center bg-background p-6 text-foreground">
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
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b bg-card">
        <div className="mx-auto flex h-14 w-full max-w-6xl items-center justify-between gap-3 px-4 md:px-6">
          <div className="flex min-w-0 items-center gap-2">
            <Button asChild variant="ghost" size="icon" aria-label="返回">
              <Link to="/">
                <ArrowLeft data-icon />
              </Link>
            </Button>
            <Separator orientation="vertical" className="h-6" />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{game.title}</p>
              <p className="text-xs text-muted-foreground">{sessionId ? `Session ${sessionId.slice(-6)}` : '准备会话'}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Badge variant={isEnded ? 'default' : 'secondary'}>
              {isEnded ? '已结束' : 'AI 主持'}
            </Badge>
            <Button asChild variant="ghost" size="icon" aria-label="设置">
              <Link to="/settings">
                <Settings data-icon />
              </Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto grid h-[calc(100vh-3.5rem)] w-full max-w-6xl grid-rows-[1fr_auto] px-4 py-4 md:px-6">
        <ScrollArea className="min-h-0 pr-3">
          <div className="flex flex-col gap-4 pb-6">
            {isStarting ? (
              <Card>
                <CardHeader className="flex-row items-center gap-3">
                  <Loader2 className="animate-spin" data-icon />
                  <div>
                    <CardTitle className="text-base">正在生成首回合</CardTitle>
                    <CardDescription>后端会创建独立 workspace，并把剧情文档复制到当前会话。</CardDescription>
                  </div>
                </CardHeader>
              </Card>
            ) : null}

            {error ? (
              <Card className="border-destructive/40">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base text-destructive">
                    <RefreshCcw data-icon />
                    会话暂不可用
                  </CardTitle>
                  <CardDescription className="break-words">{error}</CardDescription>
                </CardHeader>
                <CardContent>
                  <Button type="button" variant="outline" onClick={() => void navigate({ to: '/' })}>
                    返回游戏库
                  </Button>
                </CardContent>
              </Card>
            ) : null}

            {renderableTurns.map((turn, index) => (
              <section key={turn.id} className="flex flex-col gap-3 rounded-lg border bg-card p-4 shadow-sm">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="size-8 justify-center rounded-full p-0">
                      <Bot data-icon />
                    </Badge>
                    <div>
                      <p className="text-sm font-medium">回合 {index + 1}</p>
                      <p className="text-xs text-muted-foreground">AI 叙事输出</p>
                    </div>
                  </div>
                  {turn.ending ? (
                    <Badge>
                      <CheckCircle2 data-icon />
                      {turn.ending.kind}
                    </Badge>
                  ) : null}
                </div>
                <div className="flex flex-col gap-3 text-base leading-7">
                  {turn.payload.map((line, lineIndex) => (
                    <p key={`${turn.id}-${lineIndex}`} className="text-pretty">
                      {line}
                    </p>
                  ))}
                </div>
                {turn.ending ? (
                  <>
                    <Separator />
                    <div className="flex flex-col gap-1">
                      <p className="text-sm font-medium">{turn.ending.title}</p>
                      <p className="text-xs text-muted-foreground">Ending ID: {turn.ending.id}</p>
                    </div>
                  </>
                ) : null}
              </section>
            ))}
          </div>
        </ScrollArea>

        <section className="border-t bg-background pt-4">
          {choiceTool && !isEnded ? (
            <div className="flex flex-col gap-3">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Sparkles data-icon />
                <span>{choiceTool.prompt || '你要怎么做？'}</span>
              </div>
              <div className="grid gap-2 md:grid-cols-2">
                {choiceTool.options.map((option) => (
                  <Button
                    key={option.id}
                    type="button"
                    variant="outline"
                    className="h-auto justify-start whitespace-normal py-3 text-left"
                    disabled={Boolean(pendingChoiceId)}
                    onClick={() => void submitChoice(option.id)}
                  >
                    {pendingChoiceId === option.id ? <Loader2 className="animate-spin" data-icon="inline-start" /> : null}
                    {option.label}
                  </Button>
                ))}
              </div>
            </div>
          ) : null}

          {isEnded && latestResult?.ending ? (
            <div className="flex flex-col gap-2 rounded-lg border bg-card p-4">
              <Badge className="w-fit">
                <CheckCircle2 data-icon />
                结局
              </Badge>
              <div>
                <p className="font-medium">{latestResult.ending.title}</p>
                <p className="mt-1 text-sm text-muted-foreground">普通推进已停止。</p>
              </div>
            </div>
          ) : null}
        </section>
      </main>
    </div>
  )
}

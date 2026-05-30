import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight, MapPinned, RotateCcw, Settings } from 'lucide-react'
import * as React from 'react'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { useGameStore } from '@/stores/game-store'

export function PlayPage() {
  const { gameId } = useParams({ from: '/play/$gameId' })
  const navigate = useNavigate()
  const games = useGameStore((state) => state.games)
  const settings = useGameStore((state) => state.settings)
  const setActiveGame = useGameStore((state) => state.setActiveGame)
  const [lineIndex, setLineIndex] = React.useState(0)

  const game = games.find((item) => item.id === gameId)
  const currentLine = game?.script[lineIndex]
  const currentMap = game?.mapUrls[lineIndex % Math.max(game.mapUrls.length, 1)]

  React.useEffect(() => {
    if (game) {
      setActiveGame(game.id)
    }
  }, [game, setActiveGame])

  React.useEffect(() => {
    if (!settings.autoAdvance || !game) {
      return
    }

    const timeout = window.setTimeout(() => {
      setLineIndex((index) => Math.min(index + 1, Math.max(game.script.length - 1, 0)))
    }, 3200)

    return () => window.clearTimeout(timeout)
  }, [game, lineIndex, settings.autoAdvance])

  function advance() {
    if (!game) {
      void navigate({ to: '/' })
      return
    }

    setLineIndex((index) => Math.min(index + 1, Math.max(game.script.length - 1, 0)))
  }

  if (!game) {
    return (
      <div className="grid min-h-screen place-items-center bg-[#10130f] p-6 text-[#efe8d3]">
        <div className="max-w-sm border border-[#efe8d3]/18 bg-[#0d100d] p-6 text-center">
          <h1 className="text-2xl font-semibold">未找到游戏</h1>
          <Button asChild className="mt-5 rounded-none bg-[#d2a84f] text-[#171307] hover:bg-[#efc767]">
            <Link to="/">返回主页</Link>
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="relative h-screen overflow-hidden bg-[#090b09] text-[#f7f0dd]">
      {currentLine?.backgroundUrl ? (
        <img src={currentLine.backgroundUrl} alt="" className="absolute inset-0 size-full object-cover" />
      ) : (
        <div className="absolute inset-0 bg-[linear-gradient(125deg,#151f19,#332117_48%,#0a0c0a)]" />
      )}
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(3,4,3,0.22),rgba(3,4,3,0.18)_43%,rgba(3,4,3,0.78))]" />

      <header className="absolute left-4 right-4 top-4 z-10 flex items-start justify-between gap-4">
        <div className="flex gap-2">
          <Button asChild variant="outline" className="rounded-none border-white/20 bg-black/32 text-white backdrop-blur hover:bg-white/12">
            <Link to="/">
              <ArrowLeft className="size-4" />
              主页
            </Link>
          </Button>
          <Button asChild variant="outline" className="rounded-none border-white/20 bg-black/32 text-white backdrop-blur hover:bg-white/12">
            <Link to="/settings">
              <Settings className="size-4" />
              设置
            </Link>
          </Button>
        </div>

        {settings.showMap && (
          <Card className="w-28 rounded-none border-[#d2a84f]/45 bg-black/45 p-1 shadow-2xl backdrop-blur md:w-44">
            <AspectRatio ratio={1} className="relative overflow-hidden bg-[#151c16]">
              {currentMap ? (
                <img src={currentMap} alt="" className="size-full object-cover" />
              ) : (
                <div className="grid size-full place-items-center bg-[linear-gradient(135deg,#1d2c25,#3a281b)] text-[#d2a84f]">
                  <MapPinned className="size-8" />
                </div>
              )}
              <Badge variant="outline" className="absolute left-2 top-2 rounded-none border-[#d2a84f]/50 bg-black/55 text-[10px] uppercase tracking-[0.22em] text-[#f2d37b]">
                Map
              </Badge>
            </AspectRatio>
          </Card>
        )}
      </header>

      <section
        className="absolute inset-x-3 bottom-3 z-10 md:inset-x-8 md:bottom-7"
        style={{ transform: `scale(${settings.uiScale / 100})`, transformOrigin: 'bottom center' }}
      >
        <Card
          role="button"
          tabIndex={0}
          aria-label="推进文本"
          onClick={advance}
          onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              advance()
            }
          }}
          className="w-full cursor-pointer rounded-none border-[#d2a84f]/35 bg-[#0b0d0b]/82 p-0 text-left text-[#f7f0dd] shadow-[0_24px_80px_rgba(0,0,0,0.45)] backdrop-blur-md"
        >
          <CardHeader className="flex-row items-center justify-between gap-3 p-4 pb-0 md:p-6 md:pb-0">
            <CardTitle className="border-l-4 border-[#c2341c] pl-3 text-lg font-semibold text-[#fff6dd] md:text-xl">
              {currentLine?.speaker || '旁白'}
            </CardTitle>
            <Badge variant="ghost" className="rounded-none text-xs tabular-nums text-[#efe8d3]/48">
              {Math.min(lineIndex + 1, game.script.length)} / {game.script.length}
            </Badge>
          </CardHeader>

          <CardContent className="p-4 md:p-6">
            <p className="min-h-20 text-pretty text-base leading-8 text-[#efe8d3] md:min-h-24 md:text-xl md:leading-9">
              {currentLine?.text || '场景文件为空。'}
            </p>
          </CardContent>

          <CardFooter className="flex items-center justify-between rounded-none border-white/10 bg-white/[0.03] p-4 pt-3 text-xs text-[#efe8d3]/48 md:px-6">
            <span>{game.title}</span>
            <span className="flex items-center gap-2">
              {lineIndex >= game.script.length - 1 ? (
                <>
                  <RotateCcw className="size-3.5" />
                  终章
                </>
              ) : (
                <>
                  <ChevronRight className="size-3.5" />
                  下一段
                </>
              )}
            </span>
          </CardFooter>
        </Card>
      </section>
    </div>
  )
}

import { Link } from '@tanstack/react-router'
import { BookOpen, Filter, FolderPlus, Gamepad2, Play, Search, Settings, Upload } from 'lucide-react'
import * as React from 'react'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { importGameFromFiles } from '@/lib/import-game'
import { useGameStore } from '@/stores/game-store'

const directoryInputProps = {
  directory: '',
  webkitdirectory: '',
} as React.InputHTMLAttributes<HTMLInputElement> & {
  directory: string
  webkitdirectory: string
}

export function HomePage() {
  const inputRef = React.useRef<HTMLInputElement>(null)
  const games = useGameStore((state) => state.games)
  const addGame = useGameStore((state) => state.addGame)
  const [status, setStatus] = React.useState('等待资源包')
  const [isImporting, setIsImporting] = React.useState(false)
  const [search, setSearch] = React.useState('')

  const visibleGames = React.useMemo(() => {
    const keyword = search.trim().toLowerCase()
    if (!keyword) {
      return games
    }

    return games.filter((game) => {
      const firstLine = (game.files['scene.md'] || game.script[0]?.text || '').toLowerCase()
      return game.title.toLowerCase().includes(keyword) || firstLine.includes(keyword)
    })
  }, [games, search])

  async function handleImport(event: React.ChangeEvent<HTMLInputElement>) {
    const selectedFiles = event.currentTarget.files
    if (!selectedFiles || selectedFiles.length === 0) {
      return
    }

    setIsImporting(true)
    const report = await importGameFromFiles(selectedFiles)

    if (!report.game) {
      setStatus(`缺少文件：${report.missing.join('、')}；已识别：${report.validFiles.join('、') || '无'}`)
      setIsImporting(false)
      event.currentTarget.value = ''
      return
    }

    addGame(report.game)
    setStatus(report.warnings.length > 0 ? report.warnings.join(' ') : `已导入并验证：${report.game.title}`)
    setIsImporting(false)
    event.currentTarget.value = ''
  }

  return (
    <div className="min-h-screen bg-[#111315] text-[#f4f5f7]">
      <Input ref={inputRef} type="file" className="hidden" multiple onChange={handleImport} {...directoryInputProps} />

      <header className="sticky top-0 z-20 border-b border-[#2b3037] bg-[#1a1d20]">
        <div className="flex h-16 items-center justify-between px-5 md:px-8">
          <Link to="/" className="flex items-center gap-3">
            <Badge variant="outline" className="grid size-10 place-items-center rounded-xl border-[#596170] bg-[#23272d] p-0 text-[#dce2ea]">
              <BookOpen className="size-5" />
            </Badge>
            <div>
              <div className="text-lg font-semibold leading-none">LunaBox</div>
              <div className="mt-1 text-xs text-[#a8b0bd]">AI Roleplay Library</div>
            </div>
          </Link>

          <div className="flex items-center gap-3">
            <Badge variant="outline" className="hidden h-9 rounded-md border-[#343a43] bg-[#111315] px-3 text-[#a8b0bd] md:inline-flex">
              {status}
            </Badge>
            <Button
              type="button"
              className="h-10 rounded-lg bg-[#59677e] px-4 text-white hover:bg-[#6b7b94]"
              disabled={isImporting}
              onClick={() => inputRef.current?.click()}
            >
              <FolderPlus className="size-4" />
              {isImporting ? '导入中' : '添加游戏'}
            </Button>
            <Button asChild variant="outline" size="icon-lg" className="rounded-lg border-[#343a43] bg-[#22262b] text-[#dce2ea] hover:bg-[#2d3239]">
              <Link to="/settings" aria-label="设置">
                <Settings className="size-5" />
              </Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto flex h-[calc(100vh-64px)] w-full max-w-[1760px] flex-col px-5 py-8 md:px-12">
        <section className="space-y-7">
          <div className="flex flex-col gap-5 md:flex-row md:items-end md:justify-between">
            <div>
              <h1 className="text-5xl font-bold tracking-normal text-white">游戏库</h1>
              <p className="mt-3 text-xl text-[#9ea6b2]">共 {games.length} 个游戏</p>
            </div>
          </div>
        </section>

        <Separator className="my-7 bg-[#252a31]" />

        <ScrollArea className="min-h-0 flex-1 pr-4">
          {games.length === 0 ? (
            <Card className="grid min-h-[420px] place-items-center rounded-xl border-[#343a43] bg-[#171a1e] text-center text-[#f4f5f7]">
              <CardContent className="max-w-md space-y-5 p-8">
                <div>
                  <CardTitle className="text-2xl text-white">暂无游戏</CardTitle>
                  <CardDescription className="mt-3 text-sm leading-6 text-[#9ea6b2]">
                    去添加一个游戏吧
                  </CardDescription>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-[repeat(auto-fill,minmax(188px,1fr))] gap-5 pb-8">
              {visibleGames.map((game) => (
                <Card key={game.id} className="group overflow-hidden rounded-xl border-[#343a43] bg-[#1a1d20] p-0 text-[#f4f5f7] transition-colors hover:border-[#59677e]">
                  <AspectRatio ratio={0.74} className="relative bg-[#22262b]">
                    {game.photoUrls[0] ? (
                      <img src={game.photoUrls[0]} alt="" className="size-full object-cover transition duration-300 group-hover:scale-105" />
                    ) : (
                      <div className="grid size-full place-items-center bg-[#23272d] text-[#596170]">
                        <Gamepad2 className="size-12" />
                      </div>
                    )}
                    <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-[#1a1d20] to-transparent" />
                    <Button asChild size="icon" className="absolute right-3 top-3 rounded-full bg-[#59677e] text-white opacity-0 shadow-lg transition-opacity hover:bg-[#6b7b94] group-hover:opacity-100">
                      <Link to="/play/$gameId" params={{ gameId: game.id }} aria-label={`开始 ${game.title}`}>
                        <Play className="size-4 fill-current" />
                      </Link>
                    </Button>
                  </AspectRatio>

                  <CardHeader className="gap-1 px-3 pb-3 pt-3">
                    <CardTitle className="truncate text-lg font-bold text-white">{game.title}</CardTitle>
                    <CardDescription className="truncate text-sm text-[#9ea6b2]">
                      {Object.keys(game.files).filter((name) => name.endsWith('.md')).length} 个剧情文档
                    </CardDescription>
                  </CardHeader>
                </Card>
              ))}
            </div>
          )}
        </ScrollArea>
      </main>
    </div>
  )
}

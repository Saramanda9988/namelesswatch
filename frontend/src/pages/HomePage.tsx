import { Link, useNavigate } from '@tanstack/react-router'
import { BookOpen, FolderPlus, Gamepad2, History, Play, Settings, Trash2 } from 'lucide-react'
import * as React from 'react'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useGameStore } from '@/stores/game-store'
import { LogInfo } from '../../wailsjs/runtime/runtime'

const directoryInputProps = {
  directory: '',
  webkitdirectory: '',
} as React.InputHTMLAttributes<HTMLInputElement> & {
  directory: string
  webkitdirectory: string
}

function logRuntimeInfo(message: string) {
  try {
    LogInfo(message)
  }
  catch {
    console.info(message)
  }
}

export function HomePage() {
  const inputRef = React.useRef<HTMLInputElement>(null)
  const navigate = useNavigate()
  const games = useGameStore((state) => state.games)
  const fetchGames = useGameStore((state) => state.fetchGames)
  const importGameFiles = useGameStore((state) => state.importGameFiles)
  const deleteGame = useGameStore((state) => state.deleteGame)
  const listSessions = useGameStore((state) => state.listSessions)
  const setPendingResumeSession = useGameStore((state) => state.setPendingResumeSession)
  const [status, setStatus] = React.useState('等待资源包')
  const [isImporting, setIsImporting] = React.useState(false)
  const [deletingGameId, setDeletingGameId] = React.useState<string>()
  const [search, setSearch] = React.useState('')
  const [continueByGame, setContinueByGame] = React.useState<Record<string, string>>({})

  React.useEffect(() => {
    void fetchGames()
  }, [fetchGames])

  React.useEffect(() => {
    let cancelled = false

    async function loadSaves() {
      const entries = await Promise.all(
        games.map(async (game) => {
          try {
            const sessions = await listSessions(game.id)
            const resumable = sessions.find((session) => !session.isSnapshot)
            return [game.id, resumable?.id] as const
          }
          catch {
            return [game.id, undefined] as const
          }
        }),
      )
      if (cancelled) {
        return
      }
      const next: Record<string, string> = {}
      for (const [gameId, sessionId] of entries) {
        if (sessionId) {
          next[gameId] = sessionId
        }
      }
      setContinueByGame(next)
    }

    if (games.length > 0) {
      void loadSaves()
    }
    else {
      setContinueByGame({})
    }

    return () => {
      cancelled = true
    }
  }, [games, listSessions])

  function handleContinue(gameId: string) {
    const sessionId = continueByGame[gameId]
    if (!sessionId) {
      return
    }
    setPendingResumeSession(sessionId)
    void navigate({ to: '/play/$gameId', params: { gameId } })
  }

  const visibleGames = React.useMemo(() => {
    const keyword = search.trim().toLowerCase()
    if (!keyword) {
      return games
    }

    return games.filter((game) => {
      const firstLine = (game.files?.['scene.md'] || '').toLowerCase()
      return game.title.toLowerCase().includes(keyword) || firstLine.includes(keyword)
    })
  }, [games, search])

  async function handleImport(event: React.ChangeEvent<HTMLInputElement>) {
    const selectedFiles = event.currentTarget.files
    if (!selectedFiles || selectedFiles.length === 0) {
      return
    }

    setIsImporting(true)
    try {
      const fileContents = await readStoryFolderFiles(selectedFiles)
      const keys = Object.keys(fileContents)
      logRuntimeInfo(`[import] prepared files count=${keys.length} keys=${summarizeKeys(keys)}`)
      const report = await importGameFiles(fileContents)
      logRuntimeInfo(`[import] result game=${report.game?.id ?? ''} title=${report.game?.title ?? ''} scenes=${report.game?.scenes?.length ?? 0} photos=${report.game?.photoUrls?.length ?? 0} missing=${report.missing.join('|')}`)

      if (!report.game) {
        setStatus(`缺少文件：${report.missing.join('、')}；已识别：${report.validFiles.join('、') || '无'}`)
        return
      }

      setStatus(report.warnings.length > 0 ? report.warnings.join(' ') : `已导入并验证：${report.game.title}`)
    }
    catch (cause) {
      setStatus(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setIsImporting(false)
      event.currentTarget.value = ''
    }
  }

  async function handleDeleteGame(gameId: string, title: string) {
    const confirmed = window.confirm(`删除「${title}」？`)
    if (!confirmed) {
      return
    }

    setDeletingGameId(gameId)
    try {
      await deleteGame(gameId)
      setStatus(`已删除：${title}`)
    }
    catch (cause) {
      setStatus(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setDeletingGameId(undefined)
    }
  }

  return (
    <div className="min-h-screen bg-[#111315] text-[#f4f5f7]">
      <Input ref={inputRef} type="file" className="hidden" multiple onChange={handleImport} {...directoryInputProps} />

      <header className="sticky top-0 z-20 border-b border-[#2b3037] bg-[#1a1d20]">
        <div className="flex h-16 items-center justify-between px-5 md:px-8">
          <Link to="/" className="flex items-center gap-3">
            <div>
              <div className="text-lg font-semibold leading-none">NamelessWatch</div>
              <div className="mt-1 text-xs text-[#a8b0bd]">AI Roleplay Library</div>
            </div>
          </Link>

          <div className="flex items-center gap-3">
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
                <Card key={game.id} className="group relative isolate overflow-hidden rounded-xl border-[#343a43] bg-[#1a1d20] p-0 text-[#f4f5f7] transition-colors hover:border-[#59677e]">
                  <AspectRatio ratio={0.74} className="relative overflow-hidden bg-[#22262b]">
                    {game.photoUrls?.[0] ? (
                      <img src={game.photoUrls[0]} alt="" className="block size-full origin-center object-cover transition-transform duration-300 ease-out will-change-transform group-hover:scale-105" />
                    ) : (
                      <div className="grid size-full place-items-center bg-[#23272d] text-[#596170]">
                        <Gamepad2 className="size-12" />
                      </div>
                    )}
                    <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-[#1a1d20] to-transparent" />
                    <div className="absolute right-3 top-3 flex gap-2 opacity-0 transition-opacity group-hover:opacity-100">
                      {continueByGame[game.id] ? (
                        <Button
                          type="button"
                          size="icon"
                          className="rounded-full bg-[#3f7d52] text-white shadow-lg hover:bg-[#4c9162]"
                          aria-label={`继续 ${game.title}`}
                          title="继续游戏（读取上次进度）"
                          onClick={() => handleContinue(game.id)}
                        >
                          <History className="size-4" />
                        </Button>
                      ) : null}
                      <Button asChild size="icon" className="rounded-full bg-[#59677e] text-white shadow-lg hover:bg-[#6b7b94]">
                        <Link
                          to="/play/$gameId"
                          params={{ gameId: game.id }}
                          aria-label={`开始 ${game.title}`}
                          title={continueByGame[game.id] ? '重新开始（开新游戏）' : '开始游戏'}
                          onClick={() => setPendingResumeSession(undefined)}
                        >
                          <Play className="size-4 fill-current" />
                        </Link>
                      </Button>
                    </div>
                    <Button
                      type="button"
                      size="icon"
                      variant="destructive"
                      className="absolute left-3 top-3 rounded-full opacity-0 shadow-lg transition-opacity group-hover:opacity-100"
                      disabled={deletingGameId === game.id}
                      aria-label={`删除 ${game.title}`}
                      title="删除游戏"
                      onClick={() => void handleDeleteGame(game.id, game.title)}
                    >
                      <Trash2 className="size-4" />
                    </Button>
                  </AspectRatio>

                  <CardHeader className="relative z-10 gap-1 bg-[#1a1d20] px-3 pb-3 pt-3">
                    <CardTitle className="truncate text-lg font-bold text-white">{game.title}</CardTitle>
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

async function readStoryFolderFiles(fileList: FileList) {
  const selectedFiles = Array.from(fileList)
  const relevantFiles = selectedFiles.filter((file) => isImportableStoryFile(file))
  const relativePaths = stripCommonRoot(relevantFiles.map(filePathOf))

  const entries = await Promise.all(
    relevantFiles.map(async (file, index) => {
      const key = relativePaths[index].toLowerCase()
      const content = await readFileContent(file)
      return [key, content] as const
    }),
  )
  return Object.fromEntries(entries)
}

function isImportableStoryFile(file: File) {
  const name = filePathOf(file).toLowerCase()
  return (
    name.endsWith('.md')
    || name.endsWith('.json')
    || name.endsWith('.png')
    || name.endsWith('.jpg')
    || name.endsWith('.jpeg')
    || name.endsWith('.webp')
    || name.endsWith('.gif')
  )
}

function filePathOf(file: File) {
  return (file.webkitRelativePath || file.name).replaceAll('\\', '/')
}

function stripCommonRoot(paths: string[]) {
  if (paths.length === 0) {
    return paths
  }

  const segments = paths.map((item) => item.split('/').filter(Boolean))
  const firstSegment = segments[0]?.[0]
  if (!firstSegment || segments.some((item) => item.length < 2 || item[0] !== firstSegment)) {
    return paths
  }

  return segments.map((item) => item.slice(1).join('/'))
}

function summarizeKeys(keys: string[]) {
  const sorted = [...keys].sort()
  const photoKeys = sorted.filter((key) => key.startsWith('photo/'))
  const visible = [...sorted.filter((key) => !key.startsWith('photo/')).slice(0, 8), ...photoKeys.slice(0, 8)]
  const suffix = sorted.length > visible.length ? ` ...(+${sorted.length - visible.length})` : ''
  return `${visible.join(',')}${suffix}`
}

async function readFileContent(file: File) {
  const lowerName = file.name.toLowerCase()
  const isText = lowerName.endsWith('.md') || lowerName.endsWith('.json')
  if (isText) {
    return file.text()
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(reader.error ?? new Error('read file failed'))
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.readAsDataURL(file)
  })
}

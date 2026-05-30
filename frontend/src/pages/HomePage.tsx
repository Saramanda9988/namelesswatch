import { Link, useNavigate } from '@tanstack/react-router'
import { FolderPlus, Gamepad2, History, Play, Search, Settings, Trash2 } from 'lucide-react'
import * as React from 'react'

import { SettingsDialog } from '@/components/settings-dialog'
import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useGameStore } from '@/stores/game-store'
import type { roleplay } from '../../wailsjs/go/models'
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
  const [isSettingsOpen, setIsSettingsOpen] = React.useState(false)

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
    <div className="min-h-screen bg-background text-foreground">
      <Input ref={inputRef} type="file" className="hidden" multiple onChange={handleImport} {...directoryInputProps} />

      <header className="sticky top-0 z-30 border-b bg-background/95 backdrop-blur">
        <div className="flex h-16 items-center justify-between gap-4 px-4 md:px-8 lg:px-12">
          <Link to="/" className="flex min-w-0 items-center gap-3">
            <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary text-primary-foreground">
              <Gamepad2 className="size-5" />
            </span>
            <span className="min-w-0">
              <span className="block truncate text-base font-semibold leading-none">NamelessWatch</span>
              <span className="mt-1 block truncate text-xs text-muted-foreground">AI Roleplay Library</span>
            </span>
          </Link>

          <div className="flex min-w-0 flex-1 items-center justify-end gap-2">
            <div className="relative hidden w-full max-w-sm sm:block">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="搜索游戏..."
                aria-label="搜索游戏"
                className="bg-muted/45 pl-9"
              />
            </div>
            <Button
              type="button"
              disabled={isImporting}
              onClick={() => inputRef.current?.click()}
            >
              <FolderPlus data-icon="inline-start" />
              {isImporting ? '导入中' : '添加游戏'}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              aria-label="设置"
              title="设置"
              onClick={() => setIsSettingsOpen(true)}
            >
              <Settings data-icon />
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto flex h-[calc(100vh-64px)] min-h-0 w-full max-w-[1760px] flex-col px-4 py-6 md:px-8 lg:px-12">
        <section className="flex flex-col gap-5">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div className="min-w-0">
              <h1 className="text-4xl font-bold tracking-normal md:text-5xl">游戏库</h1>
              <div className="mt-3 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                <span>共 {games.length} 个游戏</span>
                <Badge variant="secondary">{status}</Badge>
              </div>
            </div>

            <div className="relative sm:hidden">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="搜索游戏..."
                aria-label="搜索游戏"
                className="bg-muted/45 pl-9"
              />
            </div>
          </div>
        </section>

        <Separator className="my-5" />

        <ScrollArea className="min-h-0 flex-1 pr-3">
          {games.length === 0 ? (
            <Card className="grid min-h-[420px] place-items-center border-dashed bg-card/70 text-center">
              <CardContent className="flex max-w-md flex-col items-center gap-5 p-8">
                <span className="grid size-14 place-items-center rounded-lg bg-muted text-muted-foreground">
                  <Gamepad2 className="size-7" />
                </span>
                <div className="flex flex-col gap-2">
                  <CardTitle className="text-2xl">暂无游戏</CardTitle>
                  <CardDescription className="leading-6">
                    添加一个剧情包后，游戏会以网格卡片展示在这里。
                  </CardDescription>
                </div>
                <Button type="button" onClick={() => inputRef.current?.click()}>
                  <FolderPlus data-icon="inline-start" />
                  添加游戏
                </Button>
              </CardContent>
            </Card>
          ) : visibleGames.length === 0 ? (
            <Card className="grid min-h-[320px] place-items-center bg-card/70 text-center">
              <CardContent className="flex max-w-md flex-col items-center gap-3 p-8">
                <CardTitle className="text-xl">没有匹配的游戏</CardTitle>
                <CardDescription>换一个关键词试试。</CardDescription>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-4 pb-10 sm:grid-cols-[repeat(auto-fill,minmax(180px,1fr))] lg:grid-cols-[repeat(auto-fill,minmax(196px,1fr))] 2xl:grid-cols-[repeat(auto-fill,minmax(212px,1fr))]">
              {visibleGames.map((game) => (
                <Card
                  key={game.id}
                  size="sm"
                  className="group relative isolate gap-0 overflow-hidden py-0 shadow-sm transition-all duration-300 hover:-translate-y-0.5 hover:shadow-xl"
                >
                  <AspectRatio ratio={3 / 3.7} className="relative overflow-hidden bg-muted">
                    {game.photoUrls?.[0] ? (
                      <img
                        src={game.photoUrls[0]}
                        alt={game.title}
                        className="block size-full origin-center object-cover transition-transform duration-500 ease-out group-hover:scale-110"
                        draggable={false}
                      />
                    ) : (
                      <div className="grid size-full place-items-center text-muted-foreground">
                        <Gamepad2 className="size-12" />
                      </div>
                    )}

                    {continueByGame[game.id] ? (
                      <Badge variant="secondary" className="absolute left-2 top-2 bg-background/85 backdrop-blur">
                        可继续
                      </Badge>
                    ) : null}

                    <div className="absolute inset-0 flex items-center justify-center gap-2 bg-background/55 opacity-0 backdrop-blur-[2px] transition-opacity duration-300 group-hover:opacity-100">
                      {continueByGame[game.id] ? (
                        <Button
                          type="button"
                          size="icon-lg"
                          variant="secondary"
                          aria-label={`继续 ${game.title}`}
                          title="继续游戏（读取上次进度）"
                          onClick={() => handleContinue(game.id)}
                        >
                          <History data-icon />
                        </Button>
                      ) : null}
                      <Button asChild size="icon-lg" aria-label={`开始 ${game.title}`} title={continueByGame[game.id] ? '重新开始（开新游戏）' : '开始游戏'}>
                        <Link
                          to="/play/$gameId"
                          params={{ gameId: game.id }}
                          onClick={() => setPendingResumeSession(undefined)}
                        >
                          <Play data-icon />
                        </Link>
                      </Button>
                      <Button
                        type="button"
                        size="icon-lg"
                        variant="destructive"
                        disabled={deletingGameId === game.id}
                        aria-label={`删除 ${game.title}`}
                        title="删除游戏"
                        onClick={() => void handleDeleteGame(game.id, game.title)}
                      >
                        <Trash2 data-icon />
                      </Button>
                    </div>
                  </AspectRatio>

                  <CardHeader className="gap-1 px-3 py-3">
                    <CardTitle className="truncate text-base font-semibold" title={game.title}>
                      {game.title}
                    </CardTitle>
                    <CardDescription className="truncate text-xs">
                      {gameSubtitle(game)}
                    </CardDescription>
                  </CardHeader>
                </Card>
              ))}
            </div>
          )}
        </ScrollArea>
      </main>

      <SettingsDialog open={isSettingsOpen} onOpenChange={setIsSettingsOpen} />
    </div>
  )
}

function gameSubtitle(game: roleplay.LibraryGame) {
  const sceneCount = game.scenes?.length ?? 0
  const bgmCount = game.bgms?.length ?? 0
  if (sceneCount > 0 && bgmCount > 0) {
    return `${sceneCount} 个场景 · ${bgmCount} 首 BGM`
  }
  if (sceneCount > 0) {
    return `${sceneCount} 个场景`
  }
  return 'AI 文字冒险'
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
    || name.endsWith('.mp3')
    || name.endsWith('.ogg')
    || name.endsWith('.wav')
    || name.endsWith('.m4a')
    || name.endsWith('.webm')
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

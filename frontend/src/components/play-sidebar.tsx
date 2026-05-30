import { Bot, Map, MessageSquare, User, X } from 'lucide-react'
import * as React from 'react'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

export type PlaySidebarHistoryItem = {
  key: string
  role: 'ai' | 'user'
  text: string
}

type PlaySidebarProps = {
  gameTitle: string
  mapImage?: string
  activeSceneLabel: string
  historyItems: PlaySidebarHistoryItem[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

type SidebarContentProps = {
  gameTitle: string
  mapImage?: string
  activeSceneLabel: string
  historyItems: PlaySidebarHistoryItem[]
  showCloseButton?: boolean
  onClose?: () => void
}

export function PlaySidebar({
  gameTitle,
  mapImage,
  activeSceneLabel,
  historyItems,
  open,
  onOpenChange,
}: PlaySidebarProps) {
  React.useEffect(() => {
    if (!open) {
      return
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onOpenChange(false)
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [onOpenChange, open])

  return (
    <>
      {open ? (
        <div
          role="presentation"
          className="fixed inset-0 z-30 bg-background/60 backdrop-blur-sm duration-150 animate-in fade-in-0 lg:hidden"
          onClick={() => onOpenChange(false)}
        />
      ) : null}

      {open ? (
        <aside
          aria-label="游戏侧栏"
          className={cn(
            'fixed inset-y-0 right-0 z-40 flex min-h-0 w-[min(88vw,360px)] flex-col border-l bg-background shadow-2xl shadow-black/35 duration-200 animate-in slide-in-from-right',
            'lg:hidden',
          )}
        >
          <SidebarContent
            gameTitle={gameTitle}
            mapImage={mapImage}
            activeSceneLabel={activeSceneLabel}
            historyItems={historyItems}
            showCloseButton
            onClose={() => onOpenChange(false)}
          />
        </aside>
      ) : null}

      <aside aria-label="游戏侧栏" className="hidden min-h-0 flex-col border-l bg-background lg:flex">
        <SidebarContent
          gameTitle={gameTitle}
          mapImage={mapImage}
          activeSceneLabel={activeSceneLabel}
          historyItems={historyItems}
        />
      </aside>
    </>
  )
}

function SidebarContent({
  gameTitle,
  mapImage,
  activeSceneLabel,
  historyItems,
  showCloseButton,
  onClose,
}: SidebarContentProps) {
  const historyScrollRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    const node = historyScrollRef.current
    if (node) {
      node.scrollTop = node.scrollHeight
    }
  }, [historyItems.length])

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-col gap-3 px-4 pt-4 pb-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <div className="min-w-0">
              <h2 className="truncate text-lg font-semibold">地图简览</h2>
            </div>
          </div>
          {showCloseButton ? (
            <div className="flex shrink-0 items-center gap-1">
              <Button type="button" variant="ghost" size="icon-lg" aria-label="关闭侧栏" title="关闭侧栏" onClick={onClose}>
                <X data-icon />
              </Button>
            </div>
          ) : null}
        </div>

        <AspectRatio ratio={1.54} className="overflow-hidden rounded-md border bg-card">
          {mapImage ? (
            <img src={mapImage} alt="" className="size-full object-cover" />
          ) : (
            <MiniMapPlaceholder />
          )}
        </AspectRatio>
      </div>

      <Separator />

      <section className="flex min-h-0 flex-1 flex-col gap-3 px-4 pt-3 pb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold">聊天记录</h2>
        </div>
        <div className="rounded-md border bg-card/50 px-3 py-2 text-xs text-muted-foreground">
          当前场景：{activeSceneLabel}
        </div>
        <div ref={historyScrollRef} className="thin-scrollbar min-h-0 flex-1 overflow-y-auto pr-2">
          {historyItems.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              当前对话完整展示后，会逐条归档到这里。
            </p>
          ) : (
            <div className="flex flex-col gap-3 pb-2">
              {historyItems.map((item) => (
                item.role === 'user' ? (
                  <div key={item.key} className="flex flex-col items-end gap-1.5 pl-6">
                    <p className="whitespace-pre-line rounded-md rounded-tr-sm bg-primary px-3 py-2 text-sm leading-6 text-primary-foreground">
                      {item.text}
                    </p>
                  </div>
                ) : (
                  <div key={item.key} className="flex flex-col gap-1.5 pr-6">
                    <p className="whitespace-pre-line rounded-md rounded-tl-sm border bg-card/50 px-3 py-2 text-sm leading-6 text-foreground/90">
                      {item.text}
                    </p>
                  </div>
                )
              ))}
            </div>
          )}
        </div>
      </section>
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

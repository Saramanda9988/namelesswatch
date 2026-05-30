import { Check, Loader2, Pin } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import type { PlayerBriefing } from '@/lib/player-briefing'

type PlayerBriefingPanelProps = {
  briefing: PlayerBriefing
  isPreparing: boolean
  className?: string
  onConfirm: () => void
}

export function PlayerBriefingPanel({ briefing, isPreparing, className, onConfirm }: PlayerBriefingPanelProps) {
  return (
    <section
      aria-labelledby="player-briefing-title"
      className={cn(
        'relative mx-auto flex h-[min(58vh,31rem)] w-[min(42rem,calc(100vw-3rem))] -rotate-1 flex-col overflow-hidden rounded-[6px] border border-[#b69d42]/80 bg-[#e6cf68] text-[#231c10] shadow-[0_28px_70px_rgba(0,0,0,0.58),0_2px_0_rgba(255,255,255,0.38)_inset]',
        className,
      )}
    >
      <div className="absolute left-1/2 top-0 h-8 w-32 -translate-x-1/2 -translate-y-3 rotate-2 rounded-[2px] border border-white/35 bg-white/30 shadow-sm backdrop-blur-[1px]" aria-hidden="true" />
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(90deg,rgba(255,255,255,0.18),transparent_18%,rgba(89,61,0,0.08)_100%),repeating-linear-gradient(180deg,transparent_0,transparent_2.45rem,rgba(104,75,23,0.18)_2.5rem)]" aria-hidden="true" />

      <header className="relative flex items-start gap-3 px-7 pb-4 pt-7">
        <Pin className="mt-1 size-5 shrink-0 text-[#6d4b16]" aria-hidden="true" />
        <div className="min-w-0 flex-1">
          <h2 id="player-briefing-title" className="text-2xl font-semibold leading-8 text-[#231c10]">
            {briefing.title}
          </h2>
          {briefing.description ? (
            <p className="mt-1 text-sm leading-6 text-[#5d471e]">
              {briefing.description}
            </p>
          ) : null}
        </div>
      </header>

      <ScrollArea className="relative min-h-0 flex-1">
        <ol className="flex flex-col gap-1 px-7 pb-5 pt-1">
          {briefing.items.map((item, index) => (
            <li
              key={item.id}
              className="grid grid-cols-[2rem_minmax(0,1fr)] gap-3 border-b border-[#8f6f22]/25 py-3"
            >
              <span className="pt-0.5 font-mono text-lg leading-7 text-[#7a5a18]">
                {index + 1}
              </span>
              <span className="min-w-0 text-lg leading-7 text-[#231c10]">
                {item.text}
                {item.detail ? (
                  <span className="mt-1 block text-sm leading-6 text-[#684f20]">
                    {item.detail}
                  </span>
                ) : null}
              </span>
            </li>
          ))}
        </ol>
      </ScrollArea>

      <footer className="relative flex flex-col gap-3 border-t border-[#8f6f22]/30 bg-[#dfc45c]/60 px-7 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-h-6 items-center gap-2 text-sm text-[#5d471e]" aria-live="polite">
          {isPreparing ? <Loader2 className="size-4 animate-spin" aria-hidden="true" /> : <Check className="size-4" aria-hidden="true" />}
          <span>{isPreparing ? '正在生成开场叙事' : '开场叙事已准备好'}</span>
        </div>
        <Button type="button" className="w-full bg-[#1f1a13] text-[#f6e9b6] hover:bg-[#332716] sm:w-auto" onClick={onConfirm}>
          <Check data-icon="inline-start" />
          {briefing.confirmText}
        </Button>
      </footer>
    </section>
  )
}

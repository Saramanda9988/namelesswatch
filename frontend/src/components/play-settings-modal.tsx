import { Settings, Volume2, X } from 'lucide-react'
import * as React from 'react'

import { Button } from '@/components/ui/button'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import { useGameStore } from '@/stores/game-store'

type PlaySettingsModalProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function PlaySettingsModal({ open, onOpenChange }: PlaySettingsModalProps) {
  const voiceVolume = useGameStore((state) => state.settings.voiceVolume)
  const bgmEnabled = useGameStore((state) => state.settings.bgmEnabled)
  const bgmVolume = useGameStore((state) => state.settings.bgmVolume)
  const updateSettings = useGameStore((state) => state.updateSettings)

  const isBgmOn = bgmEnabled && bgmVolume > 0

  const handleBgmEnabledChange = React.useCallback((enabled: boolean) => {
    updateSettings({
      bgmEnabled: enabled,
      bgmVolume: enabled && bgmVolume <= 0 ? 64 : bgmVolume,
    })
  }, [bgmVolume, updateSettings])

  const handleBgmVolumeChange = React.useCallback((nextVolume: number) => {
    updateSettings({
      bgmEnabled: nextVolume > 0,
      bgmVolume: nextVolume,
    })
  }, [updateSettings])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        overlayClassName="bg-black/55 backdrop-blur-sm"
        className="dark w-[min(420px,calc(100vw-2rem))] gap-0 overflow-hidden rounded-xl border border-white/10 bg-[#9f9b9b]/90 p-0 text-white shadow-2xl ring-white/10 backdrop-blur-xl sm:max-w-md"
      >
        <DialogHeader className="relative border-b border-white/10 px-6 py-5 pr-14">
          <div className="flex items-start gap-3">
            <span className="grid size-10 shrink-0 place-items-center rounded-full bg-white/15 text-white shadow-inner">
              <Settings className="size-5" />
            </span>
            <div className="min-w-0">
              <DialogTitle className="text-lg font-semibold tracking-normal text-white">设置</DialogTitle>
              <DialogDescription className="mt-1 text-xs text-white/60">游玩时的声音偏好</DialogDescription>
            </div>
          </div>
          <DialogClose asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="关闭设置"
              title="关闭设置"
              className="absolute right-4 top-4 text-white/70 hover:bg-white/10 hover:text-white"
            >
              <X data-icon />
            </Button>
          </DialogClose>
        </DialogHeader>

        <div className="flex flex-col gap-5 px-6 py-6">
          <div className="flex items-center justify-between gap-4 rounded-lg bg-black/10 px-4 py-3">
            <div className="flex min-w-0 items-center gap-3">
              <span className="grid size-8 place-items-center rounded-full bg-white/10 text-white/85">
                <Volume2 className="size-4" />
              </span>
              <div className="min-w-0">
                <p className="text-sm font-medium leading-none text-white">背景音乐</p>
                <p className="mt-1 text-xs text-white/60">{isBgmOn ? '已开启' : '已静音'}</p>
              </div>
            </div>
            <Switch
              checked={isBgmOn}
              onCheckedChange={handleBgmEnabledChange}
              aria-label="背景音乐开关"
              className="data-checked:bg-white/70 data-unchecked:bg-black/25"
            />
          </div>

          <ModalVolumeSlider
            id="play-bgm-volume"
            label="BGM"
            value={bgmVolume}
            onChange={handleBgmVolumeChange}
          />
          <ModalVolumeSlider
            id="play-voice-volume"
            label="语音"
            value={voiceVolume}
            onChange={(nextVolume) => updateSettings({ voiceVolume: nextVolume })}
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}

type ModalVolumeSliderProps = {
  id: string
  label: string
  value: number
  onChange: (value: number) => void
}

function ModalVolumeSlider({ id, label, value, onChange }: ModalVolumeSliderProps) {
  return (
    <div className="grid grid-cols-[64px_minmax(0,1fr)_48px] items-center gap-4">
      <span id={`${id}-label`} className="text-base font-medium text-white">
        {label}
      </span>
      <Slider
        min={0}
        max={100}
        value={[value]}
        aria-labelledby={`${id}-label`}
        onValueChange={([nextValue]) => onChange(Math.round(nextValue))}
        className="h-6 [&_[data-slot=slider-range]]:bg-white/80 [&_[data-slot=slider-thumb]]:size-4 [&_[data-slot=slider-thumb]]:border-white/85 [&_[data-slot=slider-thumb]]:bg-white [&_[data-slot=slider-track]]:h-3 [&_[data-slot=slider-track]]:bg-white/25"
      />
      <span className="text-right text-sm tabular-nums text-white/70">{value}%</span>
    </div>
  )
}

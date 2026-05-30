import { Link } from '@tanstack/react-router'
import { ArrowLeft, Gauge, Map, MonitorCog, Volume2 } from 'lucide-react'
import type * as React from 'react'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import { useGameStore } from '@/stores/game-store'

export function SettingsPage() {
  const settings = useGameStore((state) => state.settings)
  const updateSettings = useGameStore((state) => state.updateSettings)

  return (
    <div className="min-h-screen bg-[#e9e1ca] text-[#1a1710]">
      <div className="mx-auto flex min-h-screen w-full max-w-5xl flex-col px-5 py-5 md:px-8">
        <header className="flex items-center justify-between pb-4">
          <Button asChild variant="ghost" className="rounded-none hover:bg-[#1a1710]/8">
            <Link to="/">
              <ArrowLeft className="size-4" />
              返回
            </Link>
          </Button>
          <span className="text-xs uppercase tracking-[0.36em] text-[#9b2d1b]">Settings</span>
        </header>
        <Separator className="bg-[#1a1710]/15" />

        <main className="grid flex-1 items-center gap-8 py-10 md:grid-cols-[0.8fr_1.2fr]">
          <section>
            <p className="w-fit border-y border-[#9b2d1b]/35 py-1 text-xs uppercase tracking-[0.38em] text-[#9b2d1b]">Control Desk</p>
            <h1 className="mt-5 text-5xl font-semibold leading-none md:text-6xl">设置</h1>
            <p className="mt-5 max-w-sm text-sm leading-7 text-[#1a1710]/64">这些选项会立刻影响游玩页的文本、地图与界面呈现。</p>
          </section>

          <Card className="rounded-none border-[#1a1710]/15 bg-[#f5edda]/70 text-[#1a1710]">
            <CardContent className="p-0">
              <SettingRange
                icon={<Gauge className="size-5" />}
                label="文本速度"
                value={settings.textSpeed}
                min={10}
                max={100}
                suffix="%"
                onChange={(textSpeed) => updateSettings({ textSpeed })}
              />
              <SettingRange
                icon={<Volume2 className="size-5" />}
                label="语音音量"
                value={settings.voiceVolume}
                min={0}
                max={100}
                suffix="%"
                onChange={(voiceVolume) => updateSettings({ voiceVolume })}
              />
              <SettingRange
                icon={<MonitorCog className="size-5" />}
                label="界面缩放"
                value={settings.uiScale}
                min={80}
                max={120}
                suffix="%"
                onChange={(uiScale) => updateSettings({ uiScale })}
              />
              <SettingToggle
                icon={<Map className="size-5" />}
                label="显示地图"
                checked={settings.showMap}
                onChange={(showMap) => updateSettings({ showMap })}
              />
              <SettingToggle
                icon={<Gauge className="size-5" />}
                label="自动推进"
                checked={settings.autoAdvance}
                onChange={(autoAdvance) => updateSettings({ autoAdvance })}
              />
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  )
}

type SettingRangeProps = {
  icon: React.ReactNode
  label: string
  value: number
  min: number
  max: number
  suffix: string
  onChange: (value: number) => void
}

function SettingRange({ icon, label, value, min, max, suffix, onChange }: SettingRangeProps) {
  return (
    <div className="grid gap-4 border-b border-[#1a1710]/10 p-5 md:grid-cols-[170px_1fr_72px] md:items-center">
      <Label className="flex items-center gap-3 text-sm font-medium">
        <span className="grid size-9 place-items-center border border-[#1a1710]/15 text-[#9b2d1b]">{icon}</span>
        {label}
      </Label>
      <Slider
        min={min}
        max={max}
        value={[value]}
        onValueChange={([nextValue]) => onChange(nextValue)}
        className="[&_[data-slot=slider-range]]:bg-[#9b2d1b] [&_[data-slot=slider-thumb]]:border-[#9b2d1b]"
      />
      <span className="text-right text-sm tabular-nums text-[#1a1710]/64">
        {value}
        {suffix}
      </span>
    </div>
  )
}

type SettingToggleProps = {
  icon: React.ReactNode
  label: string
  checked: boolean
  onChange: (checked: boolean) => void
}

function SettingToggle({ icon, label, checked, onChange }: SettingToggleProps) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-[#1a1710]/10 p-5 last:border-b-0">
      <Label className="flex items-center gap-3 text-sm font-medium">
        <span className="grid size-9 place-items-center border border-[#1a1710]/15 text-[#9b2d1b]">{icon}</span>
        {label}
      </Label>
      <Switch
        checked={checked}
        onCheckedChange={onChange}
        className="data-checked:bg-[#9b2d1b]"
      />
    </div>
  )
}

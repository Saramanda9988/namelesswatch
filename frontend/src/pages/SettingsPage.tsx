import { Link } from '@tanstack/react-router'
import { ArrowLeft, Bot, Cpu, Gauge, KeyRound, Link2, Map, MonitorCog, Save, Volume2 } from 'lucide-react'
import type * as React from 'react'
import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import { useGameStore } from '@/stores/game-store'
import type { AppConfig } from '@/types/game'

type SaveState = 'idle' | 'saving' | 'saved' | 'error'

export function SettingsPage() {
  const settings = useGameStore((state) => state.settings)
  const updateSettings = useGameStore((state) => state.updateSettings)
  const config = useGameStore((state) => state.config)
  const draftConfig = useGameStore((state) => state.draftConfig)
  const fetchConfig = useGameStore((state) => state.fetchConfig)
  const saveDraftConfig = useGameStore((state) => state.saveDraftConfig)
  const setDraftConfig = useGameStore((state) => state.setDraftConfig)
  const [isConfigLoading, setIsConfigLoading] = useState(true)
  const [saveState, setSaveState] = useState<SaveState>('idle')
  const [configError, setConfigError] = useState<string>()
  const didLoadConfigRef = useRef(false)

  useEffect(() => {
    let cancelled = false

    async function loadConfig() {
      setIsConfigLoading(true)
      setConfigError(undefined)
      try {
        await fetchConfig()
        if (!cancelled) {
          didLoadConfigRef.current = true
          setSaveState('saved')
        }
      }
      catch (cause) {
        if (!cancelled) {
          setConfigError(cause instanceof Error ? cause.message : String(cause))
          setSaveState('error')
        }
      }
      finally {
        if (!cancelled) {
          setIsConfigLoading(false)
        }
      }
    }

    void loadConfig()

    return () => {
      cancelled = true
    }
  }, [fetchConfig])

  useEffect(() => {
    if (!didLoadConfigRef.current || !config || !draftConfig) {
      return
    }
    if (JSON.stringify(config) === JSON.stringify(draftConfig)) {
      return
    }

    setSaveState('saving')
    const timer = window.setTimeout(() => {
      void saveDraftConfig()
        .then(() => {
          setSaveState('saved')
          setConfigError(undefined)
        })
        .catch((cause) => {
          setSaveState('error')
          setConfigError(cause instanceof Error ? cause.message : String(cause))
        })
    }, 300)

    return () => window.clearTimeout(timer)
  }, [config, draftConfig, saveDraftConfig])

  function patchConfig(patch: Partial<AppConfig>) {
    if (!draftConfig) {
      return
    }
    setDraftConfig({ ...draftConfig, ...patch })
  }

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

        <main className="grid flex-1 gap-8 py-10 md:grid-cols-[0.75fr_1.25fr]">
          <section className="self-start">
            <p className="w-fit border-y border-[#9b2d1b]/35 py-1 text-xs uppercase tracking-[0.38em] text-[#9b2d1b]">Control Desk</p>
            <h1 className="mt-5 text-5xl font-semibold leading-none md:text-6xl">设置</h1>
            <p className="mt-5 max-w-sm text-sm leading-7 text-[#1a1710]/64">模型配置会写入本机配置文件，游玩会话会读取最新保存值。</p>
            <div className="mt-6 flex items-center gap-2 text-xs text-[#1a1710]/60">
              <Save className="size-4 text-[#9b2d1b]" />
              <span>{saveLabel(saveState, isConfigLoading)}</span>
            </div>
            {configError ? <p className="mt-3 break-words text-xs leading-5 text-[#9b2d1b]">{configError}</p> : null}
          </section>

          <div className="flex flex-col gap-5">
            <Card className="rounded-none border-[#1a1710]/15 bg-[#f5edda]/70 text-[#1a1710]">
              <CardHeader className="border-b border-[#1a1710]/10">
                <CardTitle className="flex items-center gap-2 text-base">
                  <Bot className="size-4 text-[#9b2d1b]" />
                  AI 模型
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <SettingSelect
                  icon={<Cpu className="size-5" />}
                  label="Provider"
                  value={draftConfig?.ai_provider ?? 'openai'}
                  onChange={(ai_provider) => patchConfig({ ai_provider })}
                />
                <SettingInput
                  icon={<Link2 className="size-5" />}
                  label="Base URL"
                  value={draftConfig?.ai_base_url ?? ''}
                  placeholder="https://api.openai.com/v1"
                  disabled={isConfigLoading || !draftConfig}
                  onChange={(ai_base_url) => patchConfig({ ai_base_url })}
                />
                <SettingInput
                  icon={<Bot className="size-5" />}
                  label="模型"
                  value={draftConfig?.ai_model ?? ''}
                  placeholder="gpt-4o-mini"
                  disabled={isConfigLoading || !draftConfig}
                  onChange={(ai_model) => patchConfig({ ai_model })}
                />
                <SettingInput
                  icon={<KeyRound className="size-5" />}
                  label="Token"
                  type="password"
                  value={draftConfig?.ai_token ?? ''}
                  placeholder="sk-..."
                  disabled={isConfigLoading || !draftConfig}
                  onChange={(ai_token) => patchConfig({ ai_token })}
                />
              </CardContent>
            </Card>

            <Card className="rounded-none border-[#1a1710]/15 bg-[#f5edda]/70 text-[#1a1710]">
              <CardHeader className="border-b border-[#1a1710]/10">
                <CardTitle className="text-base">游玩偏好</CardTitle>
              </CardHeader>
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
          </div>
        </main>
      </div>
    </div>
  )
}

function saveLabel(saveState: SaveState, isLoading: boolean) {
  if (isLoading) {
    return '读取配置'
  }
  if (saveState === 'saving') {
    return '保存中'
  }
  if (saveState === 'error') {
    return '保存失败'
  }
  return '已保存'
}

type SettingInputProps = {
  icon: React.ReactNode
  label: string
  value: string
  placeholder?: string
  type?: React.HTMLInputTypeAttribute
  disabled?: boolean
  onChange: (value: string) => void
}

function SettingInput({ icon, label, value, placeholder, type = 'text', disabled, onChange }: SettingInputProps) {
  return (
    <div className="grid gap-4 border-b border-[#1a1710]/10 p-5 md:grid-cols-[170px_1fr] md:items-center">
      <Label className="flex items-center gap-3 text-sm font-medium">
        <span className="grid size-9 place-items-center border border-[#1a1710]/15 text-[#9b2d1b]">{icon}</span>
        {label}
      </Label>
      <Input
        type={type}
        value={value}
        placeholder={placeholder}
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        className="rounded-none border-[#1a1710]/15 bg-[#fff9e8]/70"
      />
    </div>
  )
}

type SettingSelectProps = {
  icon: React.ReactNode
  label: string
  value: string
  onChange: (value: string) => void
}

function SettingSelect({ icon, label, value, onChange }: SettingSelectProps) {
  return (
    <div className="grid gap-4 border-b border-[#1a1710]/10 p-5 md:grid-cols-[170px_1fr] md:items-center">
      <Label className="flex items-center gap-3 text-sm font-medium">
        <span className="grid size-9 place-items-center border border-[#1a1710]/15 text-[#9b2d1b]">{icon}</span>
        {label}
      </Label>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-8 w-full border border-[#1a1710]/15 bg-[#fff9e8]/70 px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
      >
        <option value="openai">OpenAI</option>
        <option value="deepseek">DeepSeek</option>
        <option value="custom">Custom</option>
      </select>
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

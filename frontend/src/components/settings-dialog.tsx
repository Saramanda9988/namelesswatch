import {
  Bot,
  Cpu,
  Gauge,
  KeyRound,
  Link2,
  Save,
  Settings,
  Sparkles,
  X,
} from 'lucide-react'
import * as React from 'react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import { useGameStore } from '@/stores/game-store'
import type { appconf } from '../../wailsjs/go/models'

type SaveState = 'idle' | 'saving' | 'saved' | 'error'
type SettingsCategoryId = 'model' | 'context' | 'prefetch' | 'play'

type SettingsDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const settingsCategories: Array<{
  id: SettingsCategoryId
  label: string
  description: string
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>
}> = [
  {
    id: 'model',
    label: 'AI 模型',
    description: 'Provider、模型和 Token',
    icon: Bot,
  },
  {
    id: 'context',
    label: '上下文',
    description: '回合保留与预算',
    icon: Gauge,
  },
  {
    id: 'prefetch',
    label: '生成加速',
    description: '选项预生成与并发',
    icon: Sparkles,
  },
  {
    id: 'play',
    label: '游玩偏好',
    description: '声音、缩放与地图',
    icon: Settings,
  },
]

export function SettingsDialog({ open, onOpenChange }: SettingsDialogProps) {
  const settings = useGameStore((state) => state.settings)
  const updateSettings = useGameStore((state) => state.updateSettings)
  const config = useGameStore((state) => state.config)
  const draftConfig = useGameStore((state) => state.draftConfig)
  const fetchConfig = useGameStore((state) => state.fetchConfig)
  const saveDraftConfig = useGameStore((state) => state.saveDraftConfig)
  const setDraftConfig = useGameStore((state) => state.setDraftConfig)
  const [activeCategoryId, setActiveCategoryId] = React.useState<SettingsCategoryId>('model')
  const [isConfigLoading, setIsConfigLoading] = React.useState(false)
  const [saveState, setSaveState] = React.useState<SaveState>('idle')
  const [configError, setConfigError] = React.useState<string>()
  const didLoadConfigRef = React.useRef(false)

  React.useEffect(() => {
    if (!open) {
      return
    }

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
  }, [fetchConfig, open])

  React.useEffect(() => {
    if (!open || !didLoadConfigRef.current || !config || !draftConfig) {
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
  }, [config, draftConfig, open, saveDraftConfig])

  const patchConfig = React.useCallback((patch: Partial<appconf.AppConfig>) => {
    const currentDraft = useGameStore.getState().draftConfig
    if (!currentDraft) {
      return
    }
    setDraftConfig({ ...currentDraft, ...patch })
  }, [setDraftConfig])

  const contextRecentTurns = draftConfig?.ai_context_recent_turns ?? 12
  const contextCompactTurns = draftConfig?.ai_context_compact_turns ?? 24
  const contextSoftBudget = draftConfig?.ai_context_soft_budget ?? 60000
  const contextHardBudget = draftConfig?.ai_context_hard_budget ?? 120000
  const prefetchGlobalConcurrency = draftConfig?.ai_choice_prefetch_global_concurrency ?? 2
  const prefetchSessionConcurrency = draftConfig?.ai_choice_prefetch_session_concurrency ?? 2
  const prefetchTTLSeconds = Math.round((draftConfig?.ai_choice_prefetch_ttl_ms ?? 120000) / 1000)
  const activeCategory = settingsCategories.find((category) => category.id === activeCategoryId) ?? settingsCategories[0]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        overlayClassName="bg-background/75 backdrop-blur-md"
        className="w-[min(1080px,calc(100vw-2rem))] max-w-none gap-0 overflow-hidden rounded-xl border bg-popover p-0 text-popover-foreground shadow-2xl ring-border sm:max-w-none"
      >
        <DialogHeader className="relative px-5 py-4 pr-14 md:px-6">
          <div className="flex items-start gap-3">
            <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary text-primary-foreground">
              <Settings className="size-5" />
            </span>
            <div className="min-w-0">
              <DialogTitle className="text-xl font-semibold">设置</DialogTitle>
            </div>
          </div>
          <DialogClose asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="关闭设置"
              title="关闭设置"
              className="absolute right-4 top-4"
            >
              <X data-icon />
            </Button>
          </DialogClose>
        </DialogHeader>

        <Separator />

        <div className="grid min-h-0 md:grid-cols-[260px_minmax(0,1fr)]">
          <aside className="flex min-h-0 flex-col gap-4 border-b bg-muted/35 p-3 md:h-[min(72vh,720px)] md:border-b-0 md:border-r">
            <nav className="grid grid-cols-2 gap-2 md:flex md:flex-col">
              {settingsCategories.map((category) => {
                const Icon = category.icon
                const isActive = category.id === activeCategoryId

                return (
                  <Button
                    key={category.id}
                    type="button"
                    variant={isActive ? 'secondary' : 'ghost'}
                    className={cn(
                      'h-auto justify-start whitespace-normal px-3 py-3 text-left',
                      isActive && 'bg-background shadow-sm hover:bg-background',
                    )}
                    aria-pressed={isActive}
                    onClick={() => setActiveCategoryId(category.id)}
                  >
                    <Icon data-icon="inline-start" />
                    <span className="min-w-0">
                      <span className="block truncate">{category.label}</span>
                      <span className="mt-0.5 hidden text-xs font-normal text-muted-foreground md:block">
                        {category.description}
                      </span>
                    </span>
                  </Button>
                )
              })}
            </nav>
          </aside>

          <ScrollArea className="h-[min(72vh,720px)] min-h-0">
            <div className="flex flex-col gap-4 p-4 md:p-6">
              {configError ? (
                <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive md:hidden">
                  {configError}
                </div>
              ) : null}

              {isConfigLoading && !draftConfig ? (
                <Card>
                  <CardHeader>
                    <CardTitle>正在读取配置</CardTitle>
                    <CardDescription>请稍候。</CardDescription>
                  </CardHeader>
                </Card>
              ) : (
                <>
                  {activeCategory.id === 'model' ? (
                    <ModelSettingsPanel
                      draftConfig={draftConfig}
                      disabled={isConfigLoading || !draftConfig}
                      onPatch={patchConfig}
                    />
                  ) : null}

                  {activeCategory.id === 'context' ? (
                    <ContextSettingsPanel
                      contextRecentTurns={contextRecentTurns}
                      contextCompactTurns={contextCompactTurns}
                      contextSoftBudget={contextSoftBudget}
                      contextHardBudget={contextHardBudget}
                      onPatch={patchConfig}
                    />
                  ) : null}

                  {activeCategory.id === 'prefetch' ? (
                    <PrefetchSettingsPanel
                      draftConfig={draftConfig}
                      prefetchGlobalConcurrency={prefetchGlobalConcurrency}
                      prefetchSessionConcurrency={prefetchSessionConcurrency}
                      prefetchTTLSeconds={prefetchTTLSeconds}
                      onPatch={patchConfig}
                    />
                  ) : null}

                  {activeCategory.id === 'play' ? (
                    <PlayPreferencePanel
                      settings={settings}
                      onUpdate={updateSettings}
                    />
                  ) : null}
                </>
              )}
            </div>
          </ScrollArea>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function ModelSettingsPanel({
  draftConfig,
  disabled,
  onPatch,
}: {
  draftConfig?: appconf.AppConfig
  disabled: boolean
  onPatch: (patch: Partial<appconf.AppConfig>) => void
}) {
  return (
    <SettingsCard
      icon={<Bot className="size-4" />}
      title="AI 模型"
      description="配置主持人使用的模型服务。"
    >
      <ProviderField
        value={draftConfig?.ai_provider ?? 'openai'}
        disabled={disabled}
        onChange={(ai_provider) => onPatch({ ai_provider })}
      />
      <SettingInput
        id="ai-base-url"
        label="Base URL"
        description="兼容 OpenAI API 的服务地址。"
        value={draftConfig?.ai_base_url ?? ''}
        placeholder="https://api.openai.com/v1"
        disabled={disabled}
        onChange={(ai_base_url) => onPatch({ ai_base_url })}
      >
        <Link2 className="size-4" />
      </SettingInput>
      <SettingInput
        id="ai-model"
        label="模型"
        description="用于叙事、选项和上下文整理的模型名称。"
        value={draftConfig?.ai_model ?? ''}
        placeholder="gpt-4o-mini"
        disabled={disabled}
        onChange={(ai_model) => onPatch({ ai_model })}
      >
        <Bot className="size-4" />
      </SettingInput>
      <SettingInput
        id="ai-token"
        label="Token"
        description="只保存在本机配置文件中。"
        type="password"
        value={draftConfig?.ai_token ?? ''}
        placeholder="sk-..."
        disabled={disabled}
        onChange={(ai_token) => onPatch({ ai_token })}
      >
        <KeyRound className="size-4" />
      </SettingInput>
    </SettingsCard>
  )
}

function ContextSettingsPanel({
  contextRecentTurns,
  contextCompactTurns,
  contextSoftBudget,
  contextHardBudget,
  onPatch,
}: {
  contextRecentTurns: number
  contextCompactTurns: number
  contextSoftBudget: number
  contextHardBudget: number
  onPatch: (patch: Partial<appconf.AppConfig>) => void
}) {
  return (
    <SettingsCard
      icon={<Gauge className="size-4" />}
      title="上下文管理"
      description="控制长线游玩时保留多少近期内容。"
    >
      <SettingRange
        id="context-recent-turns"
        label="最近回合"
        description="始终完整保留的最新回合数。"
        value={contextRecentTurns}
        min={4}
        max={24}
        suffix=" turn"
        onChange={(ai_context_recent_turns) =>
          onPatch({
            ai_context_recent_turns,
            ai_context_compact_turns: Math.max(contextCompactTurns, ai_context_recent_turns + 1),
          })}
      />
      <SettingRange
        id="context-compact-turns"
        label="压缩阈值"
        description="超过该回合数后进入摘要整理。"
        value={contextCompactTurns}
        min={Math.min(96, contextRecentTurns + 1)}
        max={96}
        suffix=" turn"
        onChange={(ai_context_compact_turns) => onPatch({ ai_context_compact_turns })}
      />
      <SettingRange
        id="context-soft-budget"
        label="软预算"
        description="达到该字数附近会优先触发压缩。"
        value={contextSoftBudget}
        min={12000}
        max={300000}
        suffix=" 字"
        onChange={(ai_context_soft_budget) =>
          onPatch({
            ai_context_soft_budget,
            ai_context_hard_budget: Math.max(contextHardBudget, ai_context_soft_budget),
          })}
      />
      <SettingRange
        id="context-hard-budget"
        label="硬预算"
        description="上下文拼装的最大保护线。"
        value={contextHardBudget}
        min={contextSoftBudget}
        max={500000}
        suffix=" 字"
        onChange={(ai_context_hard_budget) => onPatch({ ai_context_hard_budget })}
      />
    </SettingsCard>
  )
}

function PrefetchSettingsPanel({
  draftConfig,
  prefetchGlobalConcurrency,
  prefetchSessionConcurrency,
  prefetchTTLSeconds,
  onPatch,
}: {
  draftConfig?: appconf.AppConfig
  prefetchGlobalConcurrency: number
  prefetchSessionConcurrency: number
  prefetchTTLSeconds: number
  onPatch: (patch: Partial<appconf.AppConfig>) => void
}) {
  return (
    <SettingsCard
      icon={<Sparkles className="size-4" />}
      title="生成加速"
      description="在玩家阅读时预生成可选行动，降低等待时间。"
    >
      <SettingToggle
        label="选项预生成"
        description="开启后会提前请求候选分支。"
        checked={draftConfig?.ai_choice_prefetch_enabled ?? false}
        onChange={(ai_choice_prefetch_enabled) => onPatch({ ai_choice_prefetch_enabled })}
      />
      <SettingRange
        id="prefetch-global-concurrency"
        label="全局并发"
        description="所有会话共享的最大预生成任务数。"
        value={prefetchGlobalConcurrency}
        min={1}
        max={8}
        suffix=" 路"
        onChange={(ai_choice_prefetch_global_concurrency) =>
          onPatch({
            ai_choice_prefetch_global_concurrency,
            ai_choice_prefetch_session_concurrency: Math.min(prefetchSessionConcurrency, ai_choice_prefetch_global_concurrency),
          })}
      />
      <SettingRange
        id="prefetch-session-concurrency"
        label="单局并发"
        description="单个游玩会话允许同时预生成的任务数。"
        value={prefetchSessionConcurrency}
        min={1}
        max={prefetchGlobalConcurrency}
        suffix=" 路"
        onChange={(ai_choice_prefetch_session_concurrency) => onPatch({ ai_choice_prefetch_session_concurrency })}
      />
      <SettingRange
        id="prefetch-ttl"
        label="缓存时长"
        description="预生成结果保留多久。"
        value={prefetchTTLSeconds}
        min={10}
        max={600}
        suffix=" 秒"
        onChange={(seconds) => onPatch({ ai_choice_prefetch_ttl_ms: seconds * 1000 })}
      />
      <SettingRange
        id="prefetch-wait"
        label="提交等待"
        description="提交选择后等待预生成结果的最长时间。"
        value={draftConfig?.ai_choice_prefetch_wait_ms ?? 1200}
        min={0}
        max={10000}
        suffix=" ms"
        onChange={(ai_choice_prefetch_wait_ms) => onPatch({ ai_choice_prefetch_wait_ms })}
      />
    </SettingsCard>
  )
}

type PlaySettings = ReturnType<typeof useGameStore.getState>['settings']

function PlayPreferencePanel({
  settings,
  onUpdate,
}: {
  settings: PlaySettings
  onUpdate: (settings: Partial<PlaySettings>) => void
}) {
  return (
    <SettingsCard
      icon={<Settings className="size-4" />}
      title="游玩偏好"
      description="这些选项会即时影响正在进行的游玩体验。"
    >
      <SettingRange
        id="text-speed"
        label="文本速度"
        description="叙事逐字显示的速度。"
        value={settings.textSpeed}
        min={10}
        max={100}
        suffix="%"
        onChange={(textSpeed) => onUpdate({ textSpeed })}
      />
      <SettingRange
        id="voice-volume"
        label="语音音量"
        description="角色语音或音效音量。"
        value={settings.voiceVolume}
        min={0}
        max={100}
        suffix="%"
        onChange={(voiceVolume) => onUpdate({ voiceVolume })}
      />
      <SettingToggle
        label="背景音乐"
        description="控制场景 BGM 是否播放。"
        checked={settings.bgmEnabled}
        onChange={(bgmEnabled) => onUpdate({ bgmEnabled })}
      />
      <SettingRange
        id="bgm-volume"
        label="BGM 音量"
        description="背景音乐播放音量。"
        value={settings.bgmVolume}
        min={0}
        max={100}
        suffix="%"
        onChange={(bgmVolume) => onUpdate({ bgmVolume })}
      />
      <SettingRange
        id="ui-scale"
        label="界面缩放"
        description="调整阅读与控制界面大小。"
        value={settings.uiScale}
        min={80}
        max={120}
        suffix="%"
        onChange={(uiScale) => onUpdate({ uiScale })}
      />
      <SettingToggle
        label="显示地图"
        description="游玩页侧栏显示地图概览。"
        checked={settings.showMap}
        onChange={(showMap) => onUpdate({ showMap })}
      />
      <SettingToggle
        label="自动推进"
        description="文本显示完成后自动进入下一段。"
        checked={settings.autoAdvance}
        onChange={(autoAdvance) => onUpdate({ autoAdvance })}
      />
    </SettingsCard>
  )
}

function SettingsCard({
  icon,
  title,
  description,
  children,
}: {
  icon: React.ReactNode
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <Card className="gap-0 overflow-visible">
      <CardHeader className="border-b pb-4">
        <div className="flex items-start gap-3">
          <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-muted text-muted-foreground">
            {icon}
          </span>
          <div className="min-w-0">
            <CardTitle>{title}</CardTitle>
            <CardDescription className="mt-1">{description}</CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-4">
        <FieldGroup>{children}</FieldGroup>
      </CardContent>
    </Card>
  )
}

function ProviderField({
  value,
  disabled,
  onChange,
}: {
  value: string
  disabled: boolean
  onChange: (value: string) => void
}) {
  const providers = [
    { value: 'openai', label: 'OpenAI' },
    { value: 'deepseek', label: 'DeepSeek' },
    { value: 'custom', label: 'Custom' },
  ]

  return (
    <Field className="rounded-lg border bg-card/60 p-4">
      <div className="grid gap-3 md:grid-cols-[minmax(150px,0.48fr)_minmax(0,1fr)] md:items-center">
        <FieldContent>
          <FieldLabel>Provider</FieldLabel>
          <FieldDescription>选择模型服务商。</FieldDescription>
        </FieldContent>
        <div className="grid grid-cols-3 gap-2">
          {providers.map((provider) => (
            <Button
              key={provider.value}
              type="button"
              variant={value === provider.value ? 'default' : 'outline'}
              disabled={disabled}
              className="min-w-0"
              onClick={() => onChange(provider.value)}
            >
              <Cpu data-icon="inline-start" />
              <span className="truncate">{provider.label}</span>
            </Button>
          ))}
        </div>
      </div>
    </Field>
  )
}

function SettingInput({
  id,
  label,
  description,
  value,
  placeholder,
  type = 'text',
  disabled,
  onChange,
  children,
}: {
  id: string
  label: string
  description: string
  value: string
  placeholder?: string
  type?: React.HTMLInputTypeAttribute
  disabled?: boolean
  onChange: (value: string) => void
  children: React.ReactNode
}) {
  return (
    <Field className="rounded-lg border bg-card/60 p-4">
      <div className="grid gap-3 md:grid-cols-[minmax(150px,0.48fr)_minmax(0,1fr)] md:items-center">
        <FieldContent>
          <FieldLabel htmlFor={id} className="items-center">
            <span className="text-muted-foreground">{children}</span>
            {label}
          </FieldLabel>
          <FieldDescription>{description}</FieldDescription>
        </FieldContent>
        <Input
          id={id}
          type={type}
          value={value}
          placeholder={placeholder}
          disabled={disabled}
          onChange={(event) => onChange(event.target.value)}
          className="bg-background"
        />
      </div>
    </Field>
  )
}

function SettingRange({
  id,
  label,
  description,
  value,
  min,
  max,
  suffix,
  onChange,
}: {
  id: string
  label: string
  description: string
  value: number
  min: number
  max: number
  suffix: string
  onChange: (value: number) => void
}) {
  return (
    <Field className="rounded-lg border bg-card/60 p-4">
      <div className="grid gap-4 md:grid-cols-[minmax(150px,0.48fr)_minmax(0,1fr)_80px] md:items-center">
        <FieldContent>
          <FieldLabel htmlFor={id}>{label}</FieldLabel>
          <FieldDescription>{description}</FieldDescription>
        </FieldContent>
        <Slider
          id={id}
          min={min}
          max={max}
          step={1}
          value={[value]}
          aria-label={label}
          onValueChange={([nextValue]) => onChange(Math.round(nextValue))}
        />
        <span className="text-right text-sm tabular-nums text-muted-foreground">
          {value}
          {suffix}
        </span>
      </div>
    </Field>
  )
}

function SettingToggle({
  label,
  description,
  checked,
  onChange,
}: {
  label: string
  description: string
  checked: boolean
  onChange: (checked: boolean) => void
}) {
  return (
    <Field orientation="horizontal" className="items-center justify-between rounded-lg border bg-card/60 p-4">
      <FieldContent>
        <FieldTitle>{label}</FieldTitle>
        <FieldDescription>{description}</FieldDescription>
      </FieldContent>
      <Switch
        checked={checked}
        onCheckedChange={onChange}
        aria-label={label}
      />
    </Field>
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

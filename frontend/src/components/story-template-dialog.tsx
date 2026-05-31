import { FilePlus2, FolderOpen } from 'lucide-react'
import * as React from 'react'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useGameStore } from '@/stores/game-store'
import type { main } from '../../wailsjs/go/models'

type StoryTemplateDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (result: main.StoryTemplateResult) => void
}

export function StoryTemplateDialog({ open, onOpenChange, onCreated }: StoryTemplateDialogProps) {
  const selectStoryTemplateDirectory = useGameStore((state) => state.selectStoryTemplateDirectory)
  const createStoryTemplate = useGameStore((state) => state.createStoryTemplate)
  const [parentPath, setParentPath] = React.useState('')
  const [folderName, setFolderName] = React.useState('未命名规则怪谈')
  const [title, setTitle] = React.useState('')
  const [initialScene, setInitialScene] = React.useState('entrance')
  const [force, setForce] = React.useState(false)
  const [error, setError] = React.useState('')
  const [isSelecting, setIsSelecting] = React.useState(false)
  const [isCreating, setIsCreating] = React.useState(false)

  const isBusy = isSelecting || isCreating

  React.useEffect(() => {
    if (open) {
      setError('')
    }
  }, [open])

  async function handleSelectDirectory() {
    if (isBusy) {
      return
    }

    setIsSelecting(true)
    setError('')
    try {
      const selectedPath = await selectStoryTemplateDirectory()
      if (selectedPath) {
        setParentPath(selectedPath)
      }
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setIsSelecting(false)
    }
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isBusy) {
      return
    }
    if (!parentPath.trim()) {
      setError('请选择生成位置')
      return
    }
    if (!folderName.trim()) {
      setError('请输入模板文件夹名称')
      return
    }

    setIsCreating(true)
    setError('')
    try {
      const result = await createStoryTemplate(parentPath, folderName, title, initialScene, force)
      onCreated(result)
      onOpenChange(false)
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setIsCreating(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>新建剧情模板</DialogTitle>
          <DialogDescription>
            会在生成位置下创建一个模板文件夹，可编辑后通过添加游戏导入。
          </DialogDescription>
        </DialogHeader>

        <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="template-target">生成位置</FieldLabel>
              <div className="flex gap-2">
                <Input
                  id="template-target"
                  value={parentPath}
                  readOnly
                  placeholder="未选择"
                  className="font-mono text-xs"
                />
                <Button
                  type="button"
                  variant="outline"
                  disabled={isBusy}
                  aria-label="选择生成位置"
                  title="选择生成位置"
                  onClick={handleSelectDirectory}
                >
                  <FolderOpen data-icon="inline-start" />
                  {isSelecting ? '选择中' : '选择'}
                </Button>
              </div>
            </Field>

            <Field>
              <FieldLabel htmlFor="template-folder">模板文件夹</FieldLabel>
              <Input
                id="template-folder"
                value={folderName}
                disabled={isBusy}
                onChange={(event) => setFolderName(event.target.value)}
              />
            </Field>

            <Field>
              <FieldLabel htmlFor="template-title">标题</FieldLabel>
              <Input
                id="template-title"
                value={title}
                disabled={isBusy}
                placeholder="默认使用模板文件夹名称"
                onChange={(event) => setTitle(event.target.value)}
              />
            </Field>

            <Field>
              <FieldLabel htmlFor="template-initial-scene">初始场景 ID</FieldLabel>
              <Input
                id="template-initial-scene"
                value={initialScene}
                disabled={isBusy}
                placeholder="entrance"
                onChange={(event) => setInitialScene(event.target.value)}
              />
            </Field>

            <Field orientation="horizontal" className="items-center justify-between rounded-lg border bg-muted/30 p-3">
              <FieldContent>
                <FieldLabel htmlFor="template-force">覆盖已有模板文件</FieldLabel>
                <FieldDescription>仅覆盖模板脚手架中的同名文件。</FieldDescription>
              </FieldContent>
              <Switch
                id="template-force"
                checked={force}
                disabled={isBusy}
                aria-label="覆盖已有模板文件"
                onCheckedChange={setForce}
              />
            </Field>

            {error ? (
              <FieldError>{error}</FieldError>
            ) : null}
          </FieldGroup>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={isBusy}
              onClick={() => onOpenChange(false)}
            >
              取消
            </Button>
            <Button type="submit" disabled={isBusy || !parentPath.trim() || !folderName.trim()}>
              <FilePlus2 data-icon="inline-start" />
              {isCreating ? '创建中' : '创建模板'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

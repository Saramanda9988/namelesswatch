import type { ImportedGame, ImportReport, ScriptLine } from '@/types/game'

const REQUIRED_FILES = ['@endings.md', '@memory.md', '@rule.md', '@scene.md', '@ture.md'] as const
const IMAGE_EXTENSIONS = new Set(['.apng', '.avif', '.gif', '.jpg', '.jpeg', '.png', '.webp'])

function fileNameOf(file: File) {
  return (file.webkitRelativePath || file.name).replaceAll('\\', '/').split('/').at(-1) || file.name
}

function pathPartsOf(file: File) {
  return (file.webkitRelativePath || file.name).replaceAll('\\', '/').split('/').filter(Boolean)
}

function hasImageExtension(name: string) {
  const lowerName = name.toLowerCase()
  return Array.from(IMAGE_EXTENSIONS).some((extension) => lowerName.endsWith(extension))
}

function isInsideFolder(file: File, folderName: string) {
  return pathPartsOf(file).some((part) => part.toLowerCase() === folderName)
}

function titleFromFiles(files: File[]) {
  const firstPath = files.find((file) => file.webkitRelativePath)?.webkitRelativePath
  const rootName = firstPath?.replaceAll('\\', '/').split('/')[0]
  return rootName || `未命名剧本 ${new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
}

function parseScript(sceneText: string, photoUrls: string[]): ScriptLine[] {
  const blocks = sceneText
    .split(/\n\s*\n/g)
    .map((block) => block.trim())
    .filter(Boolean)

  const rawLines = blocks.length > 1 ? blocks : sceneText.split('\n').map((line) => line.trim()).filter(Boolean)

  return rawLines.map((line, index) => {
    const match = line.match(/^([^:：]{1,16})[:：]\s*(.+)$/)
    const speaker = match?.[1]?.trim() || (index % 4 === 0 ? 'AI' : '你')
    const text = match?.[2]?.trim() || line

    return {
      id: `line-${index + 1}`,
      speaker,
      text,
      backgroundUrl: photoUrls[index % Math.max(photoUrls.length, 1)],
    }
  })
}

export async function importGameFromFiles(fileList: FileList | File[]): Promise<ImportReport> {
  const files = Array.from(fileList)
  const warnings: string[] = []
  const markdownFiles = new Map<string, File>()

  for (const file of files) {
    const name = fileNameOf(file).toLowerCase()
    if (REQUIRED_FILES.includes(name as (typeof REQUIRED_FILES)[number])) {
      markdownFiles.set(name, file)
    }
  }

  const missing = REQUIRED_FILES.filter((fileName) => !markdownFiles.has(fileName))

  if (missing.length > 0) {
    return {
      missing,
      warnings,
    }
  }

  const entries = await Promise.all(
    Array.from(markdownFiles.entries()).map(async ([name, file]) => [name, await file.text()] as const),
  )
  const mdContents = Object.fromEntries(entries)

  const photoUrls = files
    .filter((file) => isInsideFolder(file, 'photo') && (file.type.startsWith('image/') || hasImageExtension(file.name)))
    .map((file) => URL.createObjectURL(file))

  const mapUrls = files
    .filter((file) => isInsideFolder(file, 'map') && (file.type.startsWith('image/') || hasImageExtension(file.name)))
    .map((file) => URL.createObjectURL(file))

  if (photoUrls.length === 0) {
    warnings.push('photo 目录未检测到图片，游玩页会使用占位背景。')
  }

  if (mapUrls.length === 0) {
    warnings.push('map 目录未检测到图片，右上角地图会使用占位图。')
  }

  const script = parseScript(mdContents['@scene.md'] || '', photoUrls)
  const titleLine = mdContents['@rule.md']?.split('\n').find((line) => line.trim().startsWith('#'))
  const title = titleLine?.replace(/^#+\s*/, '').trim() || titleFromFiles(files)
  const id = `${title}-${Date.now()}`.replace(/[^\w\u4e00-\u9fa5-]+/g, '-')

  const game: ImportedGame = {
    id,
    title,
    importedAt: new Date().toISOString(),
    files: mdContents,
    photoUrls,
    mapUrls,
    script,
  }

  return {
    game,
    missing,
    warnings,
  }
}

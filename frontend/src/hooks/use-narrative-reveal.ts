import * as React from 'react'

const FAST_MS_PER_CHAR = 12

export type RevealPhase = 'typing' | 'waiting' | 'done'

type RevealState = {
  activeIndex: number
  charCount: number
  phase: RevealPhase
}

export type NarrativeReveal = {
  revealedLines: string[]
  activeIndex: number
  phase: RevealPhase
  isComplete: boolean
  advance: () => void
}

type UseNarrativeRevealOptions = {
  lines: string[]
  resetKey: string
  textSpeed: number
  autoAdvance: boolean
}

function slowMsPerCharFromSpeed(textSpeed: number): number {
  const clamped = Math.min(100, Math.max(10, textSpeed))
  return Math.max(8, Math.round(85 - clamped * 0.7))
}

function initialState(lines: string[]): RevealState {
  if (lines.length === 0) {
    return { activeIndex: 0, charCount: 0, phase: 'done' }
  }
  return { activeIndex: 0, charCount: 0, phase: 'typing' }
}

export function useNarrativeReveal({ lines, resetKey, textSpeed, autoAdvance }: UseNarrativeRevealOptions): NarrativeReveal {
  const [state, setState] = React.useState<RevealState>(() => initialState(lines))

  const linesRef = React.useRef(lines)
  linesRef.current = lines

  React.useEffect(() => {
    setState(initialState(linesRef.current))
  }, [resetKey])

  const advance = React.useCallback(() => {
    setState((current) => {
      const currentLines = linesRef.current
      if (currentLines.length === 0) {
        return current
      }

      if (current.phase === 'typing') {
        const fullLength = currentLines[current.activeIndex]?.length ?? 0
        const isLast = current.activeIndex >= currentLines.length - 1
        return {
          ...current,
          charCount: fullLength,
          phase: isLast ? 'done' : 'waiting',
        }
      }

      if (current.phase === 'waiting') {
        const nextIndex = current.activeIndex + 1
        if (nextIndex >= currentLines.length) {
          return { ...current, phase: 'done' }
        }
        return { activeIndex: nextIndex, charCount: 0, phase: 'typing' }
      }

      return current
    })
  }, [])

  React.useEffect(() => {
    if (state.phase !== 'typing') {
      return
    }

    const currentLine = lines[state.activeIndex] ?? ''
    if (state.charCount >= currentLine.length) {
      const isLast = state.activeIndex >= lines.length - 1
      setState((current) => {
        if (current.phase !== 'typing') {
          return current
        }
        return { ...current, phase: isLast ? 'done' : 'waiting' }
      })
      return
    }

    const msPerChar = state.activeIndex === 0 ? FAST_MS_PER_CHAR : slowMsPerCharFromSpeed(textSpeed)
    const timer = window.setTimeout(() => {
      setState((current) => {
        if (current.phase !== 'typing' || current.activeIndex !== state.activeIndex) {
          return current
        }
        return { ...current, charCount: current.charCount + 1 }
      })
    }, msPerChar)

    return () => window.clearTimeout(timer)
  }, [state, lines, textSpeed])

  React.useEffect(() => {
    if (!autoAdvance || state.phase !== 'waiting') {
      return
    }
    const timer = window.setTimeout(() => {
      advance()
    }, 600)
    return () => window.clearTimeout(timer)
  }, [autoAdvance, state.phase, state.activeIndex, advance])

  const revealedLines = React.useMemo(() => {
    if (lines.length === 0) {
      return []
    }
    const result: string[] = []
    for (let index = 0; index < state.activeIndex && index < lines.length; index += 1) {
      result.push(lines[index])
    }
    const activeLine = lines[state.activeIndex]
    if (activeLine !== undefined) {
      result.push(activeLine.slice(0, state.charCount))
    }
    return result
  }, [lines, state.activeIndex, state.charCount])

  return {
    revealedLines,
    activeIndex: state.activeIndex,
    phase: state.phase,
    isComplete: state.phase === 'done',
    advance,
  }
}

import * as React from 'react'

type BGMTrack = {
  id: string
  url: string
}

type BGMPlayerOptions = {
  track?: BGMTrack
  enabled: boolean
  volume: number
}

type BGMPlayerState = {
  currentTrackId?: string
  isBlocked: boolean
  isPlaying: boolean
  retry: () => Promise<void>
}

const FADE_MS = 420

export function useBgmPlayer({ track, enabled, volume }: BGMPlayerOptions): BGMPlayerState {
  const audioRef = React.useRef<HTMLAudioElement | undefined>(undefined)
  const fadeTokenRef = React.useRef(0)
  const mountedRef = React.useRef(false)
  const [state, setState] = React.useState<Omit<BGMPlayerState, 'retry'>>({
    currentTrackId: undefined,
    isBlocked: false,
    isPlaying: false,
  })

  const normalizedVolume = clamp(volume / 100, 0, 1)
  const targetRef = React.useRef({ track, enabled, volume: normalizedVolume })
  targetRef.current = { track, enabled, volume: normalizedVolume }

  const safeSetState = React.useCallback((nextState: Omit<BGMPlayerState, 'retry'>) => {
    if (mountedRef.current) {
      setState(nextState)
    }
  }, [])

  const ensureAudio = React.useCallback(() => {
    if (!audioRef.current) {
      audioRef.current = new Audio()
      audioRef.current.loop = true
      audioRef.current.preload = 'auto'
    }
    return audioRef.current
  }, [])

  const stopPlayback = React.useCallback(async () => {
    const audio = audioRef.current
    if (!audio) {
      safeSetState({
        currentTrackId: undefined,
        isBlocked: false,
        isPlaying: false,
      })
      return
    }

    const token = nextFadeToken(fadeTokenRef)
    await fadeVolume(audio, 0, FADE_MS, token, fadeTokenRef)
    if (token !== fadeTokenRef.current) {
      return
    }
    audio.pause()
    safeSetState({
      currentTrackId: audio.dataset.trackId,
      isBlocked: false,
      isPlaying: false,
    })
  }, [safeSetState])

  const playTarget = React.useCallback(async () => {
    const target = targetRef.current
    if (!target.track?.url || !target.enabled || target.volume <= 0) {
      await stopPlayback()
      return
    }

    const audio = ensureAudio()
    const isNewTrack = audio.dataset.trackId !== target.track.id || audio.dataset.trackUrl !== target.track.url
    if (isNewTrack) {
      if (!audio.paused && audio.dataset.trackId) {
        const fadeOutToken = nextFadeToken(fadeTokenRef)
        await fadeVolume(audio, 0, FADE_MS, fadeOutToken, fadeTokenRef)
        if (fadeOutToken !== fadeTokenRef.current) {
          return
        }
      }
      audio.pause()
      audio.src = target.track.url
      audio.dataset.trackId = target.track.id
      audio.dataset.trackUrl = target.track.url
      audio.currentTime = 0
      audio.volume = 0
    }
    else if (audio.paused) {
      audio.volume = 0
    }

    audio.loop = true
    try {
      await audio.play()
      const fadeInToken = nextFadeToken(fadeTokenRef)
      await fadeVolume(audio, target.volume, FADE_MS, fadeInToken, fadeTokenRef)
      if (fadeInToken !== fadeTokenRef.current) {
        return
      }
      safeSetState({
        currentTrackId: target.track.id,
        isBlocked: false,
        isPlaying: true,
      })
    }
    catch {
      safeSetState({
        currentTrackId: target.track.id,
        isBlocked: true,
        isPlaying: false,
      })
    }
  }, [ensureAudio, safeSetState, stopPlayback])

  React.useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
      fadeTokenRef.current += 1
      const audio = audioRef.current
      if (audio) {
        audio.pause()
        audio.removeAttribute('src')
        audio.load()
      }
    }
  }, [])

  React.useEffect(() => {
    void playTarget()
  }, [enabled, normalizedVolume, playTarget, track?.id, track?.url])

  return {
    ...state,
    retry: playTarget,
  }
}

function nextFadeToken(tokenRef: React.MutableRefObject<number>) {
  tokenRef.current += 1
  return tokenRef.current
}

function fadeVolume(
  audio: HTMLAudioElement,
  targetVolume: number,
  duration: number,
  token: number,
  tokenRef: React.MutableRefObject<number>,
) {
  const initialVolume = audio.volume
  const startedAt = performance.now()

  return new Promise<void>((resolve) => {
    function step(now: number) {
      if (token !== tokenRef.current) {
        resolve()
        return
      }

      const progress = duration <= 0 ? 1 : clamp((now - startedAt) / duration, 0, 1)
      audio.volume = initialVolume + (targetVolume - initialVolume) * progress
      if (progress >= 1) {
        resolve()
        return
      }
      window.requestAnimationFrame(step)
    }

    window.requestAnimationFrame(step)
  })
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

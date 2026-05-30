import { create } from 'zustand'

type GreetingState = {
  name: string
  resultText: string
  setName: (name: string) => void
  setResultText: (resultText: string) => void
}

export const useGreetingStore = create<GreetingState>((set) => ({
  name: '',
  resultText: 'Enter a name to call the Wails Go backend.',
  setName: (name) => set({ name }),
  setResultText: (resultText) => set({ resultText }),
}))

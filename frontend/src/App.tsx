import { useMutation } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useGreetingStore } from '@/stores/greeting-store'
import { Greet } from '../wailsjs/go/main/App'

function App() {
  const { name, resultText, setName, setResultText } = useGreetingStore()
  const greeting = useMutation({
    mutationFn: Greet,
    onSuccess: setResultText,
    onError: (error) => {
      setResultText(error instanceof Error ? error.message : 'Failed to call the Wails backend.')
    },
  })

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!name.trim()) {
      setResultText('Please enter a name first.')
      return
    }
    greeting.mutate(name)
  }

  return (
    <section className="mx-auto flex min-h-screen w-full max-w-2xl flex-col justify-center px-6 py-10">
      <div className="space-y-6">
        <div className="space-y-2">
          <p className="text-sm font-medium text-muted-foreground">Wails v2 desktop app</p>
          <h1 className="text-3xl font-semibold tracking-normal">namelesswatch</h1>
          <p className="text-sm leading-6 text-muted-foreground">
            React, TypeScript, Zustand, TanStack Router, TanStack Query, shadcn/ui, Tailwind CSS, and pnpm are wired up.
          </p>
        </div>

        <form className="flex gap-2" onSubmit={handleSubmit}>
          <Input
            value={name}
            onChange={(event) => setName(event.target.value)}
            autoComplete="off"
            name="name"
            placeholder="Your name"
          />
          <Button type="submit" disabled={greeting.isPending}>
            {greeting.isPending && <Loader2 className="animate-spin" />}
            Greet
          </Button>
        </form>

        <div className="rounded-lg border bg-card px-4 py-3 text-sm text-card-foreground shadow-sm">
          {resultText}
        </div>
      </div>
    </section>
  )
}

export default App

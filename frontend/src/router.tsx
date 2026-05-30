import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router'

import { HomePage } from '@/pages/HomePage'
import { PlayPage } from '@/pages/PlayPage'

const rootRoute = createRootRoute({
  component: () => (
    <main className="dark min-h-screen bg-background text-foreground">
      <Outlet />
    </main>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const playRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/play/$gameId',
  component: PlayPage,
})

const routeTree = rootRoute.addChildren([indexRoute, playRoute])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

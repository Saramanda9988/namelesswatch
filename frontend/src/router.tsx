import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router'

import { HomePage } from '@/pages/HomePage'
import { PlayPage } from '@/pages/PlayPage'
import { SettingsPage } from '@/pages/SettingsPage'

const rootRoute = createRootRoute({
  component: () => (
    <main className="min-h-screen bg-[#10130f] text-foreground">
      <Outlet />
    </main>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
})

const playRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/play/$gameId',
  component: PlayPage,
})

const routeTree = rootRoute.addChildren([indexRoute, settingsRoute, playRoute])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

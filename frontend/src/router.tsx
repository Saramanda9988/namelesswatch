import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router'

import App from './App'

const rootRoute = createRootRoute({
  component: () => (
    <main className="min-h-screen bg-background text-foreground">
      <Outlet />
    </main>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: App,
})

const routeTree = rootRoute.addChildren([indexRoute])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

import { Navigate, createRootRoute, createRoute, createRouter } from '@tanstack/react-router'
import App from './App'
import { AlertsRoute, NodesRoute, OverviewRoute } from './features/routes/ConsoleRoutes'

const rootRoute = createRootRoute({
  component: App,
  notFoundComponent: () => <Navigate to="/overview" replace />,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => <Navigate to="/overview" replace />,
})

const overviewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/overview',
  component: OverviewRoute,
})

const nodesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/nodes',
  component: NodesRoute,
})

const alertsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/alerts',
  component: AlertsRoute,
})

const routeTree = rootRoute.addChildren([indexRoute, overviewRoute, nodesRoute, alertsRoute])

export const router = createRouter({ routeTree, defaultPreload: 'intent' })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

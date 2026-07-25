import { createContext, useContext } from 'react'
import type { Alert, AuthedRequest, Node } from '../types'

export type ConsoleContextValue = {
  nodes: Node[]
  alerts: Alert[]
  activeAlerts: Alert[]
  authedRequest: AuthedRequest
  refresh: () => Promise<void>
  selectNode: (node: Node) => void
}

export const ConsoleContext = createContext<ConsoleContextValue | null>(null)

export function useConsole() {
  const value = useContext(ConsoleContext)
  if (!value) throw new Error('useConsole must be used inside the authenticated console')
  return value
}

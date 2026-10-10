import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { Alert, Node } from '../../types'
import { OverviewView } from './OverviewView'

const node: Node = {
  id: 'node-one', name: 'Sydney API', hostname: 'api-syd-01', environment: 'production',
  display_address: null, agent_version: '0.1.0', os_name: 'linux', kernel_version: '6.8',
  last_seen_at: new Date().toISOString(), created_at: new Date().toISOString(), status: 'online',
  grafana_url: '/grafana/node-one',
}

const nodeName = (id: string | null) => (id === node.id ? node.name : id ?? 'Control plane')

describe('OverviewView', () => {
  it('renders fleet state and opens a node', () => {
    const select = vi.fn()
    render(<OverviewView nodes={[node]} alerts={[]} activeAlerts={[]} nodeName={nodeName} onNode={select} />)
    expect(screen.getByText('Sydney API')).toBeInTheDocument()
    expect(screen.getByText('No active incidents')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Sydney API'))
    expect(select).toHaveBeenCalledWith(node)
  })

  it('names the affected node on active incidents', () => {
    const alert: Alert = {
      id: 1, node_id: 'node-one', rule_name: 'Node offline', severity: 'critical', state: 'firing',
      summary: 'Node node-one has not reported for more than two minutes.',
      starts_at: new Date().toISOString(), ends_at: null, received_at: new Date().toISOString(),
    }
    render(<OverviewView nodes={[{ ...node, status: 'offline' }]} alerts={[alert]} activeAlerts={[alert]} nodeName={nodeName} onNode={vi.fn()} />)
    expect(screen.getByText('Node offline')).toBeInTheDocument()
    expect(screen.getByText(/Sydney API · started/)).toBeInTheDocument()
  })
})

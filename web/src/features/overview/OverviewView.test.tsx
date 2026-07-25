import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { Node } from '../../types'
import { OverviewView } from './OverviewView'

const node: Node = {
  id: 'node-one', name: 'Sydney API', hostname: 'api-syd-01', environment: 'production',
  display_address: null, agent_version: '0.1.0', os_name: 'linux', kernel_version: '6.8',
  last_seen_at: new Date().toISOString(), created_at: new Date().toISOString(), status: 'online',
  grafana_url: '/grafana/node-one',
}

describe('OverviewView', () => {
  it('renders fleet state and opens a node', () => {
    const select = vi.fn()
    render(<OverviewView nodes={[node]} alerts={[]} activeAlerts={[]} onNode={select} />)
    expect(screen.getByText('Sydney API')).toBeInTheDocument()
    expect(screen.getByText('No active incidents')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Sydney API'))
    expect(select).toHaveBeenCalledWith(node)
  })
})

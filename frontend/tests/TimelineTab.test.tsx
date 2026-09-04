import { describe, it, expect, vi, beforeEach, afterEach, afterAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { AuthProvider } from '@/lib/auth'
import { TimelineTab } from '@/components/TimelineTab'

const server = setupServer(
  http.get('http://localhost:8080/api/v1/cases/1/timeline', () =>
    HttpResponse.json([
      { evidence_id: 1, title: 'Event A', type: 'text', date: '2024-01-15T00:00:00Z', description: 'First event' },
      { evidence_id: 2, title: 'Event B', type: 'url', date: '2024-02-20T00:00:00Z', description: 'Second event' },
    ]),
  ),
)

beforeEach(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <AuthProvider>{ui}</AuthProvider>
    </QueryClientProvider>,
  )
}

describe('TimelineTab', () => {
  it('renders timeline events', async () => {
    renderWithProviders(<TimelineTab caseId={1} />)
    await waitFor(() => {
      expect(screen.getByText('Event A')).toBeInTheDocument()
      expect(screen.getByText('Event B')).toBeInTheDocument()
    })
  })
})

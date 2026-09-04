import { describe, it, expect, vi, beforeEach, afterEach, afterAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { AuthProvider } from '@/lib/auth'
import { ExportTab } from '@/components/ExportTab'

// Mock fetch for download tests.
global.fetch = vi.fn()

const server = setupServer()

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

describe('ExportTab', () => {
  it('renders export buttons', () => {
    renderWithProviders(<ExportTab caseId={1} />)
    expect(screen.getByText('Download PDF Report')).toBeInTheDocument()
    expect(screen.getByText('Download CSV')).toBeInTheDocument()
  })
})

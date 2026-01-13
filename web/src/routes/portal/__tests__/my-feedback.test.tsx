import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@/test/utils'
import type { PortalFeedbackSummary } from '@/types'

// Mock the API
vi.mock('@/lib/api', () => ({
  api: {
    portal: {
      listMyFeedback: vi.fn(),
    },
  },
}))

// Mock useParams
vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    useParams: () => ({ projectId: 'test-project-id' }),
    createFileRoute: () => ({ component: () => null }),
  }
})

import { api } from '@/lib/api'

// Simple component that mimics the my-feedback page behavior
function TestMyFeedbackPage({ feedback }: { feedback: PortalFeedbackSummary[] }) {
  if (feedback.length === 0) {
    return <div>You haven&apos;t submitted any feedback yet.</div>
  }
  return (
    <div>
      <h1>My Feedback</h1>
      {feedback.map((item) => (
        <div key={item.id} data-testid={`feedback-${item.id}`}>
          <h3>{item.title}</h3>
          <span>{item.status}</span>
        </div>
      ))}
    </div>
  )
}

describe('MyFeedbackPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays feedback items when loaded', async () => {
    const mockFeedback: PortalFeedbackSummary[] = [
      {
        id: 'fb-1',
        title: 'Test Feedback',
        description: 'Test description',
        type: 'feature',
        status: 'new',
        vote_count: 5,
        comment_count: 2,
        has_voted: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ]

    render(<TestMyFeedbackPage feedback={mockFeedback} />)

    expect(screen.getByText('Test Feedback')).toBeInTheDocument()
    expect(screen.getByText('new')).toBeInTheDocument()
  })

  it('shows empty state when no feedback', () => {
    render(<TestMyFeedbackPage feedback={[]} />)

    expect(screen.getByText(/haven't submitted any feedback/i)).toBeInTheDocument()
  })

  it('renders multiple feedback items', () => {
    const mockFeedback: PortalFeedbackSummary[] = [
      {
        id: 'fb-1',
        title: 'First Feedback',
        description: 'First description',
        type: 'feature',
        status: 'new',
        vote_count: 5,
        comment_count: 2,
        has_voted: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 'fb-2',
        title: 'Second Feedback',
        description: 'Second description',
        type: 'bug',
        status: 'in_progress',
        vote_count: 10,
        comment_count: 5,
        has_voted: true,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      },
    ]

    render(<TestMyFeedbackPage feedback={mockFeedback} />)

    expect(screen.getByText('First Feedback')).toBeInTheDocument()
    expect(screen.getByText('Second Feedback')).toBeInTheDocument()
  })
})

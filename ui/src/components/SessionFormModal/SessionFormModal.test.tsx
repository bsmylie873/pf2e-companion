import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import SessionFormModal from './SessionFormModal'

describe('SessionFormModal (create mode)', () => {
  it('should render "Create Session" heading', () => {
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByRole('heading', { name: 'Create Session' })).toBeInTheDocument()
  })

  it('should render title input', () => {
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByLabelText('Title *')).toBeInTheDocument()
  })

  it('should call onSave with correct data on submit', () => {
    const onSave = vi.fn()
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={false}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'The Dark Forest' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Session' }))
    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({
      title: 'The Dark Forest',
    }))
  })

  it('should not call onSave when title is empty', () => {
    const onSave = vi.fn()
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={false}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Create Session' }))
    expect(onSave).not.toHaveBeenCalled()
  })

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should display error message when error prop is provided', () => {
    render(
      <SessionFormModal
        mode="create"
        error="Something went wrong"
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('should disable submit button when saving', () => {
    render(
      <SessionFormModal
        mode="create"
        error={null}
        saving={true}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Saving…' })).toBeDisabled()
  })

  it('should show runtime error when end time is before start time', () => {
    render(
      <SessionFormModal
        mode="create"
        initial={{
          title: 'Test',
          session_number: null,
          scheduled_at: null,
          runtime_start: '2024-03-15T20:00:00.000Z',
          runtime_end: '2024-03-15T18:00:00.000Z',
        }}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('End time must be after start time.')).toBeInTheDocument()
  })
})

describe('SessionFormModal (edit mode)', () => {
  it('should render "Edit Session" heading', () => {
    render(
      <SessionFormModal
        mode="edit"
        initial={{
          title: 'The Dark Forest',
          session_number: 1,
          scheduled_at: null,
          runtime_start: null,
          runtime_end: null,
        }}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('Edit Session')).toBeInTheDocument()
  })

  it('should pre-fill title from initial data', () => {
    render(
      <SessionFormModal
        mode="edit"
        initial={{
          title: 'The Dark Forest',
          session_number: null,
          scheduled_at: null,
          runtime_start: null,
          runtime_end: null,
        }}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByDisplayValue('The Dark Forest')).toBeInTheDocument()
  })
})

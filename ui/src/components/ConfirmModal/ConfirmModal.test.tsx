import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import ConfirmModal from './ConfirmModal'

describe('ConfirmModal', () => {
  const defaultProps = {
    title: 'Delete Item',
    message: 'Are you sure you want to delete this item?',
    confirmLabel: 'Delete',
    error: null,
    loading: false,
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
  }

  it('should render the title and message', () => {
    render(<ConfirmModal {...defaultProps} />)
    expect(screen.getByText('Delete Item')).toBeInTheDocument()
    expect(screen.getByText('Are you sure you want to delete this item?')).toBeInTheDocument()
  })

  it('should render the confirm button with the provided label', () => {
    render(<ConfirmModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
  })

  it('should render a Cancel button', () => {
    render(<ConfirmModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
  })

  it('should call onConfirm when confirm button is clicked', () => {
    const onConfirm = vi.fn()
    render(<ConfirmModal {...defaultProps} onConfirm={onConfirm} />)
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('should call onCancel when cancel button is clicked', () => {
    const onCancel = vi.fn()
    render(<ConfirmModal {...defaultProps} onCancel={onCancel} />)
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('should call onCancel when the close button is clicked', () => {
    const onCancel = vi.fn()
    render(<ConfirmModal {...defaultProps} onCancel={onCancel} />)
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('should call onCancel when the backdrop is clicked', () => {
    const onCancel = vi.fn()
    render(<ConfirmModal {...defaultProps} onCancel={onCancel} />)
    fireEvent.click(document.querySelector('.modal-backdrop')!)
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('should disable the confirm button when loading is true', () => {
    render(<ConfirmModal {...defaultProps} loading={true} />)
    const confirmBtn = screen.getByRole('button', { name: /Deleting/i })
    expect(confirmBtn).toBeDisabled()
  })

  it('should show "Deleting…" text when loading is true', () => {
    render(<ConfirmModal {...defaultProps} loading={true} />)
    expect(screen.getByText('Deleting…')).toBeInTheDocument()
  })

  it('should display an error message when error is provided', () => {
    render(<ConfirmModal {...defaultProps} error="Something went wrong" />)
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('should not display an error message when error is null', () => {
    render(<ConfirmModal {...defaultProps} error={null} />)
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument()
  })
})

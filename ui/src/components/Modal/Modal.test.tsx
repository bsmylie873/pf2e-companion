import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import Modal from './Modal'

describe('Modal', () => {
  it('should render the title', () => {
    render(
      <Modal title="Test Modal" onClose={vi.fn()}>
        <p>Modal content</p>
      </Modal>
    )
    expect(screen.getByText('Test Modal')).toBeInTheDocument()
  })

  it('should render children', () => {
    render(
      <Modal title="Test Modal" onClose={vi.fn()}>
        <p>Modal content here</p>
      </Modal>
    )
    expect(screen.getByText('Modal content here')).toBeInTheDocument()
  })

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Test Modal" onClose={onClose}>
        <p>Content</p>
      </Modal>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should call onClose when backdrop is clicked', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Test Modal" onClose={onClose}>
        <p>Content</p>
      </Modal>
    )
    fireEvent.click(document.querySelector('.modal-backdrop')!)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should not call onClose when clicking inside modal card', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Test Modal" onClose={onClose}>
        <p>Content</p>
      </Modal>
    )
    fireEvent.click(document.querySelector('.modal-card')!)
    expect(onClose).not.toHaveBeenCalled()
  })

  it('should call onClose when Escape key is pressed', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Test Modal" onClose={onClose}>
        <p>Content</p>
      </Modal>
    )
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})

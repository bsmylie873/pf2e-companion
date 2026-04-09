import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import Pagination from './Pagination'

describe('Pagination', () => {
  it('should render page info', () => {
    render(<Pagination page={1} limit={10} total={50} onPageChange={vi.fn()} />)
    expect(screen.getByText(/Page 1 of 5/)).toBeInTheDocument()
    expect(screen.getByText(/50 total/)).toBeInTheDocument()
  })

  it('should disable Previous button on first page', () => {
    render(<Pagination page={1} limit={10} total={50} onPageChange={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled()
  })

  it('should disable Next button on last page', () => {
    render(<Pagination page={5} limit={10} total={50} onPageChange={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
  })

  it('should enable both buttons on a middle page', () => {
    render(<Pagination page={3} limit={10} total={50} onPageChange={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Previous' })).not.toBeDisabled()
    expect(screen.getByRole('button', { name: 'Next' })).not.toBeDisabled()
  })

  it('should call onPageChange with page - 1 when Previous is clicked', () => {
    const onPageChange = vi.fn()
    render(<Pagination page={3} limit={10} total={50} onPageChange={onPageChange} />)
    fireEvent.click(screen.getByRole('button', { name: 'Previous' }))
    expect(onPageChange).toHaveBeenCalledWith(2)
  })

  it('should call onPageChange with page + 1 when Next is clicked', () => {
    const onPageChange = vi.fn()
    render(<Pagination page={3} limit={10} total={50} onPageChange={onPageChange} />)
    fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(onPageChange).toHaveBeenCalledWith(4)
  })

  it('should show 1 total page when total is 0', () => {
    render(<Pagination page={1} limit={10} total={0} onPageChange={vi.fn()} />)
    expect(screen.getByText(/Page 1 of 1/)).toBeInTheDocument()
  })

  it('should calculate total pages from limit and total', () => {
    render(<Pagination page={1} limit={5} total={23} onPageChange={vi.fn()} />)
    expect(screen.getByText(/Page 1 of 5/)).toBeInTheDocument()
  })
})

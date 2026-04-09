import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import InlineNameInput from './InlineNameInput'

describe('InlineNameInput', () => {
  const onCommit = vi.fn()
  const onCancel = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders an input element', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('renders with placeholder prop', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} placeholder="Folder name" />)
    expect(screen.getByPlaceholderText('Folder name')).toBeInTheDocument()
  })

  it('renders input with the initial value', () => {
    render(<InlineNameInput value="My Folder" onCommit={onCommit} onCancel={onCancel} />)
    expect(screen.getByRole('textbox')).toHaveValue('My Folder')
  })

  it('has maxLength=100 attribute', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    expect(screen.getByRole('textbox')).toHaveAttribute('maxLength', '100')
  })

  it('Enter key with a new value calls onCommit with trimmed text', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: '  New Folder  ' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCommit).toHaveBeenCalledWith('New Folder')
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('Enter key with whitespace-only value calls onCancel, not onCommit', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: '   ' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCancel).toHaveBeenCalled()
    expect(onCommit).not.toHaveBeenCalled()
  })

  it('Enter key with same value as initial does NOT call onCommit or onCancel', () => {
    render(<InlineNameInput value="My Folder" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    // Value is already 'My Folder' (unchanged)
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCommit).not.toHaveBeenCalled()
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('Enter key with same value (after trimming) does NOT call onCommit', () => {
    render(<InlineNameInput value="My Folder" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: 'My Folder' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCommit).not.toHaveBeenCalled()
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('Escape key calls onCancel', () => {
    render(<InlineNameInput value="Some Folder" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(onCancel).toHaveBeenCalled()
    expect(onCommit).not.toHaveBeenCalled()
  })

  it('blur calls onCommit when value has changed', () => {
    render(<InlineNameInput value="Old Name" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: 'New Name' } })
    fireEvent.blur(input)
    expect(onCommit).toHaveBeenCalledWith('New Name')
  })

  it('blur with empty value calls onCancel', () => {
    render(<InlineNameInput value="Some Name" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: '' } })
    fireEvent.blur(input)
    expect(onCancel).toHaveBeenCalled()
    expect(onCommit).not.toHaveBeenCalled()
  })

  it('blur with whitespace-only value calls onCancel', () => {
    render(<InlineNameInput value="Some Name" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: '   ' } })
    fireEvent.blur(input)
    expect(onCancel).toHaveBeenCalled()
    expect(onCommit).not.toHaveBeenCalled()
  })

  it('blur with unchanged value does NOT call onCommit or onCancel', () => {
    render(<InlineNameInput value="My Folder" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.blur(input)
    expect(onCommit).not.toHaveBeenCalled()
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('shows error span when error prop is provided', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} error="Name is required" />)
    expect(screen.getByText('Name is required')).toBeInTheDocument()
  })

  it('does not render error span when error is null', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} error={null} />)
    expect(screen.queryByText(/error/i)).not.toBeInTheDocument()
  })

  it('does not render error span when error is undefined', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    // No span with error class should exist
    const wrap = document.querySelector('.folder-inline-wrap')
    expect(wrap?.querySelector('.folder-inline-error')).toBeNull()
  })

  it('other key presses do not trigger commit or cancel', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: 'abc' } })
    fireEvent.keyDown(input, { key: 'Tab' })
    fireEvent.keyDown(input, { key: 'a' })
    expect(onCommit).not.toHaveBeenCalled()
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('trims whitespace from value before committing', () => {
    render(<InlineNameInput value="" onCommit={onCommit} onCancel={onCancel} />)
    const input = screen.getByRole('textbox')
    fireEvent.change(input, { target: { value: '  Trimmed Name  ' } })
    fireEvent.blur(input)
    expect(onCommit).toHaveBeenCalledWith('Trimmed Name')
  })
})

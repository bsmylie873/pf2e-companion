import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import MapOverlayPanel from './MapOverlayPanel'

const baseProps = {
  gameTitle: 'Test Campaign',
  isGM: false,
  uploading: false,
  uploadError: null,
  displayScale: 1.0,
  panelOpen: false,
  unpinnedSessionsCount: 0,
  pinsCount: 0,
  transformRef: React.createRef() as React.RefObject<null>,
  fileInputRef: React.createRef() as React.RefObject<null>,
  onClose: vi.fn(),
  onOpen: vi.fn(),
  onUploadClick: vi.fn(),
  onDeleteMap: vi.fn(),
  onFileChange: vi.fn(),
}

describe('MapOverlayPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders toggle button when panel is closed', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={false} />)
    expect(screen.getByTitle('Show controls')).toBeInTheDocument()
  })

  it('calls onOpen when toggle button is clicked', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={false} />)
    fireEvent.click(screen.getByTitle('Show controls'))
    expect(baseProps.onOpen).toHaveBeenCalledTimes(1)
  })

  it('does not render panel content when closed', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={false} />)
    expect(screen.queryByText('Test Campaign')).not.toBeInTheDocument()
  })

  it('renders panel content when panelOpen is true', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} />)
    expect(screen.getByText('Test Campaign')).toBeInTheDocument()
  })

  it('does not render toggle button when panel is open', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} />)
    expect(screen.queryByTitle('Show controls')).not.toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} />)
    fireEvent.click(screen.getByTitle('Hide controls'))
    expect(baseProps.onClose).toHaveBeenCalledTimes(1)
  })

  it('shows GM controls when isGM is true', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} />)
    expect(screen.getByText('Replace Map')).toBeInTheDocument()
    expect(screen.getByText('Remove Map')).toBeInTheDocument()
  })

  it('hides GM controls when isGM is false', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={false} />)
    expect(screen.queryByText('Replace Map')).not.toBeInTheDocument()
    expect(screen.queryByText('Remove Map')).not.toBeInTheDocument()
  })

  it('calls onUploadClick when Replace Map button is clicked', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} />)
    fireEvent.click(screen.getByText('Replace Map'))
    expect(baseProps.onUploadClick).toHaveBeenCalledTimes(1)
  })

  it('calls onDeleteMap when Remove Map button is clicked', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} />)
    fireEvent.click(screen.getByText('Remove Map'))
    expect(baseProps.onDeleteMap).toHaveBeenCalledTimes(1)
  })

  it('shows uploading text when uploading is true', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} uploading={true} />)
    expect(screen.getByText('Uploading…')).toBeInTheDocument()
  })

  it('disables Replace Map button while uploading', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} uploading={true} />)
    const btn = screen.getByText('Uploading…')
    expect(btn).toBeDisabled()
  })

  it('shows uploadError when present', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} uploadError="Upload failed" />)
    expect(screen.getByText('Upload failed')).toBeInTheDocument()
  })

  it('does not show uploadError when null', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} isGM={true} uploadError={null} />)
    expect(screen.queryByText('Upload failed')).not.toBeInTheDocument()
  })

  it('shows displayScale as percentage', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} displayScale={1.5} />)
    expect(screen.getByText('150%')).toBeInTheDocument()
  })

  it('shows 100% at default displayScale', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} displayScale={1.0} />)
    expect(screen.getByText('100%')).toBeInTheDocument()
  })

  it('shows all-pinned message when all sessions are pinned and there are pins', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} unpinnedSessionsCount={0} pinsCount={3} />)
    expect(screen.getByText(/All sessions are pinned/)).toBeInTheDocument()
  })

  it('does not show all-pinned message when there are unpinned sessions', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} unpinnedSessionsCount={2} pinsCount={3} />)
    expect(screen.queryByText(/All sessions are pinned/)).not.toBeInTheDocument()
  })

  it('does not show all-pinned message when pinsCount is 0', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} unpinnedSessionsCount={0} pinsCount={0} />)
    expect(screen.queryByText(/All sessions are pinned/)).not.toBeInTheDocument()
  })

  it('zoom in button is disabled when displayScale is at max (5)', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} displayScale={5} />)
    expect(screen.getByTitle('Zoom in')).toBeDisabled()
  })

  it('zoom out button is disabled when displayScale is at min (0.75)', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} displayScale={0.75} />)
    expect(screen.getByTitle('Zoom out')).toBeDisabled()
  })

  it('zoom buttons are enabled at normal scale', () => {
    render(<MapOverlayPanel {...baseProps} panelOpen={true} displayScale={1.0} />)
    expect(screen.getByTitle('Zoom in')).not.toBeDisabled()
    expect(screen.getByTitle('Zoom out')).not.toBeDisabled()
  })
})

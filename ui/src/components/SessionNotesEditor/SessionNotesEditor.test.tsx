import React from 'react'
import { render, screen, fireEvent, within } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import SessionNotesEditor from './SessionNotesEditor'

// ---------------------------------------------------------------------------
// Hoist all mock primitives so they are accessible inside vi.mock() factories
// ---------------------------------------------------------------------------
const { mockUseEditor, mockEditor, chainProxy, mockUseAutosave } = vi.hoisted(() => {
  // Build a chainProxy where every method returns the proxy itself for chaining
  const cp: Record<string, ReturnType<typeof vi.fn>> = {}
  const chainMethods = [
    'focus', 'toggleBold', 'toggleItalic', 'toggleUnderline', 'toggleStrike',
    'toggleSubscript', 'toggleSuperscript', 'toggleHeading', 'toggleBulletList',
    'toggleOrderedList', 'toggleTaskList', 'toggleCodeBlock', 'toggleBlockquote',
    'toggleHighlight', 'setHorizontalRule', 'setTextAlign', 'setLineHeight',
    'unsetLineHeight', 'setFontSize', 'unsetFontSize', 'setColor', 'unsetColor',
    'insertTable', 'addRowAfter', 'addColumnAfter', 'deleteRow', 'deleteColumn',
    'setTextSelection', 'extendMarkRange', 'setLink', 'unsetLink', 'setImage', 'run',
  ]
  chainMethods.forEach(m => { cp[m] = vi.fn(() => cp) })

  const me = {
    isActive: vi.fn((_: unknown) => false),
    chain: vi.fn(() => cp),
    getAttributes: vi.fn((_: string) => ({} as Record<string, unknown>)),
    storage: {
      characterCount: {
        characters: vi.fn(() => 100),
        words: vi.fn(() => 20),
      },
    },
    state: {
      doc: {
        descendants: vi.fn((cb: (node: { type: { name: string }; attrs: { level: number }; textContent: string }, pos: number) => void) => {
          cb({ type: { name: 'heading' }, attrs: { level: 1 }, textContent: 'Test Heading' }, 0)
        }),
      },
    },
    view: {
      domAtPos: vi.fn(() => ({ node: { scrollIntoView: vi.fn() } })),
    },
    isDestroyed: false,
    isEmpty: false,
  }

  const mue = vi.fn(() => me)
  const mua = vi.fn(() => ({ scheduleAutosave: vi.fn(), status: 'idle' as string }))

  return { mockUseEditor: mue, mockEditor: me, chainProxy: cp, mockUseAutosave: mua }
})

// ---------------------------------------------------------------------------
// Module-level mocks (hoisted by Vitest before any imports run)
// ---------------------------------------------------------------------------
vi.mock('@tiptap/react', () => ({
  useEditor: mockUseEditor,
  EditorContent: () => <div data-testid="editor-content" />,
}))

vi.mock('@tiptap/react/menus', () => ({
  BubbleMenu: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="bubble-menu">{children}</div>
  ),
  FloatingMenu: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="floating-menu">{children}</div>
  ),
}))

vi.mock('@tiptap/starter-kit', () => ({ default: {} }))
vi.mock('@tiptap/extension-link', () => ({ default: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-underline', () => ({ default: {} }))
vi.mock('@tiptap/extension-table', () => ({ Table: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-table-row', () => ({ default: {} }))
vi.mock('@tiptap/extension-table-cell', () => ({ default: {} }))
vi.mock('@tiptap/extension-table-header', () => ({ default: {} }))
vi.mock('@tiptap/extension-highlight', () => ({ default: {} }))
vi.mock('@tiptap/extension-task-list', () => ({ default: {} }))
vi.mock('@tiptap/extension-task-item', () => ({ default: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-placeholder', () => ({ default: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-text-align', () => ({ default: { configure: () => ({}) } }))
vi.mock('@tiptap/extension-image', () => ({ default: {} }))
vi.mock('@tiptap/extension-typography', () => ({ default: {} }))
vi.mock('@tiptap/extension-character-count', () => ({ default: {} }))
vi.mock('@tiptap/extension-subscript', () => ({ default: {} }))
vi.mock('@tiptap/extension-superscript', () => ({ default: {} }))
vi.mock('@tiptap/extension-text-style', () => ({ TextStyle: {}, FontSize: {} }))
vi.mock('@tiptap/extension-color', () => ({ Color: {} }))
vi.mock('./extensions/blockLineHeight', () => ({
  BlockLineHeight: {},
  LINE_HEIGHT_PRESETS: [
    { value: '1.2' },
    { value: '1.5' },
    { value: '2.0' },
  ],
}))
vi.mock('../../hooks/useAutosave', () => ({
  useAutosave: mockUseAutosave,
}))

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------
function renderEditor(props: Partial<React.ComponentProps<typeof SessionNotesEditor>> = {}) {
  return render(<SessionNotesEditor initialContent={null} {...props} />)
}

// ToolbarBtn uses onMouseDown, NOT onClick, so we must use fireEvent.mouseDown
function mouseDown(btn: HTMLElement) {
  fireEvent.mouseDown(btn)
}

// ---------------------------------------------------------------------------
// beforeEach: clear all call counts, then re-wire implementations.
//
// The global setup.ts runs vi.restoreAllMocks() in afterEach which resets
// implementations but may not clear call counts on hoisted vi.fn() instances.
// vi.clearAllMocks() ensures counts are always zeroed before each test.
// ---------------------------------------------------------------------------
beforeEach(() => {
  // 1. Clear call counts for ALL vi.fn() instances (including hoisted ones)
  vi.clearAllMocks()

  // 2. Re-wire all implementations (restoreAllMocks in afterEach cleared them)

  // useEditor returns the mock editor by default
  mockUseEditor.mockReturnValue(mockEditor)

  // useAutosave returns idle status by default
  mockUseAutosave.mockReturnValue({ scheduleAutosave: vi.fn(), status: 'idle' })

  // isActive returns false for everything by default
  mockEditor.isActive.mockReturnValue(false)

  // chain returns the chainProxy
  mockEditor.chain.mockReturnValue(chainProxy)

  // getAttributes returns an empty object by default
  mockEditor.getAttributes.mockReturnValue({})

  // word count = 20
  mockEditor.storage.characterCount.words.mockReturnValue(20)

  // descendants yields one heading
  mockEditor.state.doc.descendants.mockImplementation((cb) => {
    cb({ type: { name: 'heading' }, attrs: { level: 1 }, textContent: 'Test Heading' }, 0)
  })

  // domAtPos returns a scrollable node
  mockEditor.view.domAtPos.mockReturnValue({ node: { scrollIntoView: vi.fn() } })

  // Re-wire every chainProxy method to return chainProxy for chaining
  Object.keys(chainProxy).forEach(key => {
    if (key !== 'run') {
      chainProxy[key].mockReturnValue(chainProxy)
    }
  })
})

// ===========================================================================
// Tests
// ===========================================================================
describe('SessionNotesEditor', () => {
  // -------------------------------------------------------------------------
  // Null editor guard
  // -------------------------------------------------------------------------
  describe('when useEditor returns null', () => {
    it('renders nothing', () => {
      mockUseEditor.mockReturnValue(null)
      const { container } = renderEditor()
      expect(container.firstChild).toBeNull()
    })
  })

  // -------------------------------------------------------------------------
  // Basic rendering
  // -------------------------------------------------------------------------
  describe('when editor is available', () => {
    it('renders the editor content area', () => {
      renderEditor()
      expect(screen.getByTestId('editor-content')).toBeInTheDocument()
    })

    it('renders the toolbar when editable=true (default)', () => {
      renderEditor()
      // Bold appears in the main toolbar and in the BubbleMenu; at least one is present
      expect(screen.getAllByRole('button', { name: 'Bold' }).length).toBeGreaterThan(0)
    })

    it('does NOT render the toolbar when editable=false', () => {
      renderEditor({ editable: false })
      // In editable=false mode the toolbar is absent; Bold only appears in BubbleMenu
      const boldBtns = screen.getAllByRole('button', { name: 'Bold' })
      // Only the bubble-menu Bold remains (no toolbar)
      expect(boldBtns).toHaveLength(1)
      expect(within(screen.getByTestId('bubble-menu')).getByRole('button', { name: 'Bold' })).toBeInTheDocument()
    })

    it('applies readonly CSS class when editable=false', () => {
      const { container } = renderEditor({ editable: false })
      expect(container.firstChild).toHaveClass('session-notes-editor--readonly')
    })

    it('does NOT apply readonly CSS class when editable=true', () => {
      const { container } = renderEditor()
      expect(container.firstChild).not.toHaveClass('session-notes-editor--readonly')
    })

    it('renders the BubbleMenu', () => {
      renderEditor()
      expect(screen.getByTestId('bubble-menu')).toBeInTheDocument()
    })

    it('renders the FloatingMenu', () => {
      renderEditor()
      expect(screen.getByTestId('floating-menu')).toBeInTheDocument()
    })

    it('displays the word count from editor.storage.characterCount', () => {
      renderEditor()
      expect(screen.getByText('20 words')).toBeInTheDocument()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — inline formatting
  // -------------------------------------------------------------------------
  describe('toolbar – inline formatting buttons', () => {
    it('clicking Bold fires toggleBold on the chain', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Bold' })[0])
      expect(chainProxy.toggleBold).toHaveBeenCalled()
    })

    it('clicking Italic fires toggleItalic', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Italic' })[0])
      expect(chainProxy.toggleItalic).toHaveBeenCalled()
    })

    it('clicking Underline fires toggleUnderline', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Underline' })[0])
      expect(chainProxy.toggleUnderline).toHaveBeenCalled()
    })

    it('clicking Strikethrough fires toggleStrike', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Strikethrough' })[0])
      expect(chainProxy.toggleStrike).toHaveBeenCalled()
    })

    it('clicking Subscript fires toggleSubscript', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Subscript' }))
      expect(chainProxy.toggleSubscript).toHaveBeenCalled()
    })

    it('clicking Superscript fires toggleSuperscript', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Superscript' }))
      expect(chainProxy.toggleSuperscript).toHaveBeenCalled()
    })

    it('clicking Highlight fires toggleHighlight', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Highlight' })[0])
      expect(chainProxy.toggleHighlight).toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — headings
  // -------------------------------------------------------------------------
  describe('toolbar – heading buttons', () => {
    it('clicking Heading 1 fires toggleHeading with level 1', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Heading 1' })[0])
      expect(chainProxy.toggleHeading).toHaveBeenCalledWith({ level: 1 })
    })

    it('clicking Heading 2 fires toggleHeading with level 2', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Heading 2' })[0])
      expect(chainProxy.toggleHeading).toHaveBeenCalledWith({ level: 2 })
    })

    it('clicking Heading 3 fires toggleHeading with level 3', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Heading 3' })[0])
      expect(chainProxy.toggleHeading).toHaveBeenCalledWith({ level: 3 })
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — list / block
  // -------------------------------------------------------------------------
  describe('toolbar – list and block buttons', () => {
    it('clicking Bullet list fires toggleBulletList', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Bullet list' })[0])
      expect(chainProxy.toggleBulletList).toHaveBeenCalled()
    })

    it('clicking Ordered list fires toggleOrderedList', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Ordered list' })[0])
      expect(chainProxy.toggleOrderedList).toHaveBeenCalled()
    })

    it('clicking Task list fires toggleTaskList', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Task list' }))
      expect(chainProxy.toggleTaskList).toHaveBeenCalled()
    })

    it('clicking Code block fires toggleCodeBlock', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Code block' }))
      expect(chainProxy.toggleCodeBlock).toHaveBeenCalled()
    })

    it('clicking Blockquote fires toggleBlockquote', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Blockquote' })[0])
      expect(chainProxy.toggleBlockquote).toHaveBeenCalled()
    })

    it('clicking Horizontal rule fires setHorizontalRule', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Horizontal rule' })[0])
      expect(chainProxy.setHorizontalRule).toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — text alignment
  // -------------------------------------------------------------------------
  describe('toolbar – alignment buttons', () => {
    it('clicking Align left fires setTextAlign("left")', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Align left' }))
      expect(chainProxy.setTextAlign).toHaveBeenCalledWith('left')
    })

    it('clicking Align center fires setTextAlign("center")', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Align center' }))
      expect(chainProxy.setTextAlign).toHaveBeenCalledWith('center')
    })

    it('clicking Align right fires setTextAlign("right")', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Align right' }))
      expect(chainProxy.setTextAlign).toHaveBeenCalledWith('right')
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — select dropdowns
  // -------------------------------------------------------------------------
  describe('toolbar – line height select', () => {
    it('selecting a line height value fires setLineHeight', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Line height' })
      fireEvent.change(select, { target: { value: '1.2' } })
      expect(chainProxy.setLineHeight).toHaveBeenCalledWith('1.2')
    })

    it('selecting empty line height fires unsetLineHeight', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Line height' })
      fireEvent.change(select, { target: { value: '' } })
      expect(chainProxy.unsetLineHeight).toHaveBeenCalled()
    })
  })

  describe('toolbar – font size select', () => {
    it('selecting a font size fires setFontSize', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Font size' })
      fireEvent.change(select, { target: { value: '0.85rem' } })
      expect(chainProxy.setFontSize).toHaveBeenCalledWith('0.85rem')
    })

    it('selecting empty font size fires unsetFontSize', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Font size' })
      fireEvent.change(select, { target: { value: '' } })
      expect(chainProxy.unsetFontSize).toHaveBeenCalled()
    })
  })

  describe('toolbar – text colour select', () => {
    it('selecting a colour fires setColor', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Text colour' })
      fireEvent.change(select, { target: { value: '#8b2e2e' } })
      expect(chainProxy.setColor).toHaveBeenCalledWith('#8b2e2e')
    })

    it('selecting empty colour fires unsetColor', () => {
      renderEditor()
      const select = screen.getByRole('combobox', { name: 'Text colour' })
      fireEvent.change(select, { target: { value: '' } })
      expect(chainProxy.unsetColor).toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — table
  // -------------------------------------------------------------------------
  describe('toolbar – table button', () => {
    it('clicking Insert table fires insertTable', () => {
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Insert table' })[0])
      expect(chainProxy.insertTable).toHaveBeenCalledWith({ rows: 3, cols: 3, withHeaderRow: true })
    })

    it('does NOT show table row/col controls when NOT in a table', () => {
      renderEditor()
      expect(screen.queryByRole('button', { name: 'Add row' })).not.toBeInTheDocument()
      expect(screen.queryByRole('button', { name: 'Add column' })).not.toBeInTheDocument()
    })

    it('shows table row/col controls when inside a table', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      expect(screen.getByRole('button', { name: 'Add row' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Add column' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Delete row' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Delete column' })).toBeInTheDocument()
    })

    it('clicking Add row fires addRowAfter', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Add row' }))
      expect(chainProxy.addRowAfter).toHaveBeenCalled()
    })

    it('clicking Add column fires addColumnAfter', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Add column' }))
      expect(chainProxy.addColumnAfter).toHaveBeenCalled()
    })

    it('clicking Delete row fires deleteRow', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Delete row' }))
      expect(chainProxy.deleteRow).toHaveBeenCalled()
    })

    it('clicking Delete column fires deleteColumn', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Delete column' }))
      expect(chainProxy.deleteColumn).toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — Link button
  // -------------------------------------------------------------------------
  describe('toolbar – link button', () => {
    it('clicking Link prompts for URL and calls setLink when URL entered', () => {
      vi.spyOn(window, 'prompt').mockReturnValue('https://example.com')
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Link' })[0])
      expect(window.prompt).toHaveBeenCalledWith('Enter URL:')
      expect(chainProxy.setLink).toHaveBeenCalledWith({ href: 'https://example.com' })
    })

    it('clicking Link with empty URL calls unsetLink', () => {
      vi.spyOn(window, 'prompt').mockReturnValue('')
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Link' })[0])
      expect(chainProxy.unsetLink).toHaveBeenCalled()
      expect(chainProxy.setLink).not.toHaveBeenCalled()
    })

    it('clicking Link when prompt is cancelled (null) does nothing', () => {
      vi.spyOn(window, 'prompt').mockReturnValue(null)
      renderEditor()
      mouseDown(screen.getAllByRole('button', { name: 'Link' })[0])
      expect(chainProxy.setLink).not.toHaveBeenCalled()
      expect(chainProxy.unsetLink).not.toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — Image button
  // -------------------------------------------------------------------------
  describe('toolbar – image button', () => {
    it('clicking Image prompts for src and alt, then calls setImage', () => {
      vi.spyOn(window, 'prompt')
        .mockReturnValueOnce('https://example.com/img.png')
        .mockReturnValueOnce('An image')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Image' }))
      expect(chainProxy.setImage).toHaveBeenCalledWith({ src: 'https://example.com/img.png', alt: 'An image' })
    })

    it('clicking Image when src prompt is cancelled does nothing', () => {
      vi.spyOn(window, 'prompt').mockReturnValue(null)
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Image' }))
      expect(chainProxy.setImage).not.toHaveBeenCalled()
    })

    it('clicking Image when src prompt is empty string does nothing', () => {
      vi.spyOn(window, 'prompt').mockReturnValue('')
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Image' }))
      expect(chainProxy.setImage).not.toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // Toolbar — Table of Contents
  // -------------------------------------------------------------------------
  describe('toolbar – Table of Contents toggle', () => {
    it('clicking § shows the TOC nav when headings exist', () => {
      renderEditor()
      expect(screen.queryByRole('navigation', { name: 'Table of Contents' })).not.toBeInTheDocument()
      mouseDown(screen.getByRole('button', { name: 'Table of Contents' }))
      expect(screen.getByRole('navigation', { name: 'Table of Contents' })).toBeInTheDocument()
      expect(screen.getByText('Test Heading')).toBeInTheDocument()
    })

    it('clicking § again hides the TOC nav', () => {
      renderEditor()
      const tocBtn = screen.getByRole('button', { name: 'Table of Contents' })
      mouseDown(tocBtn)
      expect(screen.getByRole('navigation', { name: 'Table of Contents' })).toBeInTheDocument()
      mouseDown(tocBtn)
      expect(screen.queryByRole('navigation', { name: 'Table of Contents' })).not.toBeInTheDocument()
    })

    it('does NOT show TOC nav when no headings exist', () => {
      mockEditor.state.doc.descendants.mockImplementation(() => {
        // no headings yielded
      })
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Table of Contents' }))
      expect(screen.queryByRole('navigation', { name: 'Table of Contents' })).not.toBeInTheDocument()
    })

    it('clicking a TOC heading fires setTextSelection and scrollIntoView', () => {
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Table of Contents' }))
      fireEvent.click(screen.getByText('Test Heading'))
      expect(chainProxy.setTextSelection).toHaveBeenCalledWith(1)
      const mockNode = mockEditor.view.domAtPos.mock.results[0]?.value?.node
      expect(mockNode?.scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'center' })
    })

    it('shows "(untitled)" for headings with empty textContent', () => {
      mockEditor.state.doc.descendants.mockImplementation((cb) => {
        cb({ type: { name: 'heading' }, attrs: { level: 2 }, textContent: '' }, 5)
      })
      renderEditor()
      mouseDown(screen.getByRole('button', { name: 'Table of Contents' }))
      expect(screen.getByText('(untitled)')).toBeInTheDocument()
    })
  })

  // -------------------------------------------------------------------------
  // Status labels
  // -------------------------------------------------------------------------
  describe('status labels', () => {
    it('shows nothing for idle status', () => {
      mockUseAutosave.mockReturnValue({ scheduleAutosave: vi.fn(), status: 'idle' })
      renderEditor()
      expect(screen.queryByText('Saving…')).not.toBeInTheDocument()
      expect(screen.queryByText(/Saved/)).not.toBeInTheDocument()
      expect(screen.queryByText('Save failed')).not.toBeInTheDocument()
    })

    it('shows "Saving…" for saving status', () => {
      mockUseAutosave.mockReturnValue({ scheduleAutosave: vi.fn(), status: 'saving' })
      renderEditor()
      expect(screen.getByText('Saving…')).toBeInTheDocument()
    })

    it('shows "Saved ✦" for saved status', () => {
      mockUseAutosave.mockReturnValue({ scheduleAutosave: vi.fn(), status: 'saved' })
      renderEditor()
      expect(screen.getByText('Saved ✦')).toBeInTheDocument()
    })

    it('shows "Save failed" for error status', () => {
      mockUseAutosave.mockReturnValue({ scheduleAutosave: vi.fn(), status: 'error' })
      renderEditor()
      expect(screen.getByText('Save failed')).toBeInTheDocument()
    })
  })

  // -------------------------------------------------------------------------
  // BubbleMenu buttons
  // -------------------------------------------------------------------------
  describe('BubbleMenu buttons', () => {
    it('BubbleMenu Bold fires toggleBold', () => {
      renderEditor()
      const bubbleMenu = screen.getByTestId('bubble-menu')
      mouseDown(within(bubbleMenu).getByRole('button', { name: 'Bold' }))
      expect(chainProxy.toggleBold).toHaveBeenCalled()
    })

    it('BubbleMenu Italic fires toggleItalic', () => {
      renderEditor()
      const bubbleMenu = screen.getByTestId('bubble-menu')
      mouseDown(within(bubbleMenu).getByRole('button', { name: 'Italic' }))
      expect(chainProxy.toggleItalic).toHaveBeenCalled()
    })

    it('BubbleMenu Link fires handleLinkClick and calls setLink', () => {
      vi.spyOn(window, 'prompt').mockReturnValue('https://bubble.example.com')
      renderEditor()
      const bubbleMenu = screen.getByTestId('bubble-menu')
      mouseDown(within(bubbleMenu).getByRole('button', { name: 'Link' }))
      expect(chainProxy.setLink).toHaveBeenCalledWith({ href: 'https://bubble.example.com' })
    })

    it('BubbleMenu Highlight fires toggleHighlight', () => {
      renderEditor()
      const bubbleMenu = screen.getByTestId('bubble-menu')
      mouseDown(within(bubbleMenu).getByRole('button', { name: 'Highlight' }))
      expect(chainProxy.toggleHighlight).toHaveBeenCalled()
    })
  })

  // -------------------------------------------------------------------------
  // FloatingMenu buttons
  // -------------------------------------------------------------------------
  describe('FloatingMenu buttons', () => {
    it('FloatingMenu Heading 1 fires toggleHeading', () => {
      renderEditor()
      const floatingMenu = screen.getByTestId('floating-menu')
      mouseDown(within(floatingMenu).getByRole('button', { name: 'Heading 1' }))
      expect(chainProxy.toggleHeading).toHaveBeenCalledWith({ level: 1 })
    })

    it('FloatingMenu Insert table fires insertTable', () => {
      renderEditor()
      const floatingMenu = screen.getByTestId('floating-menu')
      mouseDown(within(floatingMenu).getByRole('button', { name: 'Insert table' }))
      expect(chainProxy.insertTable).toHaveBeenCalledWith({ rows: 3, cols: 3, withHeaderRow: true })
    })
  })

  // -------------------------------------------------------------------------
  // Active button state (aria-pressed)
  // -------------------------------------------------------------------------
  describe('active button state', () => {
    it('Bold button has aria-pressed=true when bold is active', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'bold')
      renderEditor()
      const boldBtns = screen.getAllByRole('button', { name: 'Bold' })
      expect(boldBtns[0]).toHaveAttribute('aria-pressed', 'true')
    })

    it('Bold button has aria-pressed=false when bold is not active', () => {
      renderEditor()
      const boldBtns = screen.getAllByRole('button', { name: 'Bold' })
      expect(boldBtns[0]).toHaveAttribute('aria-pressed', 'false')
    })

    it('Insert table button has aria-pressed=true when in a table', () => {
      mockEditor.isActive.mockImplementation((name: unknown) => name === 'table')
      renderEditor()
      const tableBtn = screen.getAllByRole('button', { name: 'Insert table' })[0]
      expect(tableBtn).toHaveAttribute('aria-pressed', 'true')
    })
  })
})

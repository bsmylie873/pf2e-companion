import { useState } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import { BubbleMenu, FloatingMenu } from '@tiptap/react/menus'
import StarterKit from '@tiptap/starter-kit'
import { Table } from '@tiptap/extension-table'
import TableRow from '@tiptap/extension-table-row'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import Highlight from '@tiptap/extension-highlight'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Placeholder from '@tiptap/extension-placeholder'
import TextAlign from '@tiptap/extension-text-align'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Underline from '@tiptap/extension-underline'
import Typography from '@tiptap/extension-typography'
import CharacterCount from '@tiptap/extension-character-count'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import { TextStyle, FontSize } from '@tiptap/extension-text-style'
import { Color } from '@tiptap/extension-color'
import { BlockLineHeight, LINE_HEIGHT_PRESETS, type LineHeightValue } from './extensions/blockLineHeight'
import type { JSONContent } from '@tiptap/react'
import { useAutosave } from '../../hooks/useAutosave'
import './SessionNotesEditor.css'

interface ToolbarBtnProps {
  label: string
  active: boolean
  onClick: () => void
  ariaLabel?: string
}

function ToolbarBtn({ label, active, onClick, ariaLabel }: ToolbarBtnProps) {
  return (
    <button
      className={`sne-toolbar-btn${active ? ' sne-toolbar-btn--active' : ''}`}
      onMouseDown={(e) => { e.preventDefault(); onClick() }}
      title={ariaLabel ?? label}
      aria-label={ariaLabel ?? label}
      aria-pressed={active}
    >
      {label}
    </button>
  )
}

interface SessionNotesEditorProps {
  initialContent: JSONContent | null
  editable?: boolean
  onSave?: (content: JSONContent) => Promise<unknown>
  placeholder?: string
}

const IMAGE_URL_RE = /^https?:\/\/\S+\.(?:png|jpe?g|gif|webp|svg|bmp|ico|avif)(?:\?[^\s]*)?$/i

function isImageUrl(text: string): boolean {
  return IMAGE_URL_RE.test(text.trim())
}

const STATUS_LABELS: Record<string, string> = {
  idle: '',
  saving: 'Saving…',
  saved: 'Saved ✦',
  error: 'Save failed',
}

const FONT_SIZE_PRESETS = [
  { label: 'Small', value: '0.85rem' },
  { label: 'Normal', value: '1.05rem' },
  { label: 'Large', value: '1.3rem' },
  { label: 'X-Large', value: '1.6rem' },
] as const

const COLOR_PRESETS = [
  { label: 'Default', value: '' },
  { label: 'Crimson', value: '#8b2e2e' },
  { label: 'Forest', value: '#2e5e2e' },
  { label: 'Royal', value: '#2e3e8b' },
  { label: 'Gold', value: '#8b7a2e' },
  { label: 'Purple', value: '#5e2e8b' },
  { label: 'Teal', value: '#2e7a7a' },
] as const

export default function SessionNotesEditor({
  initialContent,
  editable = true,
  onSave,
  placeholder = 'Begin your session chronicle…',
}: SessionNotesEditorProps) {
  const { scheduleAutosave, status } = useAutosave(
    onSave ?? (() => Promise.reject(new Error('No onSave provided'))),
  )

  const [tocOpen, setTocOpen] = useState(false)

  const editor = useEditor({
    editable,
    extensions: [
      StarterKit,
      Link.configure({ openOnClick: false, autolink: true }),
      Underline,
      Table.configure({ resizable: true }),
      TableRow,
      TableCell,
      TableHeader,
      Highlight,
      TaskList,
      TaskItem.configure({ nested: false }),
      Placeholder.configure({ placeholder }),
      TextAlign.configure({ types: ['heading', 'paragraph'] }),
      Image,
      BlockLineHeight,
      Typography,
      CharacterCount,
      Subscript,
      Superscript,
      TextStyle,
      FontSize,
      Color,
    ],
    content: initialContent ?? '',
    onUpdate: editable
      ? ({ editor: ed }) => {
          if (onSave) scheduleAutosave(ed.getJSON())
        }
      : undefined,
    editorProps: {
      handlePaste: (view, event) => {
        const text = event.clipboardData?.getData('text/plain') ?? ''
        if (isImageUrl(text)) {
          const { state, dispatch } = view
          const node = state.schema.nodes.image.create({ src: text.trim() })
          const tr = state.tr.replaceSelectionWith(node)
          dispatch(tr)
          return true
        }
        return false
      },
    },
  })

  if (!editor) return null

  const headings: Array<{ level: number; text: string; pos: number }> = []
  editor.state.doc.descendants((node, pos) => {
    if (node.type.name === 'heading') {
      headings.push({ level: node.attrs.level as number, text: node.textContent, pos })
    }
  })

  const handleLinkClick = () => {
    const url = window.prompt('Enter URL:')
    if (url === null) return
    if (url.trim() === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }
    editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  }

  const handleImageClick = () => {
    const src = window.prompt('Enter image URL:')
    if (!src) return
    const alt = window.prompt('Alt text:') ?? ''
    editor.chain().focus().setImage({ src, alt }).run()
  }

  const inTable = editor.isActive('table')

  return (
    <div className={`session-notes-editor${!editable ? ' session-notes-editor--readonly' : ''}`}>
      {editable && (
        <div className="sne-toolbar">
          {/* Inline format */}
          <ToolbarBtn label="B" ariaLabel="Bold" active={editor.isActive('bold')} onClick={() => editor.chain().focus().toggleBold().run()} />
          <ToolbarBtn label="I" ariaLabel="Italic" active={editor.isActive('italic')} onClick={() => editor.chain().focus().toggleItalic().run()} />
          <ToolbarBtn label="U" ariaLabel="Underline" active={editor.isActive('underline')} onClick={() => editor.chain().focus().toggleUnderline().run()} />
          <ToolbarBtn label="S̶" ariaLabel="Strikethrough" active={editor.isActive('strike')} onClick={() => editor.chain().focus().toggleStrike().run()} />
          <ToolbarBtn label="X₂" ariaLabel="Subscript" active={editor.isActive('subscript')} onClick={() => editor.chain().focus().toggleSubscript().run()} />
          <ToolbarBtn label="X²" ariaLabel="Superscript" active={editor.isActive('superscript')} onClick={() => editor.chain().focus().toggleSuperscript().run()} />

          <span className="sne-toolbar-divider" />

          {/* Headings */}
          <ToolbarBtn label="H1" ariaLabel="Heading 1" active={editor.isActive('heading', { level: 1 })} onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} />
          <ToolbarBtn label="H2" ariaLabel="Heading 2" active={editor.isActive('heading', { level: 2 })} onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} />
          <ToolbarBtn label="H3" ariaLabel="Heading 3" active={editor.isActive('heading', { level: 3 })} onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} />

          <span className="sne-toolbar-divider" />

          {/* Lists */}
          <ToolbarBtn label="• List" ariaLabel="Bullet list" active={editor.isActive('bulletList')} onClick={() => editor.chain().focus().toggleBulletList().run()} />
          <ToolbarBtn label="1. List" ariaLabel="Ordered list" active={editor.isActive('orderedList')} onClick={() => editor.chain().focus().toggleOrderedList().run()} />
          <ToolbarBtn label="☑ List" ariaLabel="Task list" active={editor.isActive('taskList')} onClick={() => editor.chain().focus().toggleTaskList().run()} />
          <ToolbarBtn label="Code" ariaLabel="Code block" active={editor.isActive('codeBlock')} onClick={() => editor.chain().focus().toggleCodeBlock().run()} />

          <span className="sne-toolbar-divider" />

          {/* Block elements */}
          <ToolbarBtn label="Quote" ariaLabel="Blockquote" active={editor.isActive('blockquote')} onClick={() => editor.chain().focus().toggleBlockquote().run()} />
          <ToolbarBtn label="―" ariaLabel="Horizontal rule" active={false} onClick={() => editor.chain().focus().setHorizontalRule().run()} />

          <span className="sne-toolbar-divider" />

          {/* Link & Image */}
          <ToolbarBtn label="🔗" ariaLabel="Link" active={editor.isActive('link')} onClick={handleLinkClick} />
          <ToolbarBtn label="🖼" ariaLabel="Image" active={false} onClick={handleImageClick} />

          <span className="sne-toolbar-divider" />

          {/* Highlight */}
          <ToolbarBtn label="Highlight" ariaLabel="Highlight" active={editor.isActive('highlight')} onClick={() => editor.chain().focus().toggleHighlight().run()} />

          <span className="sne-toolbar-divider" />

          {/* Text alignment */}
          <ToolbarBtn label="⇤" ariaLabel="Align left" active={editor.isActive({ textAlign: 'left' })} onClick={() => editor.chain().focus().setTextAlign('left').run()} />
          <ToolbarBtn label="⇔" ariaLabel="Align center" active={editor.isActive({ textAlign: 'center' })} onClick={() => editor.chain().focus().setTextAlign('center').run()} />
          <ToolbarBtn label="⇥" ariaLabel="Align right" active={editor.isActive({ textAlign: 'right' })} onClick={() => editor.chain().focus().setTextAlign('right').run()} />

          <span className="sne-toolbar-divider" />

          {/* Line height */}
          <select
            className="sne-toolbar-select"
            title="Line height"
            aria-label="Line height"
            value={
              (LINE_HEIGHT_PRESETS.find((p) =>
                editor.isActive({ lineHeight: p.value })
              )?.value ?? '') as string
            }
            onChange={(e) => {
              const val = e.target.value as LineHeightValue | ''
              if (val === '') {
                editor.chain().focus().unsetLineHeight().run()
              } else {
                editor.chain().focus().setLineHeight(val).run()
              }
            }}
          >
            <option value="">Spacing</option>
            {LINE_HEIGHT_PRESETS.map((p) => (
              <option key={p.value} value={p.value}>{p.value}×</option>
            ))}
          </select>

          <span className="sne-toolbar-divider" />

          {/* Font size */}
          <select
            className="sne-toolbar-select"
            title="Font size"
            aria-label="Font size"
            value={editor.getAttributes('textStyle').fontSize ?? ''}
            onChange={(e) => {
              const val = e.target.value
              if (val === '') {
                editor.chain().focus().unsetFontSize().run()
              } else {
                editor.chain().focus().setFontSize(val).run()
              }
            }}
          >
            <option value="">Size</option>
            {FONT_SIZE_PRESETS.map((p) => (
              <option key={p.value} value={p.value}>{p.label}</option>
            ))}
          </select>

          {/* Text colour */}
          <select
            className="sne-toolbar-select"
            title="Text colour"
            aria-label="Text colour"
            value={editor.getAttributes('textStyle').color ?? ''}
            onChange={(e) => {
              const val = e.target.value
              if (val === '') {
                editor.chain().focus().unsetColor().run()
              } else {
                editor.chain().focus().setColor(val).run()
              }
            }}
          >
            {COLOR_PRESETS.map((p) => (
              <option key={p.value} value={p.value}>{p.label}</option>
            ))}
          </select>

          <span className="sne-toolbar-divider" />

          {/* Table */}
          <ToolbarBtn
            label="Table"
            ariaLabel="Insert table"
            active={inTable}
            onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}
          />
          {inTable && (
            <>
              <ToolbarBtn label="+Row" ariaLabel="Add row" active={false} onClick={() => editor.chain().focus().addRowAfter().run()} />
              <ToolbarBtn label="+Col" ariaLabel="Add column" active={false} onClick={() => editor.chain().focus().addColumnAfter().run()} />
              <ToolbarBtn label="-Row" ariaLabel="Delete row" active={false} onClick={() => editor.chain().focus().deleteRow().run()} />
              <ToolbarBtn label="-Col" ariaLabel="Delete column" active={false} onClick={() => editor.chain().focus().deleteColumn().run()} />
            </>
          )}

          <ToolbarBtn label="§" ariaLabel="Table of Contents" active={tocOpen} onClick={() => setTocOpen(v => !v)} />

          <span className="sne-toolbar-spacer" />
          {STATUS_LABELS[status] && (
            <span className={`sne-status sne-status--${status}`}>{STATUS_LABELS[status]}</span>
          )}
          <span className="sne-status sne-status--count">
            {editor.storage.characterCount.words()} words
          </span>
        </div>
      )}
      <div className="sne-editor-body">
        <EditorContent editor={editor} className="sne-content" />
        {tocOpen && headings.length > 0 && (
          <nav className="sne-toc" aria-label="Table of Contents">
            <div className="sne-toc-title">Contents</div>
            {headings.map((h, i) => (
              <button
                key={i}
                className={`sne-toc-item sne-toc-item--h${h.level}`}
                onClick={() => {
                  editor.chain().focus().setTextSelection(h.pos + 1).run()
                  const domNode = editor.view.domAtPos(h.pos + 1).node as HTMLElement
                  domNode?.scrollIntoView?.({ behavior: 'smooth', block: 'center' })
                }}
              >
                {h.text || '(untitled)'}
              </button>
            ))}
          </nav>
        )}
      </div>
      <BubbleMenu editor={editor}>
        <div className="sne-bubble-menu">
          <ToolbarBtn label="B" ariaLabel="Bold" active={editor.isActive('bold')} onClick={() => editor.chain().focus().toggleBold().run()} />
          <ToolbarBtn label="I" ariaLabel="Italic" active={editor.isActive('italic')} onClick={() => editor.chain().focus().toggleItalic().run()} />
          <ToolbarBtn label="U" ariaLabel="Underline" active={editor.isActive('underline')} onClick={() => editor.chain().focus().toggleUnderline().run()} />
          <ToolbarBtn label="S̶" ariaLabel="Strikethrough" active={editor.isActive('strike')} onClick={() => editor.chain().focus().toggleStrike().run()} />
          <span className="sne-toolbar-divider" />
          <ToolbarBtn label="🔗" ariaLabel="Link" active={editor.isActive('link')} onClick={handleLinkClick} />
          <ToolbarBtn label="Highlight" ariaLabel="Highlight" active={editor.isActive('highlight')} onClick={() => editor.chain().focus().toggleHighlight().run()} />
        </div>
      </BubbleMenu>
      <FloatingMenu editor={editor}>
        <div className="sne-floating-menu">
          <ToolbarBtn label="H1" ariaLabel="Heading 1" active={false} onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} />
          <ToolbarBtn label="H2" ariaLabel="Heading 2" active={false} onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} />
          <ToolbarBtn label="H3" ariaLabel="Heading 3" active={false} onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} />
          <span className="sne-toolbar-divider" />
          <ToolbarBtn label="• List" ariaLabel="Bullet list" active={false} onClick={() => editor.chain().focus().toggleBulletList().run()} />
          <ToolbarBtn label="1. List" ariaLabel="Ordered list" active={false} onClick={() => editor.chain().focus().toggleOrderedList().run()} />
          <ToolbarBtn label="Quote" ariaLabel="Blockquote" active={false} onClick={() => editor.chain().focus().toggleBlockquote().run()} />
          <ToolbarBtn label="―" ariaLabel="Horizontal rule" active={false} onClick={() => editor.chain().focus().setHorizontalRule().run()} />
          <ToolbarBtn label="Table" ariaLabel="Insert table" active={false} onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()} />
        </div>
      </FloatingMenu>
    </div>
  )
}

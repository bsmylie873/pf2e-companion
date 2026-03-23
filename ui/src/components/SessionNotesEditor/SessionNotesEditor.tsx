import { useEditor, EditorContent } from '@tiptap/react'
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
import type { JSONContent } from '@tiptap/react'
import type { Session } from '../../types/session'
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
  version: number
  editable?: boolean
  onSave?: (content: JSONContent, version: number) => Promise<Session>
}

const IMAGE_URL_RE = /^https?:\/\/\S+\.(?:png|jpe?g|gif|webp|svg|bmp|ico|avif)(?:\?[^\s]*)?$/i

function isImageUrl(text: string): boolean {
  return IMAGE_URL_RE.test(text.trim())
}

const STATUS_LABELS: Record<string, string> = {
  idle: '',
  saving: 'Saving…',
  saved: 'Saved ✦',
  conflict: 'Conflict — reload to sync',
  error: 'Save failed',
}

export default function SessionNotesEditor({
  initialContent,
  version,
  editable = true,
  onSave,
}: SessionNotesEditorProps) {
  const { scheduleAutosave, status } = useAutosave(
    onSave ?? (() => Promise.reject(new Error('No onSave provided'))),
    version,
  )

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
      Placeholder.configure({ placeholder: 'Begin your session chronicle…' }),
      TextAlign.configure({ types: ['heading', 'paragraph'] }),
      Image,
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

          <span className="sne-toolbar-spacer" />
          {STATUS_LABELS[status] && (
            <span className={`sne-status sne-status--${status}`}>{STATUS_LABELS[status]}</span>
          )}
        </div>
      )}
      <EditorContent editor={editor} className="sne-content" />
    </div>
  )
}

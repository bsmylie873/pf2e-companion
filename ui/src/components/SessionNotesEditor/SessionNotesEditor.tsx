import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import type { JSONContent } from '@tiptap/react'
import type { Session } from '../../types/session'
import { useAutosave } from '../../hooks/useAutosave'
import './SessionNotesEditor.css'

interface SessionNotesEditorProps {
  initialContent: JSONContent | null
  version: number
  onSave: (content: JSONContent, version: number) => Promise<Session>
}

const STATUS_LABELS: Record<string, string> = {
  idle: '',
  saving: 'Saving…',
  saved: 'Saved ✦',
  conflict: 'Conflict — reload to sync',
  error: 'Save failed',
}

export default function SessionNotesEditor({ initialContent, version, onSave }: SessionNotesEditorProps) {
  const { scheduleAutosave, status } = useAutosave(onSave, version)

  const editor = useEditor({
    extensions: [StarterKit],
    content: initialContent ?? '',
    onUpdate: ({ editor }) => {
      scheduleAutosave(editor.getJSON())
    },
  })

  if (!editor) return null

  const ToolbarBtn = ({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) => (
    <button
      className={`sne-toolbar-btn${active ? ' sne-toolbar-btn--active' : ''}`}
      onMouseDown={(e) => { e.preventDefault(); onClick() }}
      title={label}
      aria-label={label}
      aria-pressed={active}
    >
      {label}
    </button>
  )

  return (
    <div className="session-notes-editor">
      <div className="sne-toolbar">
        <ToolbarBtn label="B" active={editor.isActive('bold')} onClick={() => editor.chain().focus().toggleBold().run()} />
        <ToolbarBtn label="I" active={editor.isActive('italic')} onClick={() => editor.chain().focus().toggleItalic().run()} />
        <span className="sne-toolbar-divider" />
        <ToolbarBtn label="H1" active={editor.isActive('heading', { level: 1 })} onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} />
        <ToolbarBtn label="H2" active={editor.isActive('heading', { level: 2 })} onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} />
        <ToolbarBtn label="H3" active={editor.isActive('heading', { level: 3 })} onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} />
        <span className="sne-toolbar-divider" />
        <ToolbarBtn label="• List" active={editor.isActive('bulletList')} onClick={() => editor.chain().focus().toggleBulletList().run()} />
        <ToolbarBtn label="1. List" active={editor.isActive('orderedList')} onClick={() => editor.chain().focus().toggleOrderedList().run()} />
        <ToolbarBtn label="Code" active={editor.isActive('codeBlock')} onClick={() => editor.chain().focus().toggleCodeBlock().run()} />
        <span className="sne-toolbar-spacer" />
        {STATUS_LABELS[status] && (
          <span className={`sne-status sne-status--${status}`}>{STATUS_LABELS[status]}</span>
        )}
      </div>
      <EditorContent editor={editor} className="sne-content" />
    </div>
  )
}

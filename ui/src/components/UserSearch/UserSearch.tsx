import { useState, useEffect, useRef, useCallback } from 'react'
import { apiFetch } from '../../api/client'
import type { User } from '../../types/user'
import './UserSearch.css'

interface UserSearchProps {
  excludeIds: string[]
  onSelect: (user: User) => void
}

export default function UserSearch({ excludeIds, onSelect }: UserSearchProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<User[]>([])
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)
  const allUsersRef = useRef<User[] | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const listboxId = 'us-listbox'

  const getActiveId = (index: number) =>
    index >= 0 && results[index] ? `us-option-${results[index].id}` : undefined

  const filterUsers = useCallback(
    (q: string, users: User[]) => {
      if (!q.trim()) return []
      const lower = q.toLowerCase()
      return users.filter(
        (u) =>
          u.username.toLowerCase().includes(lower) &&
          !excludeIds.includes(u.id),
      )
    },
    [excludeIds],
  )

  useEffect(() => {
    if (!query.trim()) {
      setResults([])
      setOpen(false)
      return
    }

    if (allUsersRef.current !== null) {
      const filtered = filterUsers(query, allUsersRef.current)
      setResults(filtered)
      setOpen(filtered.length > 0)
      setActiveIndex(-1)
      return
    }

    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(async () => {
      try {
        const users = await apiFetch<User[]>('/users')
        allUsersRef.current = users
        const filtered = filterUsers(query, users)
        setResults(filtered)
        setOpen(filtered.length > 0)
        setActiveIndex(-1)
      } catch {
        // silently fail
      }
    }, 250)
  }, [query, filterUsers])

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (!open) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActiveIndex((i) => Math.min(i + 1, results.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActiveIndex((i) => Math.max(i - 1, -1))
    } else if (e.key === 'Enter' && activeIndex >= 0) {
      e.preventDefault()
      selectUser(results[activeIndex])
    } else if (e.key === 'Escape') {
      setOpen(false)
      setActiveIndex(-1)
    }
  }

  function selectUser(user: User) {
    onSelect(user)
    setQuery('')
    setResults([])
    setOpen(false)
    setActiveIndex(-1)
  }

  function handleBlur() {
    setTimeout(() => {
      setOpen(false)
      setActiveIndex(-1)
    }, 150)
  }

  return (
    <div className="us-root">
      <input
        className="us-input ncf-input"
        type="text"
        role="combobox"
        aria-expanded={open}
        aria-autocomplete="list"
        aria-controls={listboxId}
        aria-activedescendant={getActiveId(activeIndex)}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={handleBlur}
        placeholder="Search by username..."
        autoComplete="off"
      />
      {open && (
        <ul
          id={listboxId}
          className="us-dropdown"
          role="listbox"
        >
          {results.map((user, idx) => (
            <li
              key={user.id}
              id={`us-option-${user.id}`}
              role="option"
              aria-selected={idx === activeIndex}
              className={`us-option${idx === activeIndex ? ' us-option--active' : ''}`}
              onMouseDown={() => selectUser(user)}
            >
              <span className="us-option-name">{user.username}</span>
              <span className="us-option-email">{user.email}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

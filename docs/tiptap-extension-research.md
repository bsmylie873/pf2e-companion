# Tiptap Extension Research: Enhancing the Session Notes Editor

**Date:** 2026-03-23
**Status:** Draft — Pending Review
**Scope:** `SessionNotesEditor` component (`ui/src/components/SessionNotesEditor/SessionNotesEditor.tsx`)

---

## 1. Introduction

This document evaluates candidate Tiptap extensions for expanding the session notes editor's formatting capabilities. The goal is to give GMs stronger authoring tools for composing session notes and improve readability for players consuming published notes.

### Current State

The `SessionNotesEditor` component uses `@tiptap/starter-kit` v3.20.4 as its sole extension configuration. The toolbar surfaces the following controls:

| Control | Extension | Source |
|---|---|---|
| Bold | `@tiptap/extension-bold` | StarterKit |
| Italic | `@tiptap/extension-italic` | StarterKit |
| H1, H2, H3 | `@tiptap/extension-heading` | StarterKit |
| Bullet List | `@tiptap/extension-bullet-list` | StarterKit |
| Ordered List | `@tiptap/extension-ordered-list` | StarterKit |
| Code Block | `@tiptap/extension-code-block` | StarterKit |

StarterKit also registers several extensions that are **active in the schema but have no toolbar controls**: Blockquote, Strike, Horizontal Rule, Link, and Underline. These node/mark types are already parseable from stored `JSONContent` — they simply lack UI entry points.

### Storage Format

Session notes are stored as Tiptap `JSONContent` (ProseMirror JSON) in the `sessions.notes` JSONB column. Adding new extensions introduces new node or mark types into this JSON. Backwards compatibility with existing stored content is not required, but the seed data in `database/local-data-seed/V1__seed.sql` must be updated to reflect any new node types. The current seed data uses empty JSON objects (`'{}'::jsonb`) for all session notes, so no migration of existing seed content is needed.

### Read-Only Context

`SessionNotes.tsx` currently renders `SessionNotesEditor` for all users without distinguishing between editing and viewing. There is no `editable: false` configuration anywhere in the codebase. This means every extension must be evaluated for both its authoring UX and its passive rendering quality.

---

## 2. Evaluation Criteria

Each candidate is assessed on:

| Criterion | Description |
|---|---|
| **UX Value** | How much the feature improves authoring (GM) and readability (player) |
| **Integration Effort** | Installation, toolbar UI, CSS styling, and any non-trivial configuration |
| **Already Installed** | Whether the extension is bundled with StarterKit (zero install) or requires a new dependency |
| **New JSON Types** | Node or mark types added to stored `JSONContent` |
| **Read-Only Rendering** | Quality of passive display without toolbar interaction; any gaps |
| **Accessibility Flags** | Known a11y concerns for the follow-up implementation ticket |

---

## 3. Candidate Extensions

### 3.1 Blockquote — Zero Install

**Package:** `@tiptap/extension-blockquote` (bundled with StarterKit)
**JSON type:** `blockquote` node (already registered)
**Status:** Active in schema — needs toolbar button only

**UX Value — High.** Blockquotes are the most natural formatting tool for session notes: quoting NPC dialogue, calling out prophecies, summarising plot hooks. GMs can use them to set apart important narrative text; players benefit from immediate visual distinction when scanning notes.

**Integration Effort — Trivial.** Add a single `ToolbarBtn` calling `editor.chain().focus().toggleBlockquote().run()`. CSS styles already exist in `SessionNotesEditor.css` (`.tiptap blockquote` — border-left accent, italic, muted colour).

**Read-Only Rendering — Excellent.** Blockquotes render as styled `<blockquote>` elements with the existing CSS. No gaps.

**Accessibility Flags — Minimal.** Standard `<blockquote>` semantics; screen readers handle this natively. Toolbar button needs `aria-label` (already the pattern used by `ToolbarBtn`).

---

### 3.2 Horizontal Rule — Zero Install

**Package:** `@tiptap/extension-horizontal-rule` (bundled with StarterKit)
**JSON type:** `horizontalRule` node (already registered)
**Status:** Active in schema — needs toolbar button only

**UX Value — Medium.** Useful for separating narrative sections (e.g., time jumps, scene breaks, "meanwhile..." transitions). GMs structuring longer session recaps benefit from a visual separator beyond headings.

**Integration Effort — Trivial.** Add one `ToolbarBtn` calling `editor.chain().focus().setHorizontalRule().run()`. CSS already exists (`.tiptap hr` — border-top with theme colour).

**Read-Only Rendering — Excellent.** Renders as a styled `<hr>`. No gaps.

**Accessibility Flags — None.** Semantic HTML.

---

### 3.3 Strike — Zero Install

**Package:** `@tiptap/extension-strike` (bundled with StarterKit)
**JSON type:** `strike` mark (already registered)
**Status:** Active in schema — needs toolbar button only

**UX Value — Low-Medium.** Situationally useful for marking retconned information or tracking changes informally ("the party ~~fought~~ negotiated with the bandits"). Not a primary formatting tool but a nice-to-have.

**Integration Effort — Trivial.** One `ToolbarBtn` calling `toggleStrike()`. No CSS needed — browsers render `<s>` natively.

**Read-Only Rendering — Good.** Native `<s>` styling. May want a subtle CSS override for theme consistency.

**Accessibility Flags — Minimal.** `<s>` has native semantics; screen readers announce "deleted text" by default.

---

### 3.4 Underline — Zero Install

**Package:** `@tiptap/extension-underline` (bundled with StarterKit)
**JSON type:** `underline` mark (already registered)
**Status:** Active in schema — needs toolbar button only

**UX Value — Low.** Underline is generally discouraged in digital content (conflicts with link styling). However, some TTRPG players expect it as a standard formatting option. Including it satisfies expectations without much effort.

**Integration Effort — Trivial.** One `ToolbarBtn` calling `toggleUnderline()`. No CSS needed.

**Read-Only Rendering — Good.** Native `<u>` styling. Risk of confusion with links if Link extension is also enabled — may need a CSS distinction (links get colour + underline; raw underline is text-colour only).

**Accessibility Flags — Minimal.** Potential confusion with links for screen reader users if both are present. Mitigated by ensuring links use `<a>` elements with `href`.

---

### 3.5 Link — Zero Install

**Package:** `@tiptap/extension-link` (bundled with StarterKit, depends on `linkifyjs`)
**JSON type:** `link` mark with `href`, `target`, `rel` attributes (already registered)
**Status:** Active in schema — needs toolbar UI

**UX Value — High.** Directly addresses the user context requirement: "the ability to link external resources." Players and GMs regularly reference external wikis (Archives of Nethys, Pathfinder Wiki), shared maps, loot trackers, and homebrew documents. Links are essential for a useful session journal.

**Integration Effort — Moderate.** Unlike toggle buttons, Link requires a URL input flow:
- Option A: Toolbar button opens a small prompt/popover for URL entry (`window.prompt()` as MVP, custom popover as follow-up)
- Option B: Use `@tiptap/extension-bubble-menu` (already installed) to show a link editor when text with a link mark is selected
- Configuration: `openOnClick: false` in edit mode (prevents navigation while editing), `openOnClick: true` in read-only mode; `autolink: true` to auto-detect URLs during typing

**Read-Only Rendering — Good with caveat.** Links render as `<a>` tags. In the current setup (no `editable: false`), links are not clickable because ProseMirror captures clicks. This is a strong argument for a read-only display mode (see Section 3.10).

**Accessibility Flags — Moderate.** Link popover/prompt needs keyboard accessibility. `<a>` tags need visible focus styles. Auto-linked URLs should be distinguishable from surrounding text.

---

### 3.6 Table — New Install Required

**Package:** `@tiptap/extension-table`, `@tiptap/extension-table-row`, `@tiptap/extension-table-cell`, `@tiptap/extension-table-header`
**Alternative:** `@tiptap/extension-table-kit` (registers all four in one call)
**JSON types:** `table`, `tableRow`, `tableCell`, `tableHeader` nodes (new)
**License:** MIT, open source

**UX Value — High.** Tables are the second-most requested formatting feature after links. TTRPG session notes regularly include loot tables, initiative trackers, NPC stat blocks, encounter summaries, and shopping inventories. Without tables, players resort to poorly formatted lists or external spreadsheets.

**Integration Effort — High.** Tables are the most complex extension to integrate:
- 4 new npm packages (or 1 kit package)
- Toolbar needs: Insert Table, Add Row, Add Column, Delete Row, Delete Column, Merge Cells, Toggle Header
- Substantial CSS required for table borders, cell padding, header styling, selected-cell highlighting, resize handles
- Table interactions (cell selection, drag-to-resize) add significant UI surface area
- The `@tiptap/extension-table` extension recommends disabling the `HTMLAttributes` on `tableCell` for custom cell background colours, which requires additional configuration

**Read-Only Rendering — Good.** Tables render as standard `<table>` HTML. Styling carries over. Cell selection highlighting should be suppressed in read-only mode.

**Accessibility Flags — Significant.** Tables require proper `<th>` scope attributes, keyboard navigation between cells, and screen reader announcements for cell position. This is the extension most likely to require dedicated accessibility work in the follow-up ticket.

---

### 3.7 Highlight — New Install Required

**Package:** `@tiptap/extension-highlight`
**JSON type:** `highlight` mark with optional `color` attribute (new)
**License:** MIT, open source

**UX Value — Medium.** Useful for emphasising key information: critical plot points, treasure descriptions, important NPC names. Multi-colour highlighting could serve as a lightweight tagging system (e.g., yellow for loot, red for combat, green for RP moments).

**Integration Effort — Low-Moderate.**
- 1 new npm package
- Single toolbar button for default highlight; colour picker dropdown for multi-colour mode
- CSS for `<mark>` element in both light and dark themes (default browser yellow is poor in dark mode)
- Markdown shortcut `==text==` works out of the box

**Read-Only Rendering — Good with caveat.** Renders as `<mark>` with inline colour. Default yellow background needs CSS override for dark theme legibility.

**Accessibility Flags — Low.** `<mark>` has semantic meaning. Colour alone should not convey meaning — pair with another visual indicator if used as a tagging system.

---

### 3.8 Task List — New Install Required

**Package:** `@tiptap/extension-task-list`, `@tiptap/extension-task-item`
**JSON types:** `taskList`, `taskItem` nodes with `checked` attribute (new)
**License:** MIT, open source

**UX Value — Medium.** Useful for tracking action items between sessions: "buy healing potions", "ask the blacksmith about the sword", "return to the haunted tower". GMs can create preparation checklists.

**Integration Effort — Moderate.**
- 2 new npm packages
- Toolbar button for inserting/toggling a task list
- CSS for checkbox styling (custom checkboxes to match the medieval manuscript aesthetic)
- Checkboxes are interactive in edit mode — need to decide whether they should be toggleable in read-only mode (useful for players tracking personal to-dos)

**Read-Only Rendering — Functional but needs attention.** Renders as `<ul>` with `<input type="checkbox">` elements. Checkboxes are disabled in read-only mode by default. Custom CSS needed to match theme. Whether checkboxes should be interactive in read-only mode is a product decision.

**Accessibility Flags — Moderate.** Checkboxes need proper `<label>` association and keyboard operability. Tiptap's default implementation may not include labels.

---

### 3.9 Placeholder — New Install Required (Utility)

**Package:** `@tiptap/extension-placeholder`
**JSON types:** None (rendering-only extension)
**License:** MIT, open source

**UX Value — Medium.** Shows contextual placeholder text ("Begin your session chronicle...") in the empty editor. Improves the blank-page experience for new users. The CSS for placeholder text already exists in `SessionNotesEditor.css` (`.tiptap p.is-editor-empty:first-child::before`), but it relies on a `data-placeholder` attribute that only works with the Placeholder extension installed.

**Integration Effort — Trivial.** 1 package, single configuration line: `Placeholder.configure({ placeholder: 'Begin your session chronicle...' })`. The existing CSS handles rendering.

**Read-Only Rendering — N/A.** Placeholder text only appears when the editor is empty and focused. Not relevant for read-only display.

**Accessibility Flags — None.** Placeholder is decorative.

---

### 3.10 Read-Only Display Mode — No Install Required

**Not an extension** — this is an architectural feature of the `SessionNotesEditor` / `SessionNotes` page.

**UX Value — High.** Currently, `SessionNotes.tsx` renders the full editor (with toolbar) for all users, including players who may only be reading. This creates several problems:
1. Links are not clickable (ProseMirror captures click events in edit mode)
2. The toolbar is visible but may be confusing or misleading for read-only users
3. The editing cursor appears, suggesting editability even when the user has no intent to edit
4. Any future interactive elements (task list checkboxes) behave differently in edit vs read mode

**Implementation approaches:**
- **Option A — `editable: false` prop.** Pass an `editable` boolean to `SessionNotesEditor`. When false, configure `useEditor({ editable: false })` and hide the toolbar. Tiptap natively suppresses the cursor, makes links clickable, and renders all content as static HTML. This is the lowest-effort approach.
- **Option B — Separate read-only renderer.** Render stored `JSONContent` to HTML server-side or via `generateHTML()` from `@tiptap/core`, then display in a styled `<div>`. Avoids loading the full ProseMirror editor for read-only users (performance benefit), but requires maintaining a parallel rendering path and registering the same extensions in both.
- **Recommended: Option A** for the follow-up ticket. Option B is a future optimisation if performance profiling shows the editor is a bottleneck for read-only users.

**Integration Effort — Low (Option A).** Add an `editable` prop to `SessionNotesEditorProps`, conditionally hide the toolbar, pass `editable` to `useEditor()`. The parent page (`SessionNotes.tsx`) determines editability based on user role or an explicit mode toggle.

**Read-Only Rendering — This IS the read-only rendering solution.** All extension assessments above note read-only behaviour; this feature ensures that behaviour is actually reachable.

**Accessibility Flags — Low.** Removing the toolbar for read-only users simplifies the accessible interface. The content area should have `role="document"` rather than a contenteditable region.

---

### 3.11 Text Align — New Install Required

**Package:** `@tiptap/extension-text-align`
**JSON types:** Adds `textAlign` attribute to paragraph and heading nodes (new attributes, not new nodes)
**License:** MIT, open source

**UX Value — Low.** Text alignment (left, centre, right, justify) has limited utility for session notes. Centre-aligned headings could be a minor aesthetic improvement, but the current CSS handles heading styling well. Not a commonly requested feature for this context.

**Integration Effort — Low.** 1 package, toolbar buttons for alignment options, minimal CSS.

**Read-Only Rendering — Good.** Alignment attributes render as inline styles. No gaps.

**Accessibility Flags — None.**

---

### 3.12 Image — New Install Required

**Package:** `@tiptap/extension-image`
**JSON type:** `image` node with `src`, `alt`, `title` attributes (new)
**License:** MIT, open source

**UX Value — Medium-High conceptually, but constrained.** Images in session notes (maps, character art, scene illustrations) would be extremely valuable. However, the Image extension only renders images — it does not handle upload. Image upload requires either:
- Base64 encoding (bloats JSONB storage, violates the "low storage usage" priority)
- An image upload endpoint + file storage (significant backend work, out of scope)
- External URLs only (requires users to host images elsewhere)

**Integration Effort — High (overall system).** The extension itself is simple (1 package, toolbar button, image node CSS). But without an upload solution, usability is poor. External-URL-only mode is functional but a degraded experience.

**Read-Only Rendering — Good.** `<img>` tags render natively. Need `max-width: 100%` CSS to prevent overflow.

**Accessibility Flags — Moderate.** Images require `alt` text. The insert flow should prompt for alt text. Screen readers need meaningful descriptions.

**Recommendation:** Defer until an image upload solution is designed. The extension itself is low-effort, but without upload infrastructure it provides limited value.

---

## 4. Comparison Matrix

| Extension | UX Value | Effort | Installed | New JSON Types | Read-Only Quality | A11y Flags |
|---|---|---|---|---|---|---|
| **Blockquote** | High | Trivial | Yes | None | Excellent | Minimal |
| **Link** | High | Moderate | Yes | None | Needs read-only mode | Moderate |
| **Horizontal Rule** | Medium | Trivial | Yes | None | Excellent | None |
| **Strike** | Low-Med | Trivial | Yes | None | Good | Minimal |
| **Underline** | Low | Trivial | Yes | None | Good | Minimal |
| **Read-Only Mode** | High | Low | N/A | None | This IS the solution | Low |
| **Table** | High | High | No | 4 node types | Good | Significant |
| **Highlight** | Medium | Low-Mod | No | 1 mark type | Good (needs dark CSS) | Low |
| **Task List** | Medium | Moderate | No | 2 node types | Functional | Moderate |
| **Placeholder** | Medium | Trivial | No | None | N/A | None |
| **Text Align** | Low | Low | No | Attributes only | Good | None |
| **Image** | Deferred | High (system) | No | 1 node type | Good | Moderate |

---

## 5. Prioritised Feature Shortlist

> **Decision (2026-03-23):** All 12 features will be implemented in a single pass. The original tiered deferral of Image and Text Align has been overridden — all extensions are included.

### Tier 1 — Zero-Install (StarterKit already bundles these)

1. **Blockquote** — Add toolbar button. Highest-value zero-effort feature.
2. **Horizontal Rule** — Add toolbar button. Simple scene/section separator.
3. **Link** — Add toolbar button + URL input flow. Essential for external references.
4. **Strike** — Add toolbar button. Low-effort completeness.
5. **Underline** — Add toolbar button. Low-effort completeness.

### Tier 2 — Low-Effort, High-Impact

6. **Read-Only Display Mode** — Add `editable` prop to `SessionNotesEditor`, hide toolbar when false. Prerequisite for links to work for readers.
7. **Placeholder** — Install `@tiptap/extension-placeholder`, configure placeholder text. CSS already exists.

### Tier 3 — New Installs

8. **Table** — Install `@tiptap/extension-table` + row/cell/header packages. Toolbar controls for insert, add/delete row/column.
9. **Highlight** — Install `@tiptap/extension-highlight`. Toggle button with default colour.
10. **Task List** — Install `@tiptap/extension-task-list` + `@tiptap/extension-task-item`. Checkboxes disabled in read-only mode.
11. **Image** — Install `@tiptap/extension-image`. External-URL-only mode (no upload infrastructure). Toolbar button prompts for image URL and alt text.
12. **Text Align** — Install `@tiptap/extension-text-align`. Toolbar buttons for left, centre, right alignment.

---

## 6. Seed Data Impact

The current seed data uses empty JSON objects (`'{}'::jsonb`) for all session notes. The seed data should be updated to include example content demonstrating all new node types (table, highlight, taskList, taskItem, image) so that developers can verify rendering during development.

---

## 7. Implementation Summary

**Scope:** Implement all 12 features in a single pass. Add toolbar controls for all extensions, add `editable` prop for read-only mode, install all new npm packages, update seed data with rich example content, and add CSS for all new node types.

**New npm packages:**
- `@tiptap/extension-table`
- `@tiptap/extension-table-row`
- `@tiptap/extension-table-cell`
- `@tiptap/extension-table-header`
- `@tiptap/extension-highlight`
- `@tiptap/extension-task-list`
- `@tiptap/extension-task-item`
- `@tiptap/extension-placeholder`
- `@tiptap/extension-text-align`
- `@tiptap/extension-image`

**Files modified:**
- `ui/package.json` — new dependencies
- `ui/src/components/SessionNotesEditor/SessionNotesEditor.tsx` — all extensions, toolbar, editable prop
- `ui/src/components/SessionNotesEditor/SessionNotesEditor.css` — styles for all new node types
- `ui/src/pages/SessionNotes/SessionNotes.tsx` — pass editable prop
- `database/local-data-seed/V1__seed.sql` — rich example session notes content

**Estimated Effort:** Medium (2-3 days)

---

## 8. References

- [Tiptap Extensions Overview](https://tiptap.dev/docs/editor/extensions/overview)
- [Tiptap Table Extension](https://tiptap.dev/docs/editor/extensions/nodes/table)
- [Tiptap Highlight Extension](https://tiptap.dev/docs/editor/extensions/marks/highlight)
- [Tiptap Image Extension](https://tiptap.dev/docs/editor/extensions/nodes/image)
- [Tiptap Link Extension](https://tiptap.dev/docs/editor/extensions/marks/link)
- [Tiptap Placeholder Extension](https://tiptap.dev/docs/editor/extensions/functionality/placeholder)
- [Existing research: `docs/rich-text-editor-research.md`](../docs/rich-text-editor-research.md) — Editor library selection (related but distinct)

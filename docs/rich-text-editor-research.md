# Rich-Text Editor Research: Collaborative Editing for Session Notes

**Date:** 2026-03-23
**Status:** Implemented (v1 REST-only → upgraded to WebSocket/OT)
**Scope:** `sessions.notes` JSONB column — rich-text editing in the Editor view

---

## 1. Introduction & Scope

This document evaluates candidate rich-text editor libraries for enabling collaborative editing of session notes within the PF2E Companion application. The goal is to allow multiple game members (GMs and Players) to edit session notes simultaneously through a rich-text interface, with changes persisted automatically.

### Scope

- **In scope:** Editing the `sessions.notes` JSONB column when a session is opened in the Editor view.
- **Out of scope:** Shared and private notes stored in the `notes` table.

### System Context

Real-time editing capability is needed to prevent overwrite conflicts when multiple GMs or players edit the same session notes asynchronously between game sessions. Without a coordinated editing solution, concurrent edits to `sessions.notes` risk silently overwriting one another, degrading data integrity.

### Scale Assumptions

Per `database-schema.md` section 1:

- Fewer than 100 concurrent users in total
- Maximum of 10 simultaneous editors per document
- Last-write-wins conflict resolution is acceptable at this scale
- CRDT/OT-based sync infrastructure is not a hard requirement, though advantages should be noted

### Access Control Prerequisite

`database-schema.md` section 5.7 currently grants Players only "Read access" to game-level data, the category under which sessions fall. The intended access model for session notes is that Players have **read/write access**, consistent with their access to shared notes. Section 5.7 must be updated to reflect this before implementation begins to prevent ambiguity during development.

### Current Architecture

| Component | Detail |
|---|---|
| Backend | Go/Echo REST API — no WebSocket routes in `backend/main.go` |
| Frontend | React 19.2.4, Vite 8, no existing editor library (`ui/package.json`) |
| Session model | `notes` is `datatypes.JSON` (JSONB), `version` is `INTEGER NOT NULL DEFAULT 1` (`backend/models/models.go`) |
| Session PATCH | `PATCH /sessions/:id` accepts partial updates; JSONB columns replaced wholesale |
| Frontend API | `updateSession()` in `ui/src/api/sessions.ts` calls `PATCH /sessions/:id` |
| Session type | `notes` field typed as `unknown` in `ui/src/types/session.ts` |
| Editor page | `ui/src/pages/Editor/Editor.tsx` — displays session cards, no rich-text editing implemented |

---

## 2. Evaluation Criteria

Each candidate is evaluated against the following criteria:

| Criterion | Description |
|---|---|
| **React 19 Compatibility** | Works with React 19.2.x as installed in `ui/package.json` |
| **Real-time Collaboration** | Multiple users editing simultaneously with conflict resolution |
| **Autosave** | Persists changes within 5 seconds of the last edit with no explicit user action required |
| **JSONB-compatible Output** | Native JSON serialization directly storable in PostgreSQL JSONB |
| **Licensing** | Open-source license type and any commercial/paid tiers |
| **Total Cost of Ownership** | Infrastructure, hosting, and operational complexity costs |
| **`sessions.version` Integration** | Whether the conflict resolution model can integrate with the existing integer version column |
| **Server-side Changes Required** | WebSocket, SSE, sync server, or other requirements beyond the current REST API |

---

## 3. Candidate Evaluations

### Candidate A: Tiptap + Yjs (Self-hosted via Hocuspocus)

[Tiptap](https://tiptap.dev/) is a headless rich-text editor framework built on ProseMirror. [Yjs](https://yjs.dev/) is a CRDT implementation for real-time collaboration. [Hocuspocus](https://tiptap.dev/docs/hocuspocus/introduction) is Tiptap's open-source Node.js WebSocket backend for Yjs synchronization.

**React 19 Compatibility:** Supported. `@tiptap/react` v3.x lists React 19+ as a peer dependency. Tiptap v2.10+ added React 19 support.

**Real-time Collaboration:** Full CRDT-based collaboration via `@tiptap/extension-collaboration` (MIT). Yjs handles conflict-free merging of concurrent edits automatically. Supports live cursors and user presence indicators via `@tiptap/extension-collaboration-cursor`. Hocuspocus serves as the WebSocket backend, managing document state, broadcasting updates between clients, and persisting changes.

**Autosave:** Two mechanisms available:
1. **With Hocuspocus:** The `onStoreDocument` hook fires when the document changes, allowing server-side persistence to the database. Configurable debounce interval (e.g., 2-3 seconds) ensures changes are written within the 5-second threshold.
2. **Client-side fallback:** The editor's `onUpdate` callback fires on every change; a debounced handler (2-3 seconds) can call the REST API directly.

**JSONB Output:** `editor.getJSON()` returns ProseMirror JSON — a tree structure with `type`, `content`, `marks`, and `attrs` fields. This is a stable, well-documented format that stores directly as PostgreSQL JSONB with no transformation required.

**License:** MIT for the core editor, `@tiptap/extension-collaboration`, `@tiptap/extension-collaboration-cursor`, and Hocuspocus. Pro extensions (comments, AI assistance, document history) require a paid Tiptap Cloud subscription ($49-$999/month based on document count). These pro extensions are **not required** for the functionality evaluated here.

**Total Cost of Ownership:** $0 at v1 scale when self-hosting Hocuspocus. However, this requires deploying and maintaining a separate Node.js WebSocket server alongside the Go backend — adding a new runtime dependency, deployment artifact, health monitoring, and connection management.

**`sessions.version` Integration:** Yjs CRDTs manage conflict resolution internally through vector clocks and document state vectors — the `sessions.version` column is not used as the conflict resolution mechanism. However, it can be incremented on each Hocuspocus `onStoreDocument` callback as an application-level counter for audit or UI display purposes. The version column becomes supplementary, not authoritative.

**Server-side Changes:** Requires a separate Node.js Hocuspocus server running WebSockets. This cannot run inside the Go/Echo process. An alternative approach is a minimal Go WebSocket relay that forwards raw Yjs binary updates between clients without any Yjs-specific logic on the server; however, this pattern is less battle-tested than Hocuspocus and loses access to Hocuspocus's persistence hooks.

---

### Candidate B: Lexical + Yjs (Self-hosted WebSocket)

[Lexical](https://lexical.dev/) is a Meta-backed extensible text editor framework. It integrates with Yjs for real-time collaboration via the `@lexical/yjs` package.

**React 19 Compatibility:** Supported. `@lexical/react` is compatible with React 17+ and has been tested with React 19.

**Real-time Collaboration:** Full CRDT-based collaboration via `@lexical/yjs` and `LexicalCollaborationPlugin`. Requires a Yjs WebSocket provider — either `y-websocket` (Node.js reference server) or a custom Go WebSocket relay. Supports live cursors and user awareness through the Yjs awareness protocol.

**Autosave:** `editor.registerUpdateListener()` fires on every editor state change. A debounced handler (2-3 seconds) can persist the document. With Yjs, the sync provider handles real-time updates between clients, and the server can persist on document change — same pattern as Tiptap with Hocuspocus.

**JSONB Output:** `editorState.toJSON()` produces a JSON tree: `{ root: { children: [...], type: "root", version: 1 } }`. This is directly storable as PostgreSQL JSONB. The format is Lexical-specific and less widely adopted than ProseMirror JSON, though equally functional for storage and retrieval.

**License:** MIT for all packages, including `@lexical/react`, `@lexical/yjs`, and all Lexical plugins. No paid tiers or commercial restrictions.

**Total Cost of Ownership:** $0. Fully free with no vendor lock-in. Like Candidate A, it requires running a WebSocket server for Yjs sync — either the `y-websocket` Node.js reference server or a custom Go WebSocket relay.

**`sessions.version` Integration:** Same as Candidate A — Yjs CRDTs handle conflict resolution internally. The `sessions.version` column can be incremented as an application-level counter on each persist but is not the conflict resolution mechanism.

**Server-side Changes:** Requires a WebSocket server for Yjs synchronization. The `y-websocket` package provides a Node.js reference server. A minimal Go WebSocket relay is feasible — it only needs to broadcast binary Yjs update messages between connected clients, with no Yjs-specific logic required on the server side. Neither option exists in the current architecture.

---

### Candidate C: Tiptap (REST-only, Polling-based Collaboration)

This candidate uses the same Tiptap editor as Candidate A but without Yjs or Hocuspocus. Collaboration is handled through optimistic locking against the existing REST API, using the `sessions.version` column for conflict detection.

**React 19 Compatibility:** Supported — same as Candidate A.

**Real-time Collaboration:** No real-time CRDT sync. Instead, collaboration is handled via optimistic locking: the client sends `{ notes: <json>, version: N }` with each save, and the server rejects the update if the stored version does not match. On rejection, the client must re-fetch the latest document and either prompt the user to reconcile or apply a merge strategy. Live cursors and presence indicators are not supported.

**Autosave:** Debounced `onUpdate` callback (2-3 seconds) triggers a `PATCH /sessions/:id` request with the current editor JSON and the known version number. This is straightforward to configure within the 5-second threshold. The autosave mechanism is:
1. Editor emits `onUpdate` on every change
2. Debounce timer resets on each change (e.g., 2000ms)
3. On debounce expiry, call `PATCH /sessions/:id` with `{ notes: editor.getJSON(), version: currentVersion }`
4. On success, update local `currentVersion` to the returned incremented value
5. On version conflict (HTTP 409), re-fetch and notify the user

**JSONB Output:** Same ProseMirror JSON as Candidate A — `editor.getJSON()` returns an identical format regardless of whether Yjs collaboration is enabled. This means the stored data format is forward-compatible with a future migration to Candidate A.

**License:** MIT. No Hocuspocus or Cloud subscription needed — only the core Tiptap packages.

**Total Cost of Ownership:** $0 with no additional infrastructure. Uses the existing Go/Echo REST API and `PATCH /sessions/:id` endpoint as-is. The only backend change is adding optimistic locking logic to the session update handler.

**`sessions.version` Integration:** Direct integration — `sessions.version` becomes the **primary** conflict resolution mechanism. The update query becomes:

```sql
UPDATE sessions
SET notes = $1, version = version + 1, updated_at = NOW()
WHERE id = $2 AND version = $3
```

If no rows are affected, the version has changed since the client last read it, indicating a concurrent edit. The server returns HTTP 409, and the client handles the conflict.

**Server-side Changes:** None to the infrastructure. The existing `PATCH /sessions/:id` endpoint is used as-is. Implementation requires:
1. Adding `version` to the PATCH request body validation
2. Adding the optimistic locking `WHERE version = $expected` clause to the update query
3. Returning HTTP 409 on version mismatch
4. Frontend conflict handling (re-fetch + user notification or auto-merge)

---

## 4. Comparison Matrix

| Criterion | Tiptap + Yjs (Hocuspocus) | Lexical + Yjs | Tiptap (REST-only) |
|---|---|---|---|
| **React 19** | Yes (v3.x) | Yes | Yes (v3.x) |
| **Real-time Collaboration** | CRDT (Yjs) with cursors and presence | CRDT (Yjs) with cursors and presence | Polling + optimistic locking; no cursors |
| **Conflict Resolution Model** | Yjs CRDT (automatic merge) | Yjs CRDT (automatic merge) | Last-write-wins with optimistic locking |
| **Autosave Mechanism** | Hocuspocus `onStoreDocument` hook (debounced) | Yjs provider persist or debounced `registerUpdateListener` | Debounced `onUpdate` to REST PATCH |
| **Autosave within 5s** | Yes — configurable | Yes — configurable | Yes — configurable |
| **JSONB Output** | ProseMirror JSON | Lexical EditorState JSON | ProseMirror JSON |
| **License** | MIT (editor + collab + Hocuspocus) | MIT (all packages) | MIT |
| **Monthly Cost** | $0 (self-hosted) | $0 | $0 |
| **`sessions.version` Integration** | Supplementary counter only | Supplementary counter only | Primary conflict resolution mechanism |
| **Server-side Changes** | Separate Node.js WebSocket server | WebSocket server (Node.js or Go relay) | None — existing REST API |
| **Infrastructure Complexity** | Medium — new Node.js service | Medium — new WebSocket service | Low — no new services |
| **Ecosystem Maturity** | High — ProseMirror lineage, large plugin ecosystem | Medium — newer, growing ecosystem (Meta-backed) | High — same editor, simpler architecture |
| **Learning Curve** | Moderate | Moderate-High (more low-level API) | Low |
| **Upgrade Path** | N/A (already full collaboration) | N/A | Clean upgrade to Candidate A |

---

## 5. Recommendation

**Recommended: Candidate C — Tiptap (REST-only, polling-based collaboration) for v1, with a documented upgrade path to Candidate A (Tiptap + Yjs via Hocuspocus) when scale demands it.**

### Rationale

1. **Right-sized for v1 scale.** With fewer than 100 concurrent users and a maximum of 10 simultaneous editors per document, the probability of truly simultaneous character-by-character editing is low. Most "concurrent" editing in this context means multiple users editing the same session notes between game sessions — not real-time keystroke-level collaboration. Last-write-wins with optimistic locking via the existing `sessions.version` column is explicitly acceptable per the technical constraints and provides adequate conflict protection at this scale.

2. **Zero infrastructure overhead.** No WebSocket server, no Node.js sidecar, no new deployment artifact, no additional health checks or scaling concerns. The existing Go/Echo REST API and `PATCH /sessions/:id` endpoint are sufficient. This directly aligns with the project's stated priority of low storage usage and high performance.

3. **Direct `sessions.version` integration.** The optimistic locking pattern maps directly to the existing `version` column — `UPDATE ... WHERE version = $expected` — making the conflict resolution model transparent, auditable, and simple to implement. No new conflict resolution infrastructure is introduced.

4. **Proven upgrade path.** Tiptap's architecture cleanly separates the editor core from the collaboration layer. Adding `@tiptap/extension-collaboration` + Yjs + Hocuspocus in a future iteration requires no changes to the editor component markup, the stored JSON format, or the frontend editor configuration. The ProseMirror JSON stored in `sessions.notes` is identical regardless of whether collaboration uses REST polling or CRDT sync.

5. **Ecosystem and developer experience.** Tiptap has the most mature React integration, the largest extension ecosystem (text formatting, tables, images, code blocks, etc.), and the best documentation of the three candidates. ProseMirror JSON is a widely adopted, well-understood format with extensive tooling.

6. **Autosave simplicity.** A debounced `onUpdate` callback (2-3 second debounce) calling `PATCH /sessions/:id` with `{ notes: editor.getJSON(), version: currentVersion }` meets the 5-second persistence requirement with no additional infrastructure or configuration complexity.

### Why Not Lexical?

Lexical is a strong editor framework but offers no meaningful advantage over Tiptap for this use case. Its collaboration story still depends on Yjs (same infrastructure requirements as Tiptap + Yjs). Its extension ecosystem is smaller, its API is more low-level (requiring more boilerplate for common editor features), and its EditorState JSON format — while equally JSONB-compatible — is less widely adopted than ProseMirror JSON. Choosing Lexical would increase implementation time without a proportional benefit.

### Why Not Yjs-based Options for v1?

Both Candidates A and B require deploying and maintaining a WebSocket server that does not exist in the current architecture. This adds operational complexity that is disproportionate to v1 scale:

- A new Node.js runtime dependency in an otherwise Go-only backend
- WebSocket connection management, reconnection logic, and health monitoring
- A separate deployment artifact and CI/CD pipeline
- Potential complications with load balancing and proxy configuration for WebSocket upgrades

These concerns are manageable at larger scale where the benefits of real-time CRDT collaboration justify the infrastructure investment. At v1 scale, they represent unnecessary overhead.

> **Post-implementation note (2026):** The v1 REST-only approach was implemented as recommended. The system has since been upgraded to include WebSocket-based operational transform (OT) for real-time collaborative note editing. The OT implementation uses a Go-native WebSocket handler (`backend/handlers/game_ws.go`) and an in-memory document store (`backend/ot/`), rather than the Hocuspocus/Yjs approach described in Candidate A. The upgrade path described in Section 6 was not followed exactly — a Go-native OT solution was chosen instead of a Node.js Hocuspocus sidecar.

---

## 6. Upgrade Path: v1 (REST-only) to v2 (Real-time Collaboration)

When user scale or product requirements demand real-time collaborative editing, the migration from Candidate C to Candidate A is well-defined and non-breaking:

### Migration Steps

1. **Deploy Hocuspocus** as a Node.js sidecar service alongside the Go backend. Hocuspocus handles WebSocket connections, Yjs document synchronization, and persistence.

2. **Add frontend packages:**
   - `@tiptap/extension-collaboration` — Yjs integration for Tiptap
   - `yjs` — CRDT library
   - `@hocuspocus/provider` — WebSocket client connecting to Hocuspocus

3. **Configure Hocuspocus persistence:** The `onStoreDocument` hook serializes the Yjs document to ProseMirror JSON via `yDocToProsemirrorJSON()` and writes it to `sessions.notes` via the existing REST API or direct database access.

4. **`sessions.version` continues** as an application-level counter, incremented on each Hocuspocus persist. Its role changes from conflict resolution mechanism to audit/display counter.

5. **No data migration required.** The ProseMirror JSON format stored in `sessions.notes` is identical whether produced by `editor.getJSON()` (Candidate C) or serialized from a Yjs document (Candidate A). Existing session notes are loaded into the Yjs document on first access.

6. **Frontend editor component changes are minimal:** add the collaboration extension and provider configuration. The editor's toolbar, content schema, and rendering logic remain unchanged.

### Trigger Criteria for Upgrade

Consider upgrading to Candidate A when any of the following apply:

- Users report frequent version conflicts during session note editing
- Product requirements demand live cursor/presence indicators
- Concurrent editor count regularly exceeds 5 per document
- User feedback indicates the polling-based experience is insufficient

---

## 7. Prerequisites

Before implementation begins, the following must be addressed:

1. **Update `database-schema.md` section 5.7** to grant Players read/write access to session notes. The current documentation specifies "Read access" for Players on game-level data (which includes sessions). The intended access model — consistent with shared notes — is that Players have read/write access to session notes.

2. **Add optimistic locking to the session update handler** (`backend/handlers/sessions.go`): the `PATCH /sessions/:id` endpoint must accept a `version` field, apply the `WHERE version = $expected` clause, and return HTTP 409 on mismatch.

3. **Frontend conflict handling**: the Editor view must handle HTTP 409 responses by re-fetching the latest session data and notifying the user of the conflict.

---

## 8. References

### Tiptap
- Documentation: https://tiptap.dev/docs
- GitHub: https://github.com/ueberdosis/tiptap
- React integration: https://tiptap.dev/docs/editor/getting-started/install/react
- Collaboration extension: https://tiptap.dev/docs/editor/extensions/functionality/collaboration
- Hocuspocus: https://tiptap.dev/docs/hocuspocus/introduction
- Pricing (Cloud/Pro): https://tiptap.dev/pricing

### Lexical
- Documentation: https://lexical.dev/docs/intro
- GitHub: https://github.com/facebook/lexical
- React integration: https://lexical.dev/docs/getting-started/react
- Yjs plugin: https://lexical.dev/docs/collaboration/react

### Yjs
- Documentation: https://docs.yjs.dev/
- GitHub: https://github.com/yjs/yjs
- y-websocket: https://github.com/yjs/y-websocket

### ProseMirror
- JSON format: https://prosemirror.net/docs/guide/#doc

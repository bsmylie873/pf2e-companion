# Spike: Campaign Map View — Research & Recommendation

**Date:** 2026-03-25
**Status:** Implemented
**Scope:** Map-based session view — image storage, map rendering, data model, and effort estimate

---

## 1. Introduction & Scope

This document evaluates options for adding a map-based session view to the PF2E Companion application. The map view allows GMs to assign a campaign map image to a game and spatially associate sessions with locations on that image via interactive pins. This provides geographic/narrative context that the current flat session list cannot convey.

### Scope

- **In scope:** Image storage approach, map rendering approach, data model changes, empty/default state UX, and effort estimate for the implementation phase.
- **Out of scope:** Implementation work, real-time collaborative pin editing, accessibility compliance for the interactive pin layer (v1), and zoom/pan functionality (v1).

### System Context

The map view is surfaced as a new route at `/games/:gameId/map`, navigable from the existing Editor page via a view-toggle control. The existing `App.tsx` already defines routes for `/games`, `/games/:gameId`, and `/games/:gameId/sessions/:sessionId/notes`.

Pinning a session means the GM places a marker at an x/y coordinate on the map image. Clicking the marker opens the session detail view. Each pin holds, at minimum, its coordinate position and a reference to the associated session (which provides the session title displayed on the pin).

Access and permissions: any campaign member can create, move, edit, and delete pins. The GM restriction (via the existing `is_gm` role on `GameMembership`) applies only to uploading or replacing the campaign map image itself.

---

## 2. Image Storage — Option Comparison

### Option A: Local Filesystem Storage (Recommended)

Store uploaded images as files on the server's local filesystem (e.g. `./uploads/maps/`), and serve them via a static file route on the Echo server.

**How it works:**
- GM uploads image via `POST /games/:gameId/map-image` (multipart/form-data).
- Backend saves to `./uploads/maps/{gameId}-{hash}.{ext}` (UUID-based filename to avoid collisions).
- A new static route `e.Static("/uploads", "./uploads")` serves files directly.
- The `Game` record stores the relative path (e.g. `/uploads/maps/{gameId}-{hash}.webp`).

**Pros:**
- Zero new infrastructure — no S3 bucket, no CDN, no cloud credentials.
- Matches the self-contained desktop-app goal stated in the project description.
- Echo has built-in `e.Static()` — zero library additions needed.
- Go's `mime/multipart` in the standard library handles file upload parsing.
- Simplest possible implementation path.

**Cons:**
- Not suitable for a horizontally-scaled multi-server deployment (irrelevant for this app's scope — single-user desktop).
- Backup/portability requires copying the `uploads/` directory alongside the database.
- No built-in CDN caching or image transformation.

**Recommended constraints:**
- Max file size: **10 MB** (configurable via Echo's `BodyLimit` middleware).
- Accepted formats: JPEG, PNG, WebP (validated by checking `Content-Type` and/or magic bytes).
- Single map image per game — uploading a new one replaces the previous file.

### Option B: PostgreSQL BYTEA / Large Object Storage

Store the image binary directly in the database as a `BYTEA` column or PostgreSQL large object.

**How it works:**
- Upload parsed in Go, binary stored in a `map_image BYTEA` column on `games` or a dedicated `game_maps` table.
- Served via a handler that reads the blob and writes it to the HTTP response with the appropriate `Content-Type`.

**Pros:**
- Everything in one place — backup the database, get everything.
- No filesystem management.

**Cons:**
- Significantly increases database size and backup times.
- Serving images from the DB is slower than filesystem reads — every request hits the DB connection pool.
- PostgreSQL is not optimised for serving binary blobs at HTTP scale.
- Complicates the existing GORM model layer (BYTEA columns with multi-MB payloads are awkward in ORM structs).
- Conflicts with the project priority of "low storage usage and high performance."

### Option C: Object Storage (S3 / MinIO)

Use an S3-compatible bucket (AWS S3, MinIO self-hosted, etc.).

**Pros:**
- Scalable, CDN-friendly, industry-standard for production web apps.

**Cons:**
- Introduces a new infrastructure dependency (S3 credentials, bucket policies, or a local MinIO container).
- Significantly increases complexity for what is a self-contained desktop app.
- Overkill for single-user / small-party usage patterns.

### Recommendation: Option A — Local Filesystem

For a self-contained desktop application targeting low storage and high performance, local filesystem storage is the clear winner. It adds zero infrastructure, leverages Echo's built-in static file serving, and keeps the implementation minimal. If the app ever needs to scale to a hosted multi-tenant model, migrating to S3 at that point is straightforward — the `map_image_url` field just changes from a local path to an S3 URL.

---

## 3. Map Rendering — Option Comparison

### Option A: Plain CSS/HTML with Percentage-Positioned Pins (Recommended)

Render the map as a standard `<img>` inside a `position: relative` container. Pins are `position: absolute` elements placed using `left` and `top` as percentages.

**How it works:**
```tsx
<div className="map-container" style={{ position: 'relative' }}>
  <img src={mapImageUrl} alt="Campaign map" style={{ width: '100%' }} />
  {pins.map(pin => (
    <button
      key={pin.id}
      className="map-pin"
      style={{ position: 'absolute', left: `${pin.x_pct}%`, top: `${pin.y_pct}%` }}
      onClick={() => navigate(`/games/${gameId}/sessions/${pin.session_id}/notes`)}
    >
      {pin.label}
    </button>
  ))}
</div>
```

- Clicking the image captures click coordinates as percentages: `(e.nativeEvent.offsetX / e.target.width) * 100`.
- Dragging a pin updates its percentage coordinates via a `PATCH` call.
- No zoom/pan needed for v1 — the image scales naturally with the container.

**Pros:**
- Zero new dependencies.
- Trivially simple — percentage positioning inherently solves the responsive display requirement.
- Works with React 19 out of the box.
- Native drag events or pointer events handle pin movement — no library needed.
- Extremely performant — no canvas overhead, no virtual DOM reconciliation issues.
- Easy to style pins with CSS (tooltips, hover states, animations via the existing `motion` library).

**Cons:**
- No built-in zoom/pan. If the map image is very large or detailed, users must rely on browser zoom.
- Pin placement precision is limited by container size (mitigated by storing floats to 2 decimal places).

### Option B: HTML5 Canvas (e.g. Konva / react-konva)

Use a `<canvas>`-based library like react-konva to render the image and interactive pin layers.

**How it works:**
- `<Stage>` and `<Layer>` components render the map image as a canvas image node.
- Pins are `<Circle>` or `<Group>` nodes with drag-and-drop built into Konva.
- Zoom/pan handled via `Stage` scale and position transforms.

**Pros:**
- Built-in zoom/pan/drag with no custom code.
- Performs well with hundreds of interactive elements (not a realistic concern here — campaigns have 10–50 sessions).
- Mature library with good React bindings.

**Cons:**
- Adds a new dependency (`react-konva` + `konva`, ~150 KB gzipped).
- Canvas rendering sits outside React's DOM — styling pins with CSS is not possible; all visual styling must be done imperatively in canvas draw calls.
- Text rendering on canvas is more limited than DOM (no rich typography, no CSS hover states, no native tooltips).
- Overkill for the expected pin count (< 50 per campaign).
- Conflicts with project priority of low storage / high performance (larger bundle).

### Option C: Leaflet / react-leaflet

Use Leaflet with the CRS.Simple coordinate system to render a non-geographic image as a zoomable/pannable map.

**Pros:**
- Purpose-built for interactive maps with markers.
- Excellent zoom/pan UX out of the box.
- Mature ecosystem with React bindings.

**Cons:**
- Adds `leaflet` + `react-leaflet` (~40 KB gzipped) plus Leaflet's CSS.
- CRS.Simple requires coordinate mapping between pixel space and Leaflet's internal LatLng — adds conceptual complexity.
- Marker positioning uses Leaflet's coordinate system, not simple percentages — requires conversion logic.
- Designed for geographic tile maps; using it for a single static image is a non-standard use case.
- Over-engineered for the use case of "one image with < 50 clickable dots."

### Recommendation: Option A — Plain CSS/HTML

The pin count per campaign will realistically be 10–50. There is no zoom/pan requirement for v1. Plain percentage-positioned DOM elements deliver the exact UX needed with zero dependencies, trivial implementation, and maximum styling flexibility. The existing `motion` library can be used for pin entrance animations and hover effects. If zoom/pan becomes a requirement later, a CSS `transform: scale()` wrapper or a migration to Leaflet is straightforward.

---

## 4. Data Model Assessment — `splash_image_url` Reuse vs. New Field

### Current State

The `games` table has:
```sql
splash_image_url    TEXT
```

The Go model:
```go
SplashImageURL *string `gorm:"column:splash_image_url" json:"splash_image_url"`
```

The frontend `Game` type:
```ts
splash_image_url: string | null
```

### Analysis

The `splash_image_url` serves as the campaign's **decorative cover image**, displayed on the `GameCard` component in the games list. The campaign map image is a **functional, interactive surface** that sessions are pinned to. These are semantically and functionally distinct:

| Aspect | `splash_image_url` | Campaign Map Image |
|---|---|---|
| Purpose | Visual card decoration | Interactive session-location canvas |
| Displayed where | GameCard on `/games` | Map view at `/games/:gameId/map` |
| Who sets it | Any member (via Edit Campaign) | GM only |
| Interactivity | None | Click to place/move pins |
| Replacement behaviour | Decorative — can change freely | Changing it may invalidate pin positions |

**Reusing `splash_image_url` is not appropriate** because:
1. A GM may want a splash image (e.g. party portrait) that is different from the campaign map.
2. Changing the splash image is low-stakes; changing the map image invalidates all pin positions.
3. Authorization rules differ — any member can edit splash; only GMs manage the map.

### Recommendation: Add a `map_image_url` column to `games`

This is the simplest option that correctly separates concerns. A dedicated entity (`game_maps` table) is unnecessary — there is a strict 1:1 relationship between a game and its map image, and the only metadata needed is the URL.

**Migration `V1.3__add_campaign_map.sql`:**
```sql
-- Add campaign map image URL to games
ALTER TABLE games ADD COLUMN map_image_url TEXT;

-- Session map pins
CREATE TABLE session_pins (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    session_id  UUID        NOT NULL,
    x_pct       NUMERIC(5,2) NOT NULL,
    y_pct       NUMERIC(5,2) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_session_pins_session
        FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
    CONSTRAINT uq_session_pins_session
        UNIQUE (session_id)
);
```

**Key design decisions:**
- `x_pct` and `y_pct` as `NUMERIC(5,2)` — stores percentage values 0.00–100.00.
- `UNIQUE (session_id)` — each session has at most one pin (1:1).
- `ON DELETE CASCADE` — deleting a session automatically removes its pin (no orphaned pins).
- No `game_id` on `session_pins` — the game is reachable via `sessions.game_id`, avoiding denormalization.

**Go model addition:**
```go
type SessionPin struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    SessionID uuid.UUID `gorm:"type:uuid;not null;column:session_id;uniqueIndex" json:"session_id"`
    XPct      float64   `gorm:"column:x_pct;type:numeric(5,2);not null" json:"x_pct"`
    YPct      float64   `gorm:"column:y_pct;type:numeric(5,2);not null" json:"y_pct"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

**Game model update:**
```go
MapImageURL *string `gorm:"column:map_image_url" json:"map_image_url"`
```

---

## 5. Empty/Default State Recommendation

### Proposed UX Behaviour

| State | Player View | GM View |
|---|---|---|
| No map image uploaded | "Map" nav entry is visible but greyed out with a tooltip: "No campaign map yet" | "Map" nav entry is active, links to `/games/:gameId/map` which shows an upload prompt |
| Map image uploaded, no pins | Full map displayed, read-only (no pin creation affordance beyond their permissions) | Full map displayed with ability to place pins |

### Route Reachability

The `/games/:gameId/map` route should **always exist** (not conditionally rendered). This simplifies routing and avoids conditional `<Route>` logic. The page component itself handles the empty state based on `map_image_url` presence and the user's `is_gm` status.

### Rationale

Greying out the nav entry for players (rather than hiding it) signals that the feature exists, which avoids confusion if a GM mentions the map. Making the route always reachable avoids 404s from shared/bookmarked URLs.

---

## 6. Rough Effort Estimate

### Backend (Go/Echo) — ~12 hours

| Task | Estimate |
|---|---|
| File upload infrastructure — multipart handler, filesystem write, static serving, size/format validation | 3–4 hours |
| Migration V1.3 — `map_image_url` column + `session_pins` table | 0.5 hours |
| Game model update — add `MapImageURL` field, update Game response DTO + frontend type | 0.5 hours |
| SessionPin model + repository — CRUD for pins | 2 hours |
| SessionPin service — business logic, membership auth checks (any member for pins, GM-only for image) | 2 hours |
| SessionPin handler + route registration — REST endpoints for pin CRUD | 2 hours |
| Map image upload/delete handler — `POST/DELETE /games/:gameId/map-image`, GM-only auth | 1.5 hours |

### Frontend (React) — ~13 hours

| Task | Estimate |
|---|---|
| New route + MapView page — `/games/:gameId/map` page component, data fetching | 2 hours |
| Map rendering — image display, percentage-positioned pins, click-to-navigate | 3 hours |
| Pin placement/drag — pointer event handling, PATCH on drop, optimistic UI | 3 hours |
| Image upload UI — GM upload prompt, file picker, upload progress, error handling | 2 hours |
| View toggle in Editor — "List / Map" toggle control in sessions header | 1 hour |
| Empty state handling — greyed-out nav for players, upload prompt for GMs | 1 hour |
| Game type + API client updates — `map_image_url`, pin types, API functions | 1 hour |

### Testing & Integration — ~5 hours

| Task | Estimate |
|---|---|
| Backend unit tests — pin repo/service/handler, upload handler | 3 hours |
| Integration testing — end-to-end upload + pin CRUD flow | 2 hours |

### Total Estimate: ~30 hours (4 working days)

This assumes a single developer familiar with the codebase. The largest risk area is the file upload infrastructure, as it is the only genuinely new capability — all other work follows existing patterns in the codebase (repository → service → handler, Flyway migration, React page + API client).

---

## 7. Implementation Scope Summary (for follow-up ticket)

The follow-up implementation ticket should include:

1. **Flyway migration V1.3** — `map_image_url` on `games`, `session_pins` table with cascade delete.
2. **Backend file upload** — multipart handler, filesystem storage under `./uploads/maps/`, Echo static route, 10 MB limit, JPEG/PNG/WebP validation.
3. **Backend pin CRUD** — `POST/GET/PATCH/DELETE /games/:gameId/pins` (membership-gated), `POST/DELETE /games/:gameId/map-image` (GM-only).
4. **Frontend MapView page** — new route `/games/:gameId/map`, image display, percentage-positioned pins, click-to-place, drag-to-move, click-to-navigate.
5. **Editor view toggle** — "List | Map" control in the sessions section header.
6. **Empty state** — greyed nav for players when no map; upload prompt for GMs.
7. **Model/type updates** — Go `SessionPin` struct, `Game.MapImageURL`, TS `Game` and `SessionPin` types, API client functions.

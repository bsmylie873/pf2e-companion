import { useParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { useMapViewData } from './useMapViewData'
import { BASE_URL } from '../../api/client'
import MapCanvas from './MapCanvas'
import MapOverlayPanel from './MapOverlayPanel'
import PinPickerModal from './PinPickerModal'
import PinGroupModals from './PinGroupModals'
import FolderSidebar from '../../components/FolderSidebar/FolderSidebar'
import EditorModalManager from '../../components/EditorModalManager/EditorModalManager'
import './MapView.css'

export default function MapView() {
  const { gameId } = useParams<{ gameId: string }>()
  const { user } = useAuth()
  const data = useMapViewData(gameId)

  const {
    game, sessions, pins, notes, maps,
    loading, error, isGM, unpinnedSessions,
    activeMapId, viewState, displayScale, setDisplayScale,
    pendingCoords, setPendingCoords, pendingLabel, setPendingLabel,
    pendingDescription, setPendingDescription, pendingColour, setPendingColour,
    pendingIcon, setPendingIcon,
    pinError, setPinError, editingPinId, setEditingPinId,
    pickerSearch, setPickerSearch, editLinkSearch, setEditLinkSearch,
    hoveredPinId, setHoveredPinId, dragging, dropTargetIds,
    pinGroups, activeGroupId, setActiveGroupId,
    managingGroupId, setManagingGroupId,
    groupingPrompt, setGroupingPrompt,
    setPendingGroupPinIds,
    setPendingAddToGroupId,
    dragGroupPrompt, setDragGroupPrompt,
    panelOpen, setPanelOpen, sidebarOpen, uploading, uploadError,
    openItems, setOpenItems, setPins,
    mapContainerRef, viewportContainerRef, fileInputRef, transformRef, wasDragRef,
    handleImageLoad, handleTransformed, handleTransformEnd,
    handleMapClick, handlePointerMove, handlePointerUp, handlePinPointerDown,
    handleSelectSession, handleSelectNote, handleCreateMarker,
    handleDeletePin, handleEditPinField,
    handleUploadClick, handleFileChange, handleDeleteMap,
    handleSidebarToggle, handleSessionUpdate, handleNoteUpdate,
    handleCreateSession, handleCreateNote, openItem,
    sessionForPin, noteForPin, reloadPinGroups,
    handleCreateMap,
    sidebarDragOver, toastMessage, dropLinkedItem, setDropLinkedItem, flashPinId,
    handleCanvasDragOver, handleCanvasDragLeave, handleCanvasDrop,
  } = data

  const activeMap = maps.find(m => m.id === activeMapId)
  const imageUrl = activeMap?.image_url ? `${BASE_URL}${activeMap.image_url}` : ''

  return (
    <div className="map-view-page">
      <div className="map-view-inner">
        {/* Loading state */}
        {loading && (
          <div className="map-spinner">
            <div className="spinner-ring" />
            <p className="spinner-label">Unrolling the map…</p>
          </div>
        )}

        {/* Error state */}
        {!loading && error && (
          <div className="map-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <p>{error}</p>
          </div>
        )}

        {/* Empty: no maps, non-GM */}
        {!loading && !error && maps.length === 0 && !isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">The Map Awaits</p>
            <p className="map-empty-sub">The Game Master has not yet unveiled the realm.</p>
          </div>
        )}

        {/* Empty: map selected, no image, GM */}
        {!loading && !error && activeMapId && !maps.find(m => m.id === activeMapId)?.image_url && isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">No Map Image</p>
            <p className="map-empty-sub">Upload an image for <em>{maps.find(m => m.id === activeMapId)?.name}</em> to begin placing session markers.</p>
            {uploadError && <p className="map-upload-error">{uploadError}</p>}
            <button className="map-upload-btn" onClick={handleUploadClick} disabled={uploading}>
              {uploading ? 'Uploading…' : '+ Upload Map Image'}
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              className="map-file-input"
              onChange={handleFileChange}
            />
          </div>
        )}

        {/* Empty: map selected, no image, non-GM */}
        {!loading && !error && activeMapId && !maps.find(m => m.id === activeMapId)?.image_url && !isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">The Map Awaits</p>
            <p className="map-empty-sub">The Game Master has not yet uploaded an image for this map.</p>
          </div>
        )}

        {/* Empty: no maps, GM — create first map form */}
        {!loading && !error && maps.length === 0 && isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">No Maps Yet</p>
            <p className="map-empty-sub">Name your first map to get started.</p>
            <form className="map-first-create" onSubmit={(e) => {
              e.preventDefault()
              const input = e.currentTarget.querySelector('input')
              const name = input?.value.trim()
              if (name) { handleCreateMap(name); input!.value = '' }
            }}>
              <input
                className="map-first-create-input"
                type="text"
                placeholder="e.g. Otari Region, Dungeon Level 1…"
                autoFocus
                maxLength={255}
              />
              <button className="map-upload-btn" type="submit">
                + Create Map
              </button>
            </form>
          </div>
        )}

        {/* Main map viewport — active map with image */}
        {!loading && !error && activeMapId && activeMap?.image_url && (
          <div
            className={`map-viewport-container${sidebarDragOver ? ' map-viewport-container--drop-active' : ''}`}
            ref={viewportContainerRef}
            onDragOver={handleCanvasDragOver}
            onDragLeave={handleCanvasDragLeave}
            onDrop={handleCanvasDrop}
          >
            {toastMessage && (
              <div className="map-drop-toast">{toastMessage}</div>
            )}
            <MapCanvas
              activeMapId={activeMapId}
              imageUrl={imageUrl}
              viewState={viewState}
              displayScale={displayScale}
              pins={pins}
              pinGroups={pinGroups}
              sessions={sessions}
              notes={notes}
              hoveredPinId={hoveredPinId}
              flashPinId={flashPinId}
              dragging={dragging}
              editingPinId={editingPinId}
              editLinkSearch={editLinkSearch}
              dropTargetIds={dropTargetIds}
              activeGroupId={activeGroupId}
              pinError={pinError}
              mapContainerRef={mapContainerRef}
              viewportContainerRef={viewportContainerRef}
              transformRef={transformRef}
              wasDragRef={wasDragRef}
              onTransformed={handleTransformed}
              onTransformEnd={handleTransformEnd}
              onZoom={(ref) => setDisplayScale(ref.state.scale)}
              onImageLoad={handleImageLoad}
              onMapClick={handleMapClick}
              onPointerMove={handlePointerMove}
              onPointerUp={handlePointerUp}
              onPinPointerDown={handlePinPointerDown}
              onHoverPin={setHoveredPinId}
              onEditPin={(id) => { setEditingPinId(id); setEditLinkSearch('') }}
              onDeletePin={handleDeletePin}
              onEditPinField={handleEditPinField}
              onEditLinkSearchChange={setEditLinkSearch}
              onPinClick={(pin) => {
                if (pin.note_id) openItem('note', pin.note_id, noteForPin(pin)?.title ?? pin.label ?? 'Note')
                else if (pin.session_id) openItem('session', pin.session_id, sessionForPin(pin)?.title ?? 'Session')
                else { setEditingPinId(editingPinId === pin.id ? null : pin.id); setEditLinkSearch('') }
              }}
              onGroupClick={(groupId) => setActiveGroupId(activeGroupId === groupId ? null : groupId)}
              onManageGroup={(groupId) => { setManagingGroupId(groupId); setActiveGroupId(null) }}
              onPinErrorDismiss={() => setPinError(null)}
              openItem={openItem}
              sessionForPin={sessionForPin}
              noteForPin={noteForPin}
              isGM={isGM}
            />

            <MapOverlayPanel
              gameTitle={game?.title ?? 'Campaign Map'}
              isGM={isGM}
              uploading={uploading}
              uploadError={uploadError}
              displayScale={displayScale}
              panelOpen={panelOpen}
              unpinnedSessionsCount={unpinnedSessions.length}
              pinsCount={pins.length}
              transformRef={transformRef}
              fileInputRef={fileInputRef}
              onClose={() => setPanelOpen(false)}
              onOpen={() => setPanelOpen(true)}
              onUploadClick={handleUploadClick}
              onDeleteMap={handleDeleteMap}
              onFileChange={handleFileChange}
            />

            {/* Sidebar toggle + FolderSidebar — inside viewport container */}
            {gameId && (
              <>
                <button
                  className="map-sidebar-toggle-btn"
                  onClick={handleSidebarToggle}
                  title={sidebarOpen ? 'Hide folders' : 'Show folders'}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                    <path d="M3 7h18M3 12h12M3 17h9" />
                  </svg>
                </button>
                {sidebarOpen && (
                  <FolderSidebar
                    gameId={gameId}
                    isGM={isGM}
                    userId={user?.id ?? ''}
                    sessions={sessions}
                    notes={notes}
                    onSessionClick={(id) => { const s = sessions.find(x => x.id === id); openItem('session', id, s?.title ?? 'Session') }}
                    onNoteClick={(id) => { const n = notes.find(x => x.id === id); openItem('note', id, n?.title ?? 'Note') }}
                    onSessionUpdate={handleSessionUpdate}
                    onNoteUpdate={handleNoteUpdate}
                    onCreateSession={handleCreateSession}
                    onCreateNote={handleCreateNote}
                  />
                )}
              </>
            )}
          </div>
        )}
      </div>

      {/* Portalled modals — outside layout flow */}
      {pendingCoords && (
        <PinPickerModal
          pendingColour={pendingColour}
          pendingIcon={pendingIcon}
          pendingLabel={pendingLabel}
          pendingDescription={pendingDescription}
          pickerSearch={pickerSearch}
          unpinnedSessions={unpinnedSessions}
          notes={notes}
          dropLinkedItem={dropLinkedItem}
          onClose={() => { setPendingCoords(null); setPendingLabel(''); setPendingDescription(''); setPickerSearch(''); setDropLinkedItem(null) }}
          onColourChange={setPendingColour}
          onIconChange={setPendingIcon}
          onLabelChange={setPendingLabel}
          onDescriptionChange={setPendingDescription}
          onSearchChange={setPickerSearch}
          onCreateMarker={handleCreateMarker}
          onSelectSession={handleSelectSession}
          onSelectNote={handleSelectNote}
        />
      )}

      <PinGroupModals
        gameId={gameId!}
        activeMapId={activeMapId}
        groupingPrompt={groupingPrompt}
        onDismissGroupingPrompt={() => setGroupingPrompt(null)}
        onPlaceStandalone={(coords) => { setPendingCoords(coords); setGroupingPrompt(null) }}
        onCreateGroupFromPrompt={(coords, pinIds) => { setPendingCoords(coords); setPendingGroupPinIds(pinIds); setGroupingPrompt(null) }}
        onAddToGroupFromPrompt={(coords, groupId) => { setPendingCoords(coords); setPendingAddToGroupId(groupId); setGroupingPrompt(null) }}
        dragGroupPrompt={dragGroupPrompt}
        onDismissDragGroupPrompt={() => setDragGroupPrompt(null)}
        managingGroupId={managingGroupId}
        pinGroups={pinGroups}
        pins={pins}
        sessions={sessions}
        notes={notes}
        onDismissManageGroup={() => setManagingGroupId(null)}
        onReloadPinGroups={reloadPinGroups}
        onUpdatePins={(updater) => setPins(updater)}
      />

      {openItems.length > 0 && (
        <EditorModalManager
          items={openItems}
          gameId={gameId!}
          onClose={(itemId) => setOpenItems(prev => prev.filter(i => i.itemId !== itemId))}
          onCloseAll={() => setOpenItems([])}
        />
      )}
    </div>
  )
}

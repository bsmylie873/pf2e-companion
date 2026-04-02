import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { MapNavProvider } from './context/MapNavContext'
import ProtectedRoute from './components/ProtectedRoute/ProtectedRoute'
import TopBar from './components/TopBar/TopBar'
import Login from './pages/Login/Login'
import GamesList from './pages/GamesList/GamesList'
import Editor from './pages/Editor/Editor'
import SessionNotes from './pages/SessionNotes/SessionNotes'
import NoteEditor from './pages/NoteEditor/NoteEditor'
import Profile from './pages/Profile/Profile'
import MapView from './pages/MapView/MapView'
import './App.css'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <MapNavProvider>
          <div className="app">
              <TopBar />
              <main className="app-content">
                <Routes>
                  <Route path="/" element={<Login />} />
                  <Route path="/games" element={<ProtectedRoute><GamesList /></ProtectedRoute>} />
                  <Route path="/games/:gameId" element={<ProtectedRoute><Editor /></ProtectedRoute>} />
                  <Route path="/games/:gameId/sessions/:sessionId/notes" element={<ProtectedRoute><SessionNotes /></ProtectedRoute>} />
                  <Route path="/games/:gameId/notes/:noteId" element={<ProtectedRoute><NoteEditor /></ProtectedRoute>} />
                  <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
                  <Route path="/games/:gameId/map" element={<ProtectedRoute><MapView /></ProtectedRoute>} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </main>
          </div>
        </MapNavProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}

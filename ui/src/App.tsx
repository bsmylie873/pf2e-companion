import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import TopBar from './components/TopBar/TopBar'
import Login from './pages/Login/Login'
import GamesList from './pages/GamesList/GamesList'
import Editor from './pages/Editor/Editor'
import './App.css'

export default function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <TopBar />
        <main className="app-content">
          <Routes>
            <Route path="/" element={<Login />} />
            <Route path="/games" element={<GamesList />} />
            <Route path="/games/:gameId" element={<Editor />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

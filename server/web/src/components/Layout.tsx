import { useEffect, useState } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import type { User } from '../types/user'

interface LayoutProps {
  children: React.ReactNode
}

function initials(email: string): string {
  const local = email.split('@')[0]
  return local.slice(0, 2).toUpperCase()
}

export default function Layout({ children }: LayoutProps) {
  const [user, setUser] = useState<User | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    fetch('/api/user/me', { credentials: 'include' })
      .then(res => {
        if (res.status === 401) { navigate('/login'); return null }
        return res.json()
      })
      .then(data => data && setUser(data))
      .catch(() => navigate('/login'))
  }, [navigate])

  return (
    <div className="shell">
      <header className="topbar">
        <NavLink to="/" className="topbar-brand">Cierge</NavLink>
        <nav className="topbar-nav">
          <NavLink to="/" className={({ isActive }) => 'topbar-link' + (isActive ? ' active' : '')}>
            Bookings
          </NavLink>
          <NavLink to="/settings" className={({ isActive }) => 'topbar-link' + (isActive ? ' active' : '')}>
            Settings
          </NavLink>
        </nav>
        <div className="topbar-right">
          {user && <div className="topbar-initials">{initials(user.email)}</div>}
        </div>
      </header>
      <main style={{ flex: 1 }}>
        {children}
      </main>
    </div>
  )
}

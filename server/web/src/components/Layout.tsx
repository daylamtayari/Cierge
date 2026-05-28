import { useEffect, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { apiFetch } from '../lib/apiFetch'
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

  useEffect(() => {
    apiFetch('/api/user/me')
      .then(res => res.ok ? res.json() : null)
      .then(data => data && setUser(data))
      .catch(() => {})
  }, [])

  return (
    <div className="shell">
      <header className="topbar">
        <NavLink to="/" className="topbar-brand">Cierge</NavLink>
        <nav className="topbar-nav">
          <NavLink to="/" className={({ isActive }) => 'topbar-link' + (isActive ? ' active' : '')}>
            Bookings
          </NavLink>
          {user?.is_admin && (
            <NavLink to="/admin/bookings" className={({ isActive }) => 'topbar-link' + (isActive ? ' active' : '')}>
              All Bookings
            </NavLink>
          )}
          <NavLink to="/settings" className={({ isActive }) => 'topbar-link' + (isActive ? ' active' : '')}>
            Settings
          </NavLink>
        </nav>
        <div className="topbar-right">
          {user && <div className="topbar-initials">{initials(user.email)}</div>}
        </div>
      </header>
      <main className="flex-1">
        {children}
      </main>
    </div>
  )
}

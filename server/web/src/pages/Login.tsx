import { useState, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'

export default function Login() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    setLoading(true)

    const form = e.currentTarget
    const email = (form.elements.namedItem('email') as HTMLInputElement).value
    const password = (form.elements.namedItem('password') as HTMLInputElement).value

    try {
      const res = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
        credentials: 'include',
      })

      if (res.ok) {
        navigate('/')
      } else if (res.status === 429) {
        setError('Too many attempts. Please try again later.')
      } else {
        setError('Incorrect email or password.')
      }
    } catch {
      setError('Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-shell">
      <div className="login-box">
        <div className="login-brand">
          <h1>Cierge</h1>
          <p>Your reservation, handled.</p>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label className="field-label" htmlFor="email">Email</label>
            <input
              className="field-input"
              id="email"
              name="email"
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              required
            />
          </div>
          <div className="field" style={{ marginBottom: 0 }}>
            <label className="field-label" htmlFor="password">Password</label>
            <input
              className="field-input"
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
            />
          </div>
          <p className="feedback-err" style={{ marginTop: 'var(--sp-3)', visibility: error ? 'visible' : 'hidden' }}>
            {error || '\u00A0'}
          </p>
          <div style={{ paddingTop: 'var(--sp-3)' }}>
            <button
              className="btn btn-primary btn-full"
              type="submit"
              disabled={loading}
            >
              {loading ? 'Signing in…' : 'Sign in'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

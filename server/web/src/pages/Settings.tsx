import { useEffect, useState, FormEvent } from 'react'
import Layout from '../components/Layout'
import { apiFetch } from '../lib/apiFetch'
import type { PlatformToken } from '../types/platform_token'
import type { User } from '../types/user'

function PlatformStatus({ token }: { token: PlatformToken | null }) {
  if (!token) return <span className="platform-status">Not connected</span>
  const expired = token.expires_at && new Date(token.expires_at) < new Date()
  if (expired) return <span className="platform-status platform-status-expired">Expired</span>
  return <span className="platform-status platform-status-connected">Connected</span>
}

export default function Settings() {
  const [user, setUser] = useState<User | null>(null)
  const [tokens, setTokens] = useState<PlatformToken[]>([])

  // Resy connect form
  const [connecting, setConnecting] = useState<string | null>(null)
  const [resyAuthToken, setResyAuthToken] = useState('')
  const [resyRefresh, setResyRefresh] = useState('')
  const [resyApiKey, setResyApiKey] = useState('')
  const [connectError, setConnectError] = useState('')
  const [connectLoading, setConnectLoading] = useState(false)

  // API key
  const [newApiKey, setNewApiKey] = useState<string | null>(null)
  const [apiKeyCopied, setApiKeyCopied] = useState(false)
  const [apiKeyLoading, setApiKeyLoading] = useState(false)

  // Password
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState(false)
  const [passwordLoading, setPasswordLoading] = useState(false)

  useEffect(() => {
    apiFetch('/api/user/me').then(r => r.ok ? r.json() : null).then(d => d && setUser(d))
    apiFetch('/api/user/token').then(r => r.ok ? r.json() : null).then(d => d && setTokens(d))
  }, [])

  function tokenFor(platform: string): PlatformToken | null {
    return tokens.find(t => t.platform === platform) ?? null
  }

  function openConnect(platform: string) {
    setConnecting(platform)
    setConnectError('')
    setResyAuthToken('')
    setResyRefresh('')
    setResyApiKey('')
  }

  async function handleConnectResy(e: FormEvent) {
    e.preventDefault()
    setConnectError('')
    setConnectLoading(true)
    try {
      const body: Record<string, string> = { Token: resyAuthToken, Refresh: resyRefresh }
      if (resyApiKey) body.ApiKey = resyApiKey
      const res = await apiFetch('/api/user/token?platform=resy', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        setConnectError(data.message || 'Failed to connect. Check your token values.')
        return
      }
      const saved: PlatformToken = await res.json()
      setTokens(prev => [...prev.filter(t => t.platform !== 'resy'), saved])
      setConnecting(null)
    } finally {
      setConnectLoading(false)
    }
  }

  async function handleGenerateApiKey() {
    setApiKeyLoading(true)
    setNewApiKey(null)
    setApiKeyCopied(false)
    try {
      const res = await apiFetch('/api/user/api-key', { method: 'POST' })
      if (res.ok) {
        const data = await res.json()
        setNewApiKey(data.api_key)
        setUser(u => u ? { ...u, has_api_key: true } : u)
      }
    } finally {
      setApiKeyLoading(false)
    }
  }

  async function copyApiKey() {
    if (!newApiKey) return
    await navigator.clipboard.writeText(newApiKey)
    setApiKeyCopied(true)
    setTimeout(() => setApiKeyCopied(false), 2000)
  }

  async function handleChangePassword(e: FormEvent) {
    e.preventDefault()
    setPasswordError('')
    setPasswordSuccess(false)
    setPasswordLoading(true)
    try {
      const res = await apiFetch('/api/user/password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
      })
      if (res.ok) {
        setPasswordSuccess(true)
        setOldPassword('')
        setNewPassword('')
      } else {
        const data = await res.json().catch(() => ({}))
        setPasswordError(data.message || 'Failed to update password.')
      }
    } finally {
      setPasswordLoading(false)
    }
  }

  const resyToken = tokenFor('resy')

  return (
    <Layout>
      <div className="container">
        <h1 className="heading-page mb-8">Settings</h1>

        {/* Platform Connections */}
        <section className="section">
          <div className="section-head">
            <h2 className="heading-section">Platform connections</h2>
          </div>

          <div className="platform-block">
            <div className="platform-row">
              <div>
                <div className="platform-name">Resy</div>
                <PlatformStatus token={resyToken} />
              </div>
              <div className="platform-actions">
                {connecting === 'resy' ? (
                  <button className="btn btn-sm btn-subtle" onClick={() => setConnecting(null)}>
                    Cancel
                  </button>
                ) : (
                  <button className="btn btn-sm btn-secondary" onClick={() => openConnect('resy')}>
                    {resyToken ? 'Reconnect' : 'Connect'}
                  </button>
                )}
              </div>
            </div>

            {connecting === 'resy' && (
              <form className="platform-form" onSubmit={handleConnectResy}>
                <div className="field">
                  <label className="field-label" htmlFor="resy-token">Auth token</label>
                  <input
                    className="field-input"
                    id="resy-token"
                    type="password"
                    autoComplete="off"
                    value={resyAuthToken}
                    onChange={e => setResyAuthToken(e.target.value)}
                    required
                  />
                </div>
                <div className="field">
                  <label className="field-label" htmlFor="resy-refresh">Refresh token</label>
                  <input
                    className="field-input"
                    id="resy-refresh"
                    type="password"
                    autoComplete="off"
                    value={resyRefresh}
                    onChange={e => setResyRefresh(e.target.value)}
                    required
                  />
                </div>
                <div className="field">
                  <label className="field-label" htmlFor="resy-apikey">API key <span className="field-label-optional">(optional)</span></label>
                  <input
                    className="field-input"
                    id="resy-apikey"
                    type="text"
                    autoComplete="off"
                    placeholder="Leave blank to use default"
                    value={resyApiKey}
                    onChange={e => setResyApiKey(e.target.value)}
                  />
                  <p className="field-hint">Only set this if you have a custom Resy API key.</p>
                </div>
                {connectError && <p className="feedback-err">{connectError}</p>}
                <button className="btn btn-primary btn-sm" type="submit" disabled={connectLoading}>
                  {connectLoading ? 'Connecting…' : 'Connect Resy'}
                </button>
              </form>
            )}
          </div>

          <div className="platform-block">
            <div className="platform-row">
              <div>
                <div className="platform-name">OpenTable</div>
                <span className="platform-status">Coming soon</span>
              </div>
            </div>
          </div>
        </section>

        <hr className="divider" />

        {/* API Key */}
        <section className="section">
          <div className="section-head">
            <h2 className="heading-section">API key</h2>
          </div>

          {newApiKey ? (
            <div className="mb-4">
              <p className="text-secondary mb-3">
                Copy your API key now — it won't be shown again.
              </p>
              <div className="row-center">
                <code className="code-inline flex-1">{newApiKey}</code>
                <button className="btn btn-sm btn-secondary" onClick={copyApiKey}>
                  {apiKeyCopied ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>
          ) : user?.has_api_key ? (
            <p className="text-secondary mb-4">
              An API key is active. Regenerating will invalidate the current key.
            </p>
          ) : (
            <p className="text-secondary mb-4">
              No API key yet. Generate one to use the Cierge API directly.
            </p>
          )}

          <button
            className="btn btn-secondary"
            onClick={handleGenerateApiKey}
            disabled={apiKeyLoading}
          >
            {apiKeyLoading ? 'Generating…' : user?.has_api_key ? 'Regenerate API key' : 'Generate API key'}
          </button>
        </section>

        <hr className="divider" />

        {/* Password */}
        <section className="section">
          <div className="section-head">
            <h2 className="heading-section">Password</h2>
          </div>
          <form onSubmit={handleChangePassword} className="form-narrow">
            <div className="field">
              <label className="field-label" htmlFor="old-password">Current password</label>
              <input
                className="field-input"
                id="old-password"
                type="password"
                autoComplete="current-password"
                value={oldPassword}
                onChange={e => setOldPassword(e.target.value)}
                required
              />
            </div>
            <div className="field">
              <label className="field-label" htmlFor="new-password">New password</label>
              <input
                className="field-input"
                id="new-password"
                type="password"
                autoComplete="new-password"
                value={newPassword}
                onChange={e => setNewPassword(e.target.value)}
                required
              />
            </div>
            {passwordError && <p className="feedback-err">{passwordError}</p>}
            {passwordSuccess && <p className="feedback-ok">Password updated.</p>}
            <button className="btn btn-primary" type="submit" disabled={passwordLoading}>
              {passwordLoading ? 'Updating…' : 'Update password'}
            </button>
          </form>
        </section>
      </div>
    </Layout>
  )
}

let refreshPromise: Promise<boolean> | null = null

async function tryRefresh(): Promise<boolean> {
  if (refreshPromise) return refreshPromise

  refreshPromise = fetch('/auth/refresh', {
    method: 'POST',
    credentials: 'include',
  })
    .then(res => res.ok)
    .catch(() => false)
    .finally(() => { refreshPromise = null })

  return refreshPromise
}

export async function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const opts: RequestInit = { credentials: 'include', ...init }

  const res = await fetch(input, opts)
  if (res.status !== 401) return res

  const refreshed = await tryRefresh()
  if (!refreshed) {
    window.location.href = '/login'
    return res
  }

  const retry = await fetch(input, opts)
  if (retry.status === 401) {
    window.location.href = '/login'
  }
  return retry
}

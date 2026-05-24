import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import Layout from '../components/Layout'
import type { Job, JobStatus } from '../types/job'
import type { Restaurant } from '../types/restaurant'

function StatusTag({ status }: { status: JobStatus }) {
  const map: Record<JobStatus, string> = {
    created:   'tag tag-scheduled',
    scheduled: 'tag tag-scheduled',
    success:   'tag tag-confirmed',
    failed:    'tag tag-failed',
    cancelled: 'tag tag-cancelled',
  }
  const label: Record<JobStatus, string> = {
    created:   'Scheduled',
    scheduled: 'Scheduled',
    success:   'Confirmed',
    failed:    'Failed',
    cancelled: 'Cancelled',
  }
  return <span className={map[status]}>{label[status]}</span>
}

function formatDate(dateStr: string): string {
  return new Date(dateStr + 'T00:00:00').toLocaleDateString('en-US', {
    weekday: 'long', month: 'long', day: 'numeric', year: 'numeric',
  })
}

function formatDateTime(dateStr: string): string {
  const d = new Date(dateStr)
  const time = d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  const date = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  return `${time} on ${date}`
}

export default function Booking() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [job, setJob] = useState<Job | null>(null)
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch(`/api/job/${id}`, { credentials: 'include' })
      .then(res => {
        if (res.status === 401) { navigate('/login'); return null }
        if (res.status === 404) { navigate('/'); return null }
        return res.json()
      })
      .then(async (data: Job | null) => {
        if (!data) return
        setJob(data)
        const r = await fetch(`/api/restaurant/${data.restaurant_id}`, { credentials: 'include' })
        if (r.ok) setRestaurant(await r.json())
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [id, navigate])

  if (loading) {
    return (
      <Layout>
        <div className="container">
          <div className="empty-state"><p>Loading…</p></div>
        </div>
      </Layout>
    )
  }

  if (!job) return null

  return (
    <Layout>
      <div className="container">
        <Link to="/" className="back-link">← Bookings</Link>

        <div className="detail-header">
          <div>
            <h1>{restaurant?.name ?? 'Booking'}</h1>
            {restaurant?.city && (
              <p className="text-secondary" style={{ marginTop: 'var(--sp-1)' }}>
                {[restaurant.city, restaurant.state].filter(Boolean).join(', ')}
              </p>
            )}
          </div>
          <StatusTag status={job.status} />
        </div>

        <div className="section">
          <dl className="details">
            <dt>Date</dt>
            <dd>{formatDate(job.reservation_date)}</dd>

            <dt>Party size</dt>
            <dd>{job.party_size}</dd>

            <dt>Preferred times</dt>
            <dd>{job.preferred_times.join(', ')}</dd>

            <dt>Platform</dt>
            <dd style={{ textTransform: 'capitalize' }}>{job.platform}</dd>

            <dt>Runs at</dt>
            <dd>{formatDateTime(job.scheduled_at)}</dd>

            {job.reserved_time && (
              <>
                <dt>Reserved time</dt>
                <dd>{formatDateTime(job.reserved_time)}</dd>
              </>
            )}

            {job.confirmation && (
              <>
                <dt>Confirmation</dt>
                <dd>{job.confirmation}</dd>
              </>
            )}
          </dl>
        </div>

        {job.status === 'success' && job.confirmation && (
          <div className="banner banner-confirmed" style={{ marginBottom: 'var(--sp-6)' }}>
            <div className="banner-label">Reservation confirmed</div>
            <div className="banner-value">{job.confirmation}</div>
            {job.reserved_time && (
              <div className="banner-detail">{formatDateTime(job.reserved_time)}</div>
            )}
          </div>
        )}

        {job.status === 'failed' && (
          <div className="banner banner-failed" style={{ marginBottom: 'var(--sp-6)' }}>
            <div className="banner-label">Booking failed</div>
            {job.error_message && (
              <div className="banner-value" style={{ marginTop: 'var(--sp-1)' }}>
                {job.error_message}
              </div>
            )}
          </div>
        )}

        {job.logs && (
          <div className="section">
            <div className="section-head">
              <h2 className="heading-section">Logs</h2>
            </div>
            <pre className="log-viewer">{job.logs}</pre>
          </div>
        )}
      </div>
    </Layout>
  )
}

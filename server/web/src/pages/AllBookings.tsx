import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Layout from '../components/Layout'
import { apiFetch } from '../lib/apiFetch'
import type { Job, JobStatus } from '../types/job'
import type { Restaurant } from '../types/restaurant'

function StatusTag({ status }: { status: JobStatus }) {
  const cls: Record<JobStatus, string> = {
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
  return <span className={cls[status]}>{label[status]}</span>
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  })
}

function formatDateTime(dateStr: string): string {
  const d = new Date(dateStr)
  const time = d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  const date = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  return `${time} on ${date}`
}

export default function AllBookings() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [restaurants, setRestaurants] = useState<Record<string, Restaurant>>({})
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    apiFetch('/api/user/me')
      .then(res => res.ok ? res.json() : null)
      .then(user => {
        if (!user?.is_admin) { navigate('/'); return }
        return apiFetch('/api/admin/job/list')
      })
      .then(res => {
        if (!res || !res.ok) return null
        return res.json()
      })
      .then(async (data: Job[] | null) => {
        if (!data) return
        setJobs([...data].sort((a, b) =>
          new Date(b.scheduled_at).getTime() - new Date(a.scheduled_at).getTime()
        ))

        const uniqueIds = [...new Set(data.map(j => j.restaurant_id))]
        const results = await Promise.all(
          uniqueIds.map(id =>
            apiFetch(`/api/restaurant/${id}`)
              .then(r => r.ok ? r.json() as Promise<Restaurant> : null)
              .catch(() => null)
          )
        )
        const map: Record<string, Restaurant> = {}
        results.forEach(r => { if (r) map[r.id] = r })
        setRestaurants(map)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [navigate])

  return (
    <Layout>
      <div className="container container--wide">
        <div className="page-header">
          <h1 className="heading-page">All Bookings</h1>
        </div>

        {loading ? (
          <div className="empty-state"><p>Loading…</p></div>
        ) : jobs.length === 0 ? (
          <div className="empty-state"><p>No bookings yet.</p></div>
        ) : (
          <div className="job-table-wrap">
            <table className="job-table">
              <thead>
                <tr>
                  <th>Restaurant</th>
                  <th>User</th>
                  <th>Date</th>
                  <th>Party</th>
                  <th>Preferred times</th>
                  <th>Runs at</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {jobs.map(job => (
                  <tr key={job.id} className="job-table-row" onClick={() => navigate(`/booking/${job.id}`, { state: { from: '/admin/bookings', label: 'All Bookings' } })}>
                    <td className="job-table-platform">{restaurants[job.restaurant_id]?.name ?? '—'}</td>
                    <td className="job-table-user">{job.user_id.slice(0, 8)}</td>
                    <td>{formatDate(job.reservation_date + 'T00:00:00')}</td>
                    <td>{job.party_size}</td>
                    <td>{job.preferred_times.join(', ')}</td>
                    <td>{formatDateTime(job.scheduled_at)}</td>
                    <td><StatusTag status={job.status} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </Layout>
  )
}

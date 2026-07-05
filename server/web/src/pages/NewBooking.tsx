import { useCallback, useEffect, useRef, useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import Layout from '../components/Layout'
import { apiFetch } from '../lib/apiFetch'
import type { Venue } from '../types/venue'
import type { Restaurant } from '../types/restaurant'
import type { DropConfig } from '../types/drop_config'
import type { Job } from '../types/job'
import type { PlatformToken } from '../types/platform_token'

type Step = 1 | 2 | 3

// Search is re-run at most once a second, and only when the query has changed
// since the last search — mirroring the CLI's live venue search.
const SEARCH_INTERVAL_MS = 1000

// 15-minute slots, ordered so dinner times (from 18:00) appear first, wrapping
// back around to the early hours — matches the CLI time slot picker ordering.
function buildSlots(): { value: string; label: string }[] {
    const slots: { value: string; label: string }[] = []
    for (let i = 0; i < 96; i++) {
        const idx = (72 + i) % 96
        const hour = Math.floor(idx / 4)
        const minute = (idx % 4) * 15
        const value = `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
        slots.push({ value, label: formatTime(value) })
    }
    return slots
}

const ALL_SLOTS = buildSlots()

// Formats an HH:mm 24-hour string into a friendly 12-hour label (e.g. 6:00 PM).
function formatTime(hhmm: string): string {
    const [h, m] = hhmm.split(':').map(Number)
    const period = h < 12 ? 'AM' : 'PM'
    const hour = h % 12 === 0 ? 12 : h % 12
    return `${hour}:${String(m).padStart(2, '0')} ${period}`
}

function todayISO(): string {
    const now = new Date()
    const local = new Date(now.getTime() - now.getTimezoneOffset() * 60000)
    return local.toISOString().slice(0, 10)
}

// Formats an ISO YYYY-MM-DD string as DD/MM/YYYY (used for our own displays;
// the native date input itself follows the browser locale).
function formatDMY(iso: string): string {
    const [y, m, d] = iso.split('-')
    return `${d}/${m}/${y}`
}

// Short, friendly date for the confirmation page, e.g. "Jul 6, 2026".
function formatShort(iso: string): string {
    return new Date(iso + 'T00:00:00').toLocaleDateString('en-US', {
        month: 'short', day: 'numeric', year: 'numeric',
    })
}

// A date + time rendered in the user's own local time zone, e.g. "Jul 6, 2026 at 9:00 AM".
function formatLocalDateTime(d: Date): string {
    const date = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    const time = d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
    return `${date} at ${time}`
}

// The calendar date reservations open: the reservation date minus the lead time.
// Whole-day subtraction is timezone-independent, so this is safe to do locally.
function openDateISO(reservationISO: string, daysInAdvance: number): string | null {
    if (!reservationISO || !Number.isFinite(daysInAdvance)) return null
    const d = new Date(reservationISO + 'T00:00:00')
    d.setDate(d.getDate() - daysInAdvance)
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    return `${yyyy}-${mm}-${dd}`
}

// Resolves the absolute instant of a wall-clock date + time in a given IANA time
// zone (falling back to local time). Used to check whether reservations open in
// the future. Uses the standard offset-diff technique; good enough for a warning.
function dropInstant(dateISO: string, timeHHmm: string, timeZone?: string): Date {
    const [y, mo, d] = dateISO.split('-').map(Number)
    const [h, mi] = timeHHmm.split(':').map(Number)
    if (!timeZone) return new Date(y, mo - 1, d, h, mi)
    const utcGuess = Date.UTC(y, mo - 1, d, h, mi)
    const base = new Date(utcGuess)
    const tzWall = new Date(base.toLocaleString('en-US', { timeZone })).getTime()
    const utcWall = new Date(base.toLocaleString('en-US', { timeZone: 'UTC' })).getTime()
    return new Date(utcGuess + (utcWall - tzWall))
}

function locationLabel(parts: (string | undefined)[]): string {
    return parts.filter(Boolean).join(', ')
}

const ChevronUp = () => (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor"
        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M4 10l4-4 4 4" />
    </svg>
)
const ChevronDown = () => (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor"
        strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M4 6l4 4 4-4" />
    </svg>
)
const TrashIcon = () => (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor"
        strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M3 4.5h10M6.5 4.5V3h3v1.5M5 4.5l.5 8.5h5l.5-8.5" />
    </svg>
)

export default function NewBooking() {
    const navigate = useNavigate()
    const [step, setStep] = useState<Step>(1)

    // Step 1 — platform + restaurant
    const platform = 'resy'
    const [tokens, setTokens] = useState<PlatformToken[]>([])
    const [query, setQuery] = useState('')
    const [venues, setVenues] = useState<Venue[]>([])
    const [searching, setSearching] = useState(false)
    const [searchError, setSearchError] = useState('')
    const [restaurant, setRestaurant] = useState<Restaurant | null>(null)
    const [loadingRestaurant, setLoadingRestaurant] = useState(false)
    const [restaurantError, setRestaurantError] = useState('')

    // Step 2 — details
    const [partySize, setPartySize] = useState('2')
    const [reservationDate, setReservationDate] = useState('')
    const [times, setTimes] = useState<string[]>([])
    const [slotToAdd, setSlotToAdd] = useState(ALL_SLOTS[0].value)
    const [dropConfigs, setDropConfigs] = useState<DropConfig[]>([])
    const [dropConfigsLoading, setDropConfigsLoading] = useState(false)
    const [selectedDropConfigId, setSelectedDropConfigId] = useState<string>('') // config id or 'new'
    const [newDaysInAdvance, setNewDaysInAdvance] = useState('')
    const [newDropTime, setNewDropTime] = useState('09:00')

    // Step 3 — submit
    const [submitting, setSubmitting] = useState(false)
    const [submitError, setSubmitError] = useState('')

    // --- Platform connection -------------------------------------------------
    // The selected platform can only be booked if the user has a live token for
    // it. Mirrors the Settings page: an expired token counts as not connected.
    useEffect(() => {
        apiFetch('/api/user/token')
            .then(r => (r.ok ? r.json() : null))
            .then(d => Array.isArray(d) && setTokens(d))
            .catch(() => { /* leave tokens empty; banner will prompt to connect */ })
    }, [])

    const platformToken = tokens.find(t => t.platform === platform) ?? null
    const platformConnected =
        !!platformToken && !(platformToken.expires_at && new Date(platformToken.expires_at) < new Date())

    // --- Live search ---------------------------------------------------------
    // The interval reads the latest query via refs so it never closes over stale
    // state. lastSearchedRef tracks what we last searched for; seqRef discards
    // responses that arrive out of order.
    const queryRef = useRef(query)
    queryRef.current = query
    const lastSearchedRef = useRef('')
    const seqRef = useRef(0)

    const runSearch = useCallback(async (raw: string) => {
        const q = raw.trim()
        lastSearchedRef.current = q
        const seq = ++seqRef.current

        if (q === '') {
            setVenues([])
            setSearching(false)
            setSearchError('')
            return
        }

        setSearching(true)
        setSearchError('')
        try {
            const res = await apiFetch('/proxy/resy/restaurant', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: q }),
            })
            if (seq !== seqRef.current) return // a newer search superseded this one
            if (res.ok) {
                const data = await res.json()
                setVenues(Array.isArray(data) ? data : [])
            } else {
                setSearchError('Could not search restaurants. Please try again.')
            }
        } catch {
            if (seq === seqRef.current) setSearchError('Could not search restaurants. Please try again.')
        } finally {
            if (seq === seqRef.current) setSearching(false)
        }
    }, [])

    useEffect(() => {
        const interval = setInterval(() => {
            const q = queryRef.current.trim()
            if (q !== lastSearchedRef.current) runSearch(q)
        }, SEARCH_INTERVAL_MS)
        return () => clearInterval(interval)
    }, [runSearch])

    function handleSearchSubmit(e: FormEvent) {
        e.preventDefault()
        runSearch(query)
    }

    // --- Restaurant selection ------------------------------------------------
    async function selectVenue(v: Venue) {
        setRestaurantError('')
        setLoadingRestaurant(true)
        try {
            const res = await apiFetch(`/api/restaurant?platform=${platform}&platform-id=${v.id.resy}`)
            if (res.ok) {
                setRestaurant(await res.json())
            } else {
                setRestaurantError('Could not load this restaurant. Please try another.')
            }
        } catch {
            setRestaurantError('Could not load this restaurant. Please try another.')
        } finally {
            setLoadingRestaurant(false)
        }
    }

    function changeRestaurant() {
        setRestaurant(null)
        setRestaurantError('')
    }

    // --- Step transitions ----------------------------------------------------
    async function goToDetails() {
        if (!restaurant) return
        setStep(2)
        setDropConfigsLoading(true)
        try {
            const res = await apiFetch(`/api/drop-config?restaurant=${restaurant.id}`)
            const data = res.ok ? await res.json() : []
            const list: DropConfig[] = Array.isArray(data) ? data : []
            setDropConfigs(list)
            setSelectedDropConfigId(list.length ? list[0].id : 'new')
        } catch {
            setDropConfigs([])
            setSelectedDropConfigId('new')
        } finally {
            setDropConfigsLoading(false)
        }
    }

    // --- Time slots ----------------------------------------------------------
    function addTime() {
        if (!slotToAdd || times.includes(slotToAdd)) return
        setTimes(prev => [...prev, slotToAdd])
    }
    function moveTime(index: number, dir: -1 | 1) {
        const target = index + dir
        if (target < 0 || target >= times.length) return
        setTimes(prev => {
            const next = [...prev]
                ;[next[index], next[target]] = [next[target], next[index]]
            return next
        })
    }
    function removeTime(index: number) {
        setTimes(prev => prev.filter((_, i) => i !== index))
    }

    // --- Validation ----------------------------------------------------------
    const partySizeNum = Number(partySize)
    const partySizeValid = Number.isInteger(partySizeNum) && partySizeNum >= 1
    const reservationISO = reservationDate || null
    const dateValid = !!reservationISO && reservationISO >= todayISO()
    const daysNum = Number(newDaysInAdvance)
    const newScheduleValid =
        Number.isInteger(daysNum) && daysNum >= 1 && /^\d{2}:\d{2}$/.test(newDropTime)
    const dropConfigValid =
        selectedDropConfigId === 'new' ? newScheduleValid : selectedDropConfigId !== ''

    // Resolved release schedule (whichever source is selected).
    const chosenSchedule =
        selectedDropConfigId === 'new'
            ? (newScheduleValid ? { days_in_advance: daysNum, drop_time: newDropTime } : null)
            : dropConfigs.find(dc => dc.id === selectedDropConfigId) ?? null

    // The date reservations would open for the chosen date + schedule, and whether
    // that release moment has already passed (in the restaurant's local time).
    const openISO =
        chosenSchedule && reservationISO
            ? openDateISO(reservationISO, chosenSchedule.days_in_advance)
            : null
    const releaseInPast =
        !!(chosenSchedule && openISO) &&
        dropInstant(openISO, chosenSchedule.drop_time, restaurant?.timezone).getTime() <= Date.now()

    // The absolute moment the booking runs (interpreted in the restaurant's zone),
    // shown on the confirmation page in the user's own local time.
    const runsAtInstant =
        chosenSchedule && openISO
            ? dropInstant(openISO, chosenSchedule.drop_time, restaurant?.timezone)
            : null

    const detailsValid =
        partySizeValid && dateValid && times.length >= 1 && dropConfigValid && !releaseInPast

    // --- Submission ----------------------------------------------------------
    async function handleSubmit() {
        if (!restaurant || !reservationISO || !detailsValid) return
        setSubmitting(true)
        setSubmitError('')
        try {
            let dropConfigId = selectedDropConfigId
            if (dropConfigId === 'new') {
                const res = await apiFetch('/api/drop-config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        Restaurant: restaurant.id,
                        DaysInAdvance: daysNum,
                        DropTime: newDropTime,
                    }),
                })
                if (!res.ok) {
                    const data = await res.json().catch(() => ({}))
                    setSubmitError(data.message || 'Could not save the release schedule.')
                    return
                }
                const created: DropConfig = await res.json()
                dropConfigId = created.id
            }

            const res = await apiFetch('/api/job', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    restaurant_id: restaurant.id,
                    reservation_date: reservationISO,
                    party_size: partySizeNum,
                    preferred_times: times,
                    drop_config_id: dropConfigId,
                }),
            })
            if (res.ok) {
                const job: Job = await res.json()
                navigate(`/booking/${job.id}`)
            } else {
                const data = await res.json().catch(() => ({}))
                setSubmitError(data.message || 'Could not schedule this booking. Please try again.')
            }
        } catch {
            setSubmitError('Could not schedule this booking. Please try again.')
        } finally {
            setSubmitting(false)
        }
    }

    return (
        <Layout>
            <div className="container">
                {step === 1 ? (
                    <Link to="/" className="back-link">← Bookings</Link>
                ) : (
                    <button className="back-link" onClick={() => setStep((step - 1) as Step)}>
                        ← Back
                    </button>
                )}

                <p className="step-indicator">Step {step} of 3</p>
                <h1 className="heading-page mb-8">
                    {step === 1 ? 'Choose a restaurant' : step === 2 ? 'Booking details' : 'Confirm booking'}
                </h1>

                {/* ===== Step 1: platform + restaurant ===== */}
                {step === 1 && (
                    <>
                        {!platformConnected && (
                            <div className="notice-warn mb-6" role="alert">
                                <span className="notice-warn-icon" aria-hidden="true">⚠</span>
                                <span>
                                    Your {platform === 'resy' ? 'Resy' : platform} account isn't connected
                                    {platformToken ? ' — its connection has expired' : ''}. Connect it in{' '}
                                    <Link to="/settings">Settings</Link> before scheduling a booking.
                                </span>
                            </div>
                        )}

                        <section className="section">
                            <div className="section-head">
                                <h2 className="heading-section">Platform</h2>
                            </div>
                            <div className="option-list">
                                <div className="option" aria-selected="true">
                                    <span className="option-label">Resy</span>
                                </div>
                                <div className="option" aria-disabled="true">
                                    <span className="option-label">OpenTable</span>
                                    <span className="option-meta option-soon">Coming soon</span>
                                </div>
                            </div>
                        </section>

                        <section className="section">
                            <div className="section-head">
                                <h2 className="heading-section">Restaurant</h2>
                            </div>

                            {restaurant ? (
                                <div className="selected-restaurant">
                                    <div>
                                        <div className="name">{restaurant.name}</div>
                                        {locationLabel([restaurant.city, restaurant.state]) && (
                                            <div className="meta">{locationLabel([restaurant.city, restaurant.state])}</div>
                                        )}
                                    </div>
                                    <button className="btn btn-sm btn-subtle" onClick={changeRestaurant}>
                                        Change
                                    </button>
                                </div>
                            ) : (
                                <>
                                    <div className="field mb-4">
                                        <label className="field-label" htmlFor="restaurant-search">
                                            Search by name
                                        </label>
                                        <form className="search-bar" onSubmit={handleSearchSubmit}>
                                            <input
                                                className="field-input"
                                                id="restaurant-search"
                                                type="text"
                                                placeholder="e.g. Carbone"
                                                autoComplete="off"
                                                value={query}
                                                onChange={e => setQuery(e.target.value)}
                                            />
                                            <button className="btn btn-secondary" type="submit">Search</button>
                                        </form>
                                    </div>

                                    {restaurantError && <p className="feedback-err mb-3">{restaurantError}</p>}
                                    {searchError && <p className="feedback-err mb-3">{searchError}</p>}

                                    {loadingRestaurant ? (
                                        <p className="text-secondary">Loading restaurant…</p>
                                    ) : venues.length > 0 ? (
                                        <div className="option-list">
                                            {venues.map(v => (
                                                <button
                                                    key={v.id.resy}
                                                    className="option"
                                                    onClick={() => selectVenue(v)}
                                                >
                                                    <span className="option-label">{v.name}</span>
                                                    <span className="option-meta">
                                                        {locationLabel([v.neighborhood || v.locality, v.region])}
                                                    </span>
                                                </button>
                                            ))}
                                        </div>
                                    ) : query.trim() && !searching ? (
                                        <p className="text-secondary">No restaurants found.</p>
                                    ) : searching ? (
                                        <p className="text-secondary">Searching…</p>
                                    ) : null}
                                </>
                            )}
                        </section>

                        <button
                            className="btn btn-primary"
                            onClick={goToDetails}
                            disabled={!restaurant || !platformConnected}
                        >
                            Continue
                        </button>
                    </>
                )}

                {/* ===== Step 2: details ===== */}
                {step === 2 && (
                    <>
                        <div className="field-row">
                            <div className="field">
                                <label className="field-label" htmlFor="party-size">Guests</label>
                                <input
                                    className="field-input"
                                    id="party-size"
                                    type="number"
                                    min="1"
                                    value={partySize}
                                    onChange={e => setPartySize(e.target.value)}
                                />
                            </div>
                            <div className="field">
                                <label className="field-label" htmlFor="reservation-date">Reservation date</label>
                                <input
                                    className="field-input"
                                    id="reservation-date"
                                    type="date"
                                    min={todayISO()}
                                    value={reservationDate}
                                    onChange={e => setReservationDate(e.target.value)}
                                />
                            </div>
                        </div>

                        <section className="section">
                            <div className="section-head">
                                <h2 className="heading-section">Preferred times</h2>
                            </div>
                            <p className="field-hint mb-3">
                                Add one or more times in order of preference. The first will be tried first.
                            </p>
                            <div className="search-bar mb-3">
                                <select
                                    className="field-input"
                                    aria-label="Time to add"
                                    value={slotToAdd}
                                    onChange={e => setSlotToAdd(e.target.value)}
                                >
                                    {ALL_SLOTS.map(s => (
                                        <option key={s.value} value={s.value}>{s.label}</option>
                                    ))}
                                </select>
                                <button className="btn btn-secondary" type="button" onClick={addTime}>
                                    Add time
                                </button>
                            </div>

                            {times.length > 0 ? (
                                <ol className="slot-list">
                                    {times.map((t, i) => (
                                        <li className="slot-row" key={t}>
                                            <span className="slot-rank">{i + 1}</span>
                                            <span className="slot-time">{formatTime(t)}</span>
                                            <div className="slot-actions">
                                                <button
                                                    className="btn btn-subtle btn-sm btn-icon"
                                                    onClick={() => moveTime(i, -1)}
                                                    disabled={i === 0}
                                                    aria-label={`Move ${formatTime(t)} up`}
                                                    title="Move up"
                                                >
                                                    <ChevronUp />
                                                </button>
                                                <button
                                                    className="btn btn-subtle btn-sm btn-icon"
                                                    onClick={() => moveTime(i, 1)}
                                                    disabled={i === times.length - 1}
                                                    aria-label={`Move ${formatTime(t)} down`}
                                                    title="Move down"
                                                >
                                                    <ChevronDown />
                                                </button>
                                                <button
                                                    className="btn btn-subtle btn-sm btn-icon"
                                                    onClick={() => removeTime(i)}
                                                    aria-label={`Remove ${formatTime(t)}`}
                                                    title="Remove"
                                                >
                                                    <TrashIcon />
                                                </button>
                                            </div>
                                        </li>
                                    ))}
                                </ol>
                            ) : (
                                <p className="text-secondary">No times added yet.</p>
                            )}
                        </section>

                        <section className="section">
                            <div className="section-head">
                                <h2 className="heading-section">When do reservations open?</h2>
                            </div>

                            {dropConfigsLoading ? (
                                <p className="text-secondary">Loading schedules…</p>
                            ) : (
                                <div className="option-list">
                                    {dropConfigs.map(dc => {
                                        const opens = reservationISO ? openDateISO(reservationISO, dc.days_in_advance) : null
                                        return (
                                            <button
                                                key={dc.id}
                                                className="option"
                                                aria-selected={selectedDropConfigId === dc.id}
                                                onClick={() => setSelectedDropConfigId(dc.id)}
                                            >
                                                <span className="option-label">
                                                    {dc.days_in_advance} days before · {formatTime(dc.drop_time)}
                                                </span>
                                                <span className="option-meta">
                                                    {opens ? `Opens ${formatDMY(opens)} · ` : ''}
                                                    Used {dc.confidence} {dc.confidence === 1 ? 'time' : 'times'}
                                                </span>
                                            </button>
                                        )
                                    })}
                                    <button
                                        className="option"
                                        aria-selected={selectedDropConfigId === 'new'}
                                        onClick={() => setSelectedDropConfigId('new')}
                                    >
                                        <span className="option-label">Create a new release schedule</span>
                                    </button>
                                </div>
                            )}

                            {selectedDropConfigId === 'new' && (
                                <div className="field-row mt-3">
                                    <div className="field">
                                        <label className="field-label" htmlFor="days-in-advance">Days in advance</label>
                                        <input
                                            className="field-input"
                                            id="days-in-advance"
                                            type="number"
                                            min="1"
                                            placeholder="e.g. 14"
                                            value={newDaysInAdvance}
                                            onChange={e => setNewDaysInAdvance(e.target.value)}
                                        />
                                        <p className="field-hint">How many days before the date reservations open.</p>
                                    </div>
                                    <div className="field">
                                        <label className="field-label" htmlFor="drop-time">Release time</label>
                                        <input
                                            className="field-input"
                                            id="drop-time"
                                            type="time"
                                            step="900"
                                            value={newDropTime}
                                            onChange={e => setNewDropTime(e.target.value)}
                                        />
                                        <p className="field-hint">Restaurant's local time.</p>
                                    </div>
                                </div>
                            )}
                        </section>

                        {releaseInPast && openISO && chosenSchedule && (
                            <div className="notice-warn mb-4" role="alert">
                                <span className="notice-warn-icon" aria-hidden="true">⚠</span>
                                <span>
                                    Reservations for this date and schedule would have opened on{' '}
                                    {formatDMY(openISO)} at {formatTime(chosenSchedule.drop_time)} — that has already
                                    passed. Pick an earlier release schedule or a later reservation date.
                                </span>
                            </div>
                        )}
                        <button className="btn btn-primary" onClick={() => setStep(3)} disabled={!detailsValid}>
                            Continue
                        </button>
                    </>
                )}

                {/* ===== Step 3: confirm ===== */}
                {step === 3 && restaurant && (
                    <>
                        <div className="summary mb-6">
                            <dl className="details">
                                <dt>Restaurant</dt>
                                <dd>{restaurant.name}</dd>

                                {locationLabel([restaurant.city, restaurant.state]) && (
                                    <>
                                        <dt>Location</dt>
                                        <dd>{locationLabel([restaurant.city, restaurant.state])}</dd>
                                    </>
                                )}

                                <dt>Date</dt>
                                <dd>{reservationISO ? formatShort(reservationISO) : ''}</dd>

                                <dt>Guests</dt>
                                <dd>{partySizeNum}</dd>

                                <dt>Requested times</dt>
                                <dd>{times.map(formatTime).join(', ')}</dd>

                                <dt>Platform</dt>
                                <dd className="text-capitalize">{platform}</dd>

                                {runsAtInstant && (
                                    <>
                                        <dt>Runs at</dt>
                                        <dd>{formatLocalDateTime(runsAtInstant)}</dd>
                                    </>
                                )}
                            </dl>
                        </div>

                        {submitError && <p className="feedback-err mb-3">{submitError}</p>}

                        <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting}>
                            {submitting ? 'Scheduling…' : 'Schedule this booking'}
                        </button>
                    </>
                )}
            </div>
        </Layout>
    )
}

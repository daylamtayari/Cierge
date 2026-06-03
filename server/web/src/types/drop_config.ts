// A release schedule for a restaurant — when reservations open relative to the
// reservation date. Returned by /api/drop-config, ordered by confidence desc.
export interface DropConfig {
  id: string
  days_in_advance: number
  drop_time: string // HH:mm, restaurant-local
  confidence: number
  created_at: string
}

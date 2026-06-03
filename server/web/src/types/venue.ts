// Search result from POST /proxy/resy/restaurant. The Resy API returns many
// more fields, but the booking flow only needs these. See resy/venue.go.
export interface Venue {
  id: { resy: number; google: string }
  name: string
  locality: string
  region: string
  neighborhood: string
}

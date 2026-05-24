export interface PlatformToken {
  id: string
  user_id: string
  platform: string
  expires_at?: string
  has_refresh: boolean
  refresh_expires_at?: string
  created_at: string
}

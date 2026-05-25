export type JobStatus = 'created' | 'scheduled' | 'success' | 'failed' | 'cancelled'

export interface Job {
  id: string
  user_id: string
  restaurant_id: string
  platform: string
  reservation_date: string
  party_size: number
  preferred_times: string[]
  scheduled_at: string
  drop_config_id: string
  callbacked: boolean
  status: JobStatus
  started_at?: string
  completed_at?: string
  reserved_time?: string
  confirmation?: string
  error_message?: string
  logs?: string
  created_at: string
  updated_at: string
}

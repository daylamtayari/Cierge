export interface User {
  id: string
  email: string
  has_api_key: boolean
  is_admin: boolean
  auth_method: string
  last_login_at?: string
  created_at: string
}

export type UserStatus = 'subscribed' | 'unsubscribed' | 'bounced'
export type SaleStatus = 'paid' | 'failed' | 'refunded'

export interface Mail {
  id: number
  unread?: boolean
  from: User
  subject: string
  body: string
  date: string
}

export interface Member {
  name: string
  username: string
  role: 'member' | 'owner'
  avatar: Avatar
}

export interface Stat {
  title: string
  icon: string
  value: number | string
  variation: number
  formatter?: (value: number) => string
}

export interface Sale {
  id: string
  date: string
  status: SaleStatus
  email: string
  amount: number
}

export interface Notification {
  id: number
  unread?: boolean
  sender: User
  body: string
  date: string
}

export type Period = 'daily' | 'weekly' | 'monthly'

export interface Range {
  start: Date
  end: Date
}

export interface Project {
  id: number
  name: string
  created_at: string
  updated_at: string
}

export interface Inbox {
  id: number
  project_id: number
  email: string
  created_at: string
  updated_at: string
}

export interface Messages {
  id: number
  inbox_id: number
  sender: string
  receiver: string
  subject: string
  body: string
  is_read: boolean
  created_at: string
  updated_at: string
}

export interface ApiResponse<T> {
  data: T[]
  pagination: {
    total: number
    limit: number
    offset: number
  }
}

export interface User {
  id: string
  email: string
  username: string
  name: string
  role: string
}

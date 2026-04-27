export interface WebhookEvent {
  ID: string
  TunnelID: string
  ReceivedAt: string
  Method: string
  Path: string
  Headers: string
  RequestBody: string
  ResponseStatus: number
  ResponseBody: string
  ResponseMS: number
  Forwarded: boolean
  ReplayedAt: string | null
}

export interface Tunnel {
  ID: string
  Type: 'personal' | 'org'
  Subdomain: string
  Status: string
  ActiveUserID: string
}

export type Role = 'admin' | 'member'

export interface User {
  ID: string
  OrgID: string
  Email: string
  Name: string
  APIKey: string
  Role: Role
}

export interface ConfirmState {
  message: string
  detail?: string
  onConfirm: () => void
}

export interface Org {
  ID: string
  Name: string
  CreatedAt: string
}

export interface Me {
  id: string
  email: string
  name: string
  role: string
  org_id: string
}

export interface TableInfo {
  name: string
  row_count: number
}

export interface TableResult {
  columns: string[]
  rows: unknown[][]
}

export interface QueryResult {
  columns: string[]
  rows: unknown[][]
  affected: number
}

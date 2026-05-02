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
  DisplayName: string
  Status: 'active' | 'inactive'
  ActiveUserID: string
  ActiveDevice: string
}

export type RoleName = string

export interface User {
  ID: string
  OrgID: string
  Email: string
  Name: string
  APIKey: string
  Role: RoleName
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
  org_name: string
  api_key: string
  permissions: string[]
}

export interface OrgRole {
  name: string
  display_name: string
  permissions: string[]
  is_system: boolean
  created_at: string
}

export interface OrgMember {
  ID: string
  Name: string
  Email: string
  Role: string
  ActiveTunnelSubdomain: string
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
}

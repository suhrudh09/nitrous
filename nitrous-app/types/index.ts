// ── Event ─────────────────────────────────────────────────────────────────────

export interface Event {
  id: string
  title: string
  location: string
  date: string
  time?: string
  isLive: boolean
  category: EventCategory
  thumbnailUrl?: string
  createdAt?: string
}

export type EventCategory = 'motorsport' | 'water' | 'air' | 'offroad'

// ── Category ──────────────────────────────────────────────────────────────────

export interface Category {
  id: string
  name: string
  slug: EventCategory
  icon: string
  liveCount: number
  description: string
  color: CategoryColor
}

export type CategoryColor = 'cyan' | 'blue' | 'purple' | 'orange'

// ── Journey ───────────────────────────────────────────────────────────────────

export interface Journey {
  id: string
  title: string
  category: string
  description: string
  badge: 'EXCLUSIVE' | 'MEMBERS ONLY' | 'LIMITED'
  slotsLeft: number
  date: string
  price: number
  thumbnailUrl?: string
}

// ── MerchItem ─────────────────────────────────────────────────────────────────

export interface MerchItem {
  id: string
  name: string
  icon: string
  price: number
  category: 'apparel' | 'accessories' | 'collectibles'
}

// ── Stream ────────────────────────────────────────────────────────────────────

export interface Stream {
  id: string
  eventId: string
  title: string
  subtitle: string      // e.g. "Lap 87 / 200"
  category: string
  location: string
  quality: string       // "4K" | "HD"
  viewers: number
  isLive: boolean
  currentLeader: string
  currentSpeed: string
  color: string
  createdAt?: string
}

// Matches the StreamTelemetry struct broadcast over WebSocket
export interface StreamTelemetry {
  streamId: string
  viewers: number
  currentLeader: string
  currentSpeed: string
  subtitle: string
}

// ── Team ──────────────────────────────────────────────────────────────────────

export interface Team {
  id: string
  name: string
  category: string      // e.g. "MOTORSPORT · F1"
  country: string
  founded: number
  rank: number
  wins: number
  points: number
  following: number
  drivers: string[]
  color: string
  createdAt?: string
}

// ── Hero nav card ─────────────────────────────────────────────────────────────

export interface NavCard {
  id: string
  label: string
  icon: string
  href: string
  color: 'grey' | 'red' | 'cyan' | 'orange' | 'blue' | 'gold'
  progress: number
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export interface User {
  id: string
  email: string
  name: string
  createdAt: string
}

export interface AuthResponse {
  user: User
  token: string
}

// ── Order ─────────────────────────────────────────────────────────────────────

export interface OrderItem {
  merchId: string
  name: string
  price: number
  quantity: number
  size?: string
}

export interface Order {
  id: string
  userId: string
  items: OrderItem[]
  total: number
  status: 'pending' | 'confirmed' | 'shipped'
  createdAt: string
}
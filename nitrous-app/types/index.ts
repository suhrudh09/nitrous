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
  playbackUrl?: string
  externalWatch?: {
    platform: string
    label: string
    url: string
  }[]
  date_start?: string
  date_end?: string
  country_name?: string
  circuit_short_name?: string
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

export interface OpenF1RecentSession {
  session_key: number
  session_name: string
  date_start: string
  date_end: string
  country_name: string
  circuit_short_name: string
}

export interface OpenF1SessionTelemetry {
  session_key: number
  current_leader: string
  speed_kph: number
  rpm: number
  gear: number
  g_force: number
  captured_at: string
}

export interface OpenF1VideoResult {
  videoId: string
  title: string
  channelTitle: string
  embedUrl: string
  watchUrl: string
  query: string
  mode: 'live' | 'recent'
  sessionKey: number
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
  role: 'viewer' | 'participant' | 'manager' | 'sponsor' | 'admin'
  plan: 'FREE' | 'VIP' | 'PLATINUM'
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
  status: 'pending' | 'confirmed' | 'shipped' | 'failed' | 'cancelled'
  createdAt: string
}

export interface CartItem {
  merchId: string
  name: string
  icon: string
  price: number
  category: 'apparel' | 'accessories' | 'collectibles' | string
  quantity: number
  size?: string
}

export interface PaymentIntentResponse {
  clientSecret: string
  paymentId: string
  amount: number
  currency: string
}

// ── JourneyBooking ────────────────────────────────────────────────────────────

export interface JourneyBooking {
  id: string
  title: string
  category: string
  description: string
  badge: string
  slotsLeft: number
  date: string
  price: number
  thumbnailUrl?: string
  bookingId: string
  quantity: number
  bookedAt: string
}

export interface PaymentState {
  status: 'idle' | 'processing' | 'success' | 'failed'
  paymentId?: string
  error?: string
}

export interface Reminder {
  id: string
  userId: string
  eventId: string
  message?: string
  remindAt: string
  createdAt: string
}

export interface Notification {
  id: string
  userId: string
  title: string
  body: string
  type: string
  readAt?: string
  createdAt: string
}
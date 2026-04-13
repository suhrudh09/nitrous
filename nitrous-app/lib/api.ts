import type {
  Event,
  Category,
  Journey,
  MerchItem,
  Stream,
  OpenF1RecentSession,
  OpenF1SessionTelemetry,
  Team,
  User,
  AuthResponse,
  Order,
  OrderItem,
} from '@/types'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

// ── Generic fetch wrapper ──────────────────────────────────────────────────────

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_URL}${endpoint}`

  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (!res.ok) {
      const error = await res.json()
      throw new Error(error.error || 'API request failed')
    }

    return res.json()
  } catch (error) {
    console.error(`API Error [${endpoint}]:`, error)
    throw error
  }
}

// ── Events ────────────────────────────────────────────────────────────────────

export async function getEvents(category?: string): Promise<Event[]> {
  const query = category ? `?category=${encodeURIComponent(category)}` : ''
  const data = await fetchAPI<{ events: Event[]; count: number }>(`/events${query}`)
  return data.events ?? []
}

export async function getLiveEvents(): Promise<Event[]> {
  const data = await fetchAPI<{ events: Event[]; count: number }>('/events/live')
  return data.events ?? []
}

export async function getEventById(id: string): Promise<Event> {
  return fetchAPI<Event>(`/events/${id}`)
}

export async function setReminder(
  eventId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/events/${eventId}/remind`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function deleteReminder(
  eventId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/events/${eventId}/remind`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Categories ────────────────────────────────────────────────────────────────

export async function getCategories(): Promise<Category[]> {
  const data = await fetchAPI<{ categories: Category[]; count: number }>('/categories')
  return data.categories ?? []
}

export async function getCategoryBySlug(slug: string): Promise<Category> {
  return fetchAPI<Category>(`/categories/${slug}`)
}

// ── Journeys ──────────────────────────────────────────────────────────────────

export async function getJourneys(): Promise<Journey[]> {
  const data = await fetchAPI<{ journeys: Journey[]; count: number }>('/journeys')
  return data.journeys ?? []
}

export async function getJourneyById(id: string): Promise<Journey> {
  return fetchAPI<Journey>(`/journeys/${id}`)
}

export async function bookJourney(
  id: string,
  token: string
): Promise<{ message: string; journey: Journey }> {
  return fetchAPI<{ message: string; journey: Journey }>(`/journeys/${id}/book`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Merch ─────────────────────────────────────────────────────────────────────

export async function getMerchItems(): Promise<MerchItem[]> {
  const data = await fetchAPI<{ items: MerchItem[]; count: number }>('/merch')
  return data.items ?? []
}

export async function getMerchItemById(id: string): Promise<MerchItem> {
  return fetchAPI<MerchItem>(`/merch/${id}`)
}

// ── Streams ───────────────────────────────────────────────────────────────────

export async function getStreams(): Promise<Stream[]> {
  const data = await fetchAPI<{ streams: Stream[]; count: number }>('/streams')
  return data.streams ?? []
}

export async function getStreamById(id: string): Promise<Stream> {
  return fetchAPI<Stream>(`/streams/${id}`)
}

export async function getOpenF1RecentSessions(limit = 8): Promise<OpenF1RecentSession[]> {
  const data = await fetchAPI<{ sessions: OpenF1RecentSession[]; count: number }>(
    `/streams/openf1/sessions?limit=${limit}`
  )
  return data.sessions ?? []
}

export async function getOpenF1SessionTelemetry(
  sessionKey: number
): Promise<OpenF1SessionTelemetry> {
  return fetchAPI<OpenF1SessionTelemetry>(`/streams/openf1/sessions/${sessionKey}/telemetry`)
}

// ── Teams ─────────────────────────────────────────────────────────────────────

export async function getTeams(): Promise<Team[]> {
  const data = await fetchAPI<{ teams: Team[]; count: number }>('/teams')
  return data.teams ?? []
}

export async function getTeamById(id: string): Promise<Team> {
  return fetchAPI<Team>(`/teams/${id}`)
}

export async function followTeam(
  id: string,
  token: string
): Promise<{ message: string; following: number }> {
  return fetchAPI<{ message: string; following: number }>(`/teams/${id}/follow`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function unfollowTeam(
  id: string,
  token: string
): Promise<{ message: string; following: number }> {
  return fetchAPI<{ message: string; following: number }>(`/teams/${id}/follow`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Orders ────────────────────────────────────────────────────────────────────

export async function createOrder(
  items: OrderItem[],
  token: string
): Promise<{ message: string; order: Order }> {
  return fetchAPI<{ message: string; order: Order }>('/orders', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ items }),
  })
}

export async function getMyOrders(token: string): Promise<Order[]> {
  const data = await fetchAPI<{ orders: Order[]; count: number }>('/orders', {
    headers: { Authorization: `Bearer ${token}` },
  })
  return data.orders ?? []
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export async function register(
  email: string,
  password: string,
  name: string
): Promise<AuthResponse> {
  return fetchAPI<AuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, name }),
  })
}

export async function login(
  email: string,
  password: string
): Promise<AuthResponse> {
  return fetchAPI<AuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function getCurrentUser(token: string): Promise<User> {
  return fetchAPI<User>('/auth/me', {
    headers: { Authorization: `Bearer ${token}` },
  })
}
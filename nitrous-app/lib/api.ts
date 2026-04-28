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

    const contentType = res.headers.get('content-type') || ''
    const isJSON = contentType.includes('application/json')

    if (!res.ok) {
      if (isJSON) {
        const error = await res.json()
        throw new Error(error.error || 'API request failed')
      }

      const raw = await res.text()
      const bodyPreview = raw.trim().slice(0, 120)
      throw new Error(bodyPreview || `Request failed (${res.status})`)
    }

    if (isJSON) {
      return res.json()
    }

    const raw = await res.text()
    throw new Error(`Expected JSON response from ${endpoint}, received: ${raw.trim().slice(0, 120)}`)
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
  token: string,
  userId?: string
): Promise<{ message: string; journey: Journey }> {
  const body = userId ? { userId } : {}
  return fetchAPI<{ message: string; journey: Journey }>(`/journeys/${id}/book`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
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

export async function createTeam(
  data: { name: string; country: string; isPrivate: boolean; drivers?: string[] },
  token: string
): Promise<Team> {
  return fetchAPI<Team>('/teams', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  })
}

export async function updateTeam(
  id: string,
  data: { name?: string; country?: string; isPrivate?: boolean; drivers?: string[] },
  token: string
): Promise<Team> {
  return fetchAPI<Team>(`/teams/${id}`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(data),
  })
}

export async function deleteTeam(id: string, token: string): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function addTeamManager(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/managers`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ userId }),
  })
}

export async function removeTeamManager(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/managers/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function addTeamMember(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/members`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ userId }),
  })
}

export async function removeTeamMember(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/members/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function addTeamSponsor(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/sponsors`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ userId }),
  })
}

export async function removeTeamSponsor(
  teamId: string,
  userId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/teams/${teamId}/sponsors/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Admin ─────────────────────────────────────────────────────────────────────

export async function triggerSync(token: string): Promise<{ message: string; results: Record<string, { success: boolean; error?: string }> }> {
  return fetchAPI<{ message: string; results: Record<string, { success: boolean; error?: string }> }>('/admin/sync', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function updateUserRole(
  userId: string,
  role: string,
  token: string
): Promise<{ message: string; user: User }> {
  return fetchAPI<{ message: string; user: User }>(`/admin/users/${userId}/role`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ role }),
  })
}

// ── Orders ────────────────────────────────────────────────────────────────────

export async function createOrder(
  items: OrderItem[],
  token: string
): Promise<{ message: string; order: Order }> {
  const payload = {
    merchItemIds: items.map((item) => item.merchId),
    quantities: items.map((item) => item.quantity),
    unitPrices: items.map((item) => item.price),
  }

  const response = await fetchAPI<any>('/orders', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })

  // Backward/forward compatibility: some handlers return raw order, others wrap it.
  const rawOrder = response?.order ?? response
  if (!rawOrder || typeof rawOrder.id !== 'string') {
    throw new Error('Invalid order response from server')
  }

  const normalizedOrder: Order = {
    id: rawOrder.id,
    userId: rawOrder.userId,
    items: Array.isArray(rawOrder.items) ? rawOrder.items : [],
    total: typeof rawOrder.total === 'number' ? rawOrder.total : 0,
    status: rawOrder.status,
    createdAt: rawOrder.createdAt,
  }

  return {
    message: typeof response?.message === 'string' ? response.message : 'Order created',
    order: normalizedOrder,
  }
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
  name: string,
  role: string
): Promise<AuthResponse> {
  return fetchAPI<AuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, name, role }),
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
export async function purchasePass(
  id: string,
  token: string
): Promise<{ message: string; passId: string }> {
  return fetchAPI<{ message: string; passId: string }>(`/passes/${id}/purchase`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}
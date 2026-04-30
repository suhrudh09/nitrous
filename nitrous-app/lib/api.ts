import type {
  Event,
  Category,
  Journey,
  MerchItem,
  Stream,
  OpenF1RecentSession,
  OpenF1SessionTelemetry,
  OpenF1VideoResult,
  Team,
  User,
  AuthResponse,
  Order,
  OrderItem,
  CartItem,
  PaymentIntentResponse,
  Reminder,
  Notification,
} from '@/types'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

// ── Pass interfaces ───────────────────────────────────────────────────────────
export interface AvailablePass {
  id: string
  tier: string
  event: string
  location: string
  date: string
  category: string
  price: number
  perks: string[]
  spotsLeft: number
  totalSpots: number
  badge?: string | null
  tierColor: string
}

export interface UserPass {
  purchaseId: string
  createdAt: string
  id: string
  tier: string
  event: string
  location: string
  date: string
  category: string
  price: number
  perks: string[]
  spotsLeft: number
  totalSpots: number
  badge?: string
  tierColor: string
}

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
  input: {
    eventId: string
    remindAt: string
    message?: string
  },
  token: string
): Promise<Reminder> {
  return fetchAPI<Reminder>('/reminders', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(input),
  })
}

export async function deleteReminder(
  reminderId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/reminders/${reminderId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getMyReminders(token: string): Promise<Reminder[]> {
  const data = await fetchAPI<{ reminders: Reminder[]; count: number }>('/reminders', {
    headers: { Authorization: `Bearer ${token}` },
  })
  return data.reminders ?? []
}

export async function getNotifications(
  token: string
): Promise<{ notifications: Notification[]; unread: number }> {
  const data = await fetchAPI<{ notifications: Notification[]; count: number; unread: number }>('/notifications', {
    headers: { Authorization: `Bearer ${token}` },
  })

  return {
    notifications: data.notifications ?? [],
    unread: data.unread ?? 0,
  }
}

export async function markNotificationRead(
  notificationId: string,
  token: string
): Promise<{ message: string; readAt: string }> {
  return fetchAPI<{ message: string; readAt: string }>(`/notifications/${notificationId}/read`, {
    method: 'POST',
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
  userId?: string,
  quantity = 1
): Promise<{ message: string; journey: Journey }> {
  const body: Record<string, unknown> = { quantity }
  if (userId) body.userId = userId
  return fetchAPI<{ message: string; journey: Journey }>(`/journeys/${id}/book`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  })
}

export async function getMyJourneyBookings(token: string): Promise<import('@/types').JourneyBooking[]> {
  const data = await fetchAPI<{ bookings: import('@/types').JourneyBooking[]; count: number }>('/journeys/my', {
    headers: { Authorization: `Bearer ${token}` },
  })
  return data.bookings ?? []
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

export async function getOpenF1SessionVideo(params: {
  mode: 'live' | 'recent'
  sessionKey?: number
}): Promise<OpenF1VideoResult> {
  const query = new URLSearchParams({ mode: params.mode })
  if (typeof params.sessionKey === 'number') {
    query.set('sessionKey', String(params.sessionKey))
  }
  return fetchAPI<OpenF1VideoResult>(`/streams/openf1/video?${query.toString()}`)
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
): Promise<{ message: string; team: Team }> {
  return fetchAPI<{ message: string; team: Team }>(`/teams/${id}/follow`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function unfollowTeam(
  id: string,
  token: string
): Promise<{ message: string; team: Team }> {
  return fetchAPI<{ message: string; team: Team }>(`/teams/${id}/unfollow`, {
    method: 'POST',
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
  const data = await fetchAPI<{ results: Record<string, string> }>('/admin/sync', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })

  const normalizedResults: Record<string, { success: boolean; error?: string }> = {}
  for (const [provider, value] of Object.entries(data.results ?? {})) {
    if (value === 'ok') {
      normalizedResults[provider] = { success: true }
    } else {
      normalizedResults[provider] = { success: false, error: value || 'Failed' }
    }
  }

  return {
    message: 'Sync completed',
    results: normalizedResults,
  }
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
  const data = await fetchAPI<{ orders: any[]; count: number }>('/orders', {
    headers: { Authorization: `Bearer ${token}` },
  })

  return (data.orders ?? []).map((rawOrder) => {
    const merchIds = Array.isArray(rawOrder?.merchItemIds) ? rawOrder.merchItemIds : []
    const quantities = Array.isArray(rawOrder?.quantities) ? rawOrder.quantities : []
    const unitPrices = Array.isArray(rawOrder?.unitPrices) ? rawOrder.unitPrices : []

    const items: OrderItem[] = Array.isArray(rawOrder?.items)
      ? rawOrder.items
      : merchIds.map((merchId: string, index: number) => ({
          merchId,
          name: '',
          price: typeof unitPrices[index] === 'number' ? unitPrices[index] : 0,
          quantity: typeof quantities[index] === 'number' ? quantities[index] : 0,
        }))

    return {
      id: rawOrder.id,
      userId: rawOrder.userId,
      items,
      total: typeof rawOrder.total === 'number' ? rawOrder.total : 0,
      status: rawOrder.status,
      createdAt: rawOrder.createdAt,
    }
  })
}

export async function getAvailablePasses(): Promise<AvailablePass[]> {
  const data = await fetchAPI<{ passes: AvailablePass[]; count: number }>('/passes/catalog')
  return (data.passes ?? []).map((pass) => ({
    ...pass,
    perks: Array.isArray(pass.perks) ? pass.perks : [],
    badge: pass.badge ?? undefined,
  }))
}

export async function getMyPasses(token: string): Promise<UserPass[]> {
  const data = await fetchAPI<{
    purchases: Array<{
      purchaseId: string
      createdAt: string
      pass: {
        id: string
        tier: string
        event: string
        location: string
        date: string
        category: string
        price: number
        perks?: string[] | null
        spotsLeft: number
        totalSpots: number
        badge?: string | null
        tierColor: string
      }
    }>
    count: number
  }>('/passes', {
    headers: { Authorization: `Bearer ${token}` },
  })

  return (data.purchases ?? []).map((purchase) => ({
    purchaseId: purchase.purchaseId,
    createdAt: purchase.createdAt,
    id: purchase.pass.id,
    tier: purchase.pass.tier,
    event: purchase.pass.event,
    location: purchase.pass.location,
    date: purchase.pass.date,
    category: purchase.pass.category,
    price: purchase.pass.price,
    perks: Array.isArray(purchase.pass.perks) ? purchase.pass.perks : [],
    spotsLeft: purchase.pass.spotsLeft,
    totalSpots: purchase.pass.totalSpots,
    badge: purchase.pass.badge ?? undefined,
    tierColor: purchase.pass.tierColor,
  }))
}

// ── Cart ──────────────────────────────────────────────────────────────────────

export async function getCart(token: string): Promise<CartItem[]> {
  const data = await fetchAPI<{ items: CartItem[]; count: number }>('/cart', {
    headers: { Authorization: `Bearer ${token}` },
  })
  return data.items ?? []
}

export async function saveCart(items: CartItem[], token: string): Promise<CartItem[]> {
  const data = await fetchAPI<{ items: CartItem[]; count: number }>('/cart', {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ items }),
  })
  return data.items ?? []
}

export async function clearCart(token: string): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>('/cart', {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ── Payments ──────────────────────────────────────────────────────────────────

export async function createPaymentIntent(
  amount: number,
  referenceType: string,
  referenceId: string,
  token: string,
  currency = 'usd'
): Promise<PaymentIntentResponse> {
  return fetchAPI<PaymentIntentResponse>('/payments/create-intent', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ amount, currency, referenceType, referenceId }),
  })
}

export async function confirmPayment(
  paymentId: string,
  token: string
): Promise<{ message: string }> {
  return fetchAPI<{ message: string }>(`/payments/${paymentId}/confirm`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
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

export async function updateCurrentUserPlan(
  plan: 'FREE' | 'VIP' | 'PLATINUM',
  token: string
): Promise<User> {
  const data = await fetchAPI<{ user: User; message: string }>('/auth/me/plan', {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ plan }),
  })
  return data.user
}

export async function updateCurrentUserRole(
  role: 'viewer' | 'participant' | 'manager' | 'sponsor',
  token: string
): Promise<User> {
  const data = await fetchAPI<{ user: User; message: string }>('/auth/me/role', {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ role }),
  })
  return data.user
}
export async function purchasePass(
  id: string,
  token: string,
  quantity = 1
): Promise<{ message: string; passId: string; quantity: number }> {
  return fetchAPI<{ message: string; passId: string; quantity: number }>(`/passes/${id}/purchase`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ quantity }),
  })
}

export async function getOrderById(orderId: string, token: string): Promise<Order> {
  const rawOrder = await fetchAPI<any>(`/orders/${orderId}`, {
    headers: { Authorization: `Bearer ${token}` },
  })

  const merchIds = Array.isArray(rawOrder?.merchItemIds) ? rawOrder.merchItemIds : []
  const quantities = Array.isArray(rawOrder?.quantities) ? rawOrder.quantities : []
  const unitPrices = Array.isArray(rawOrder?.unitPrices) ? rawOrder.unitPrices : []

  const items: OrderItem[] = Array.isArray(rawOrder?.items)
    ? rawOrder.items
    : merchIds.map((merchId: string, index: number) => ({
        merchId,
        name: '',
        price: typeof unitPrices[index] === 'number' ? unitPrices[index] : 0,
        quantity: typeof quantities[index] === 'number' ? quantities[index] : 0,
      }))

  return {
    id: rawOrder.id,
    userId: rawOrder.userId,
    items,
    total: typeof rawOrder.total === 'number' ? rawOrder.total : 0,
    status: rawOrder.status,
    createdAt: rawOrder.createdAt,
  }
}

export async function cancelOrder(orderId: string, token: string): Promise<Order> {
  const rawOrder = await fetchAPI<any>(`/orders/${orderId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })

  const merchIds = Array.isArray(rawOrder?.merchItemIds) ? rawOrder.merchItemIds : []
  const quantities = Array.isArray(rawOrder?.quantities) ? rawOrder.quantities : []
  const unitPrices = Array.isArray(rawOrder?.unitPrices) ? rawOrder.unitPrices : []

  const items: OrderItem[] = Array.isArray(rawOrder?.items)
    ? rawOrder.items
    : merchIds.map((merchId: string, index: number) => ({
        merchId,
        name: '',
        price: typeof unitPrices[index] === 'number' ? unitPrices[index] : 0,
        quantity: typeof quantities[index] === 'number' ? quantities[index] : 0,
      }))

  return {
    id: rawOrder.id,
    userId: rawOrder.userId,
    items,
    total: typeof rawOrder.total === 'number' ? rawOrder.total : 0,
    status: rawOrder.status,
    createdAt: rawOrder.createdAt,
  }
}
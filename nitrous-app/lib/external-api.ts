// lib/external-api.ts
//
// Typed client functions for all open-source external APIs integrated
// into the Nitrous Go backend. Matches the exact pattern of lib/api.ts —
// same fetchAPI wrapper, same error handling, same NEXT_PUBLIC_API_URL base.
//
// Import alongside your existing api.ts:
//   import { getF1Session, getAircraftAtAirshow, getRacingCircuits } from '@/lib/external-api'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_URL}${endpoint}`
  try {
    const res = await fetch(url, {
      ...options,
      headers: { 'Content-Type': 'application/json', ...options?.headers },
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

// ═══════════════════════════════════════════════════════════════════════════════
// TYPES
// ═══════════════════════════════════════════════════════════════════════════════

// ── OpenF1 types ──────────────────────────────────────────────────────────────

export interface F1Session {
  session_key: number
  session_name: string       // "Race" | "Qualifying" | "Sprint" | "Practice 1" | …
  circuit_short_name: string
  date_start: string         // ISO-8601
}

export interface F1Position {
  driver_number: number
  position: number
  date: string
}

export interface F1Driver {
  driver_number: number
  full_name: string
  name_acronym: string       // e.g. "VER", "HAM", "NOR"
  team_name: string
  team_colour: string        // hex without #, e.g. "3671C6"
  country_code: string
  headshot_url: string
}

export interface F1Lap {
  driver_number: number
  lap_number: number
  lap_duration: number | null  // seconds
  duration_sector_1: number | null
  duration_sector_2: number | null
  duration_sector_3: number | null
  is_pit_out_lap: boolean
  date_start: string
}

export interface F1Weather {
  air_temperature: number    // °C
  track_temperature: number  // °C
  humidity: number           // %
  wind_speed: number         // m/s
  wind_direction: number     // degrees
  rainfall: boolean
  date: string
}

export interface F1RaceControlMessage {
  category: string           // "Flag" | "SafetyCar" | "Drs" | …
  message: string
  flag: string               // "GREEN" | "YELLOW" | "RED" | "CHEQUERED" | "SC" | "VSC"
  date: string
}

export interface F1LivePayload {
  stream_id: string
  current_leader: string
  current_speed: number      // km/h
  subtitle: string
  session_key: number
  session_name: string
  circuit_name: string
  positions: F1Position[]
  top_speed_kmh: number
  leader_lap: number
  flag: string
  weather: F1Weather | null
  drivers: F1Driver[]
  recent_pits: F1Pit[]
  recent_laps: F1Lap[]
}

export interface F1Pit {
  driver_number: number
  lap_number: number
  pit_duration: number       // seconds
  date: string
}

// Jolpica returns Ergast-shaped MRData — we type the most common shapes here.
// Full reference: https://github.com/jolpica/jolpica-f1/blob/main/docs/README.md

export interface JolpicaDriver {
  driverId: string
  permanentNumber: string
  code: string               // "VER", "HAM", …
  givenName: string
  familyName: string
  nationality: string
}

export interface JolpicaConstructor {
  constructorId: string
  name: string
  nationality: string
}

export interface DriverStanding {
  position: string
  positionText: string
  points: string
  wins: string
  Driver: JolpicaDriver
  Constructors: JolpicaConstructor[]
}

export interface ConstructorStanding {
  position: string
  positionText: string
  points: string
  wins: string
  Constructor: JolpicaConstructor
}

export interface F1Race {
  season: string
  round: string
  raceName: string
  date: string
  time: string
  Circuit: {
    circuitId: string
    circuitName: string
    Location: {
      lat: string
      long: string
      locality: string
      country: string
    }
  }
}

export interface F1RaceResult {
  number: string
  position: string
  positionText: string
  points: string
  Driver: JolpicaDriver
  Constructor: JolpicaConstructor
  grid: string
  laps: string
  status: string
  Time?: { millis: string; time: string }
  FastestLap?: {
    rank: string
    lap: string
    Time: { time: string }
    AverageSpeed: { units: string; speed: string }
  }
}

// ── OpenSky types ─────────────────────────────────────────────────────────────

export interface Aircraft {
  icao24: string             // hex transponder address
  callsign: string           // flight number / tail number (padded with spaces)
  origin_country: string
  time_position: number | null
  last_contact: number
  longitude: number | null   // WGS-84 decimal
  latitude: number | null
  baro_altitude: number | null  // metres
  on_ground: boolean
  velocity: number | null    // m/s ground speed
  true_track: number | null  // degrees from north
  vertical_rate: number | null
  geo_altitude: number | null
  squawk: string
  spi: boolean
  position_source: 0 | 1 | 2 | 3  // 0=ADS-B, 1=ASTERIX, 2=MLAT, 3=FLARM
}

export interface AircraftResponse {
  time: number
  states: Aircraft[]
  count: number
}

export interface AirshowResponse extends AircraftResponse {
  venue: string
}

// ── OSM / Venue types ─────────────────────────────────────────────────────────

export interface OSMElement {
  type: 'node' | 'way' | 'relation'
  id: number
  lat?: number
  lon?: number
  tags?: Record<string, string>
  center?: { lat: number; lon: number }
}

export interface VenueResponse {
  elements: OSMElement[]
  count: number
}

export interface NearbyVenueResponse extends VenueResponse {
  lat: number
  lon: number
  radius_m: number
  venues: OSMElement[]
}

// ═══════════════════════════════════════════════════════════════════════════════
// F1 LIVE — OpenF1 (openf1.org)
// ═══════════════════════════════════════════════════════════════════════════════

/** Current or most recent F1 session metadata. */
export async function getF1Session(): Promise<F1Session> {
  return fetchAPI<F1Session>('/f1/session')
}

/** Live driver positions for the current session. */
export async function getF1Positions(): Promise<{
  session_key: number
  positions: F1Position[]
}> {
  return fetchAPI('/f1/positions')
}

/** All drivers in the current session with team colours and headshots. */
export async function getF1Drivers(): Promise<{
  session_key: number
  drivers: F1Driver[]
}> {
  return fetchAPI('/f1/drivers')
}

/** Current track weather. */
export async function getF1Weather(): Promise<{
  session_key: number
  weather: F1Weather[]
}> {
  return fetchAPI('/f1/weather')
}

/**
 * Lap times for the current session.
 * @param driverNumber — optional, e.g. 44 for Hamilton
 */
export async function getF1Laps(driverNumber?: number): Promise<{
  session_key: number
  laps: F1Lap[]
}> {
  const q = driverNumber ? `?driver_number=${driverNumber}` : ''
  return fetchAPI(`/f1/laps${q}`)
}

/** Race control messages — flags, safety car, DRS, incidents. */
export async function getF1RaceControl(): Promise<{
  session_key: number
  messages: F1RaceControlMessage[]
}> {
  return fetchAPI('/f1/race-control')
}

// ═══════════════════════════════════════════════════════════════════════════════
// F1 STANDINGS & HISTORY — Jolpica (jolpi.ca)
// ═══════════════════════════════════════════════════════════════════════════════

/**
 * Current driver championship standings.
 * Response is raw Jolpica MRData — access via data.StandingsTable.StandingsLists[0].DriverStandings
 */
export async function getF1DriverStandings(): Promise<unknown> {
  return fetchAPI('/f1/standings/drivers')
}

/** Current constructor championship standings. */
export async function getF1ConstructorStandings(): Promise<unknown> {
  return fetchAPI('/f1/standings/constructors')
}

/** Full race calendar for the current season. */
export async function getF1Schedule(): Promise<unknown> {
  return fetchAPI('/f1/schedule')
}

/** Results for the most recently completed race. */
export async function getF1LastResult(): Promise<unknown> {
  return fetchAPI('/f1/results/last')
}

/** Results for a specific round number. */
export async function getF1RoundResult(round: number): Promise<unknown> {
  return fetchAPI(`/f1/results/${round}`)
}

/** Qualifying results for a specific round. */
export async function getF1Qualifying(round: number): Promise<unknown> {
  return fetchAPI(`/f1/qualifying/${round}`)
}

/** All circuits on the current calendar with GPS coordinates. */
export async function getF1Circuits(): Promise<unknown> {
  return fetchAPI('/f1/circuits')
}

/** Pit stop data for a specific round. */
export async function getF1PitStops(round: number): Promise<unknown> {
  return fetchAPI(`/f1/pitstops/${round}`)
}

/** Lap-by-lap timing for a specific round. */
export async function getF1LapTimes(round: number): Promise<unknown> {
  return fetchAPI(`/f1/laptimes/${round}`)
}

/** Sprint race results for a specific round. */
export async function getF1Sprint(round: number): Promise<unknown> {
  return fetchAPI(`/f1/sprint/${round}`)
}

// ═══════════════════════════════════════════════════════════════════════════════
// AVIATION — OpenSky Network (opensky-network.org)
// ═══════════════════════════════════════════════════════════════════════════════

/**
 * All aircraft within a bounding box (WGS-84 decimal degrees).
 * Great for building a live map overlay during an airshow.
 *
 * @example
 * // EAA AirVenture Oshkosh airspace
 * getAircraftInBoundingBox(43.95, -88.65, 44.05, -88.50)
 */
export async function getAircraftInBoundingBox(
  lamin: number,
  lomin: number,
  lamax: number,
  lomax: number
): Promise<AircraftResponse> {
  return fetchAPI(`/aviation/aircraft?lamin=${lamin}&lomin=${lomin}&lamax=${lamax}&lomax=${lomax}`)
}

/**
 * Single aircraft by ICAO24 transponder address.
 * @param icao24 — hex string, e.g. "a0b1c2"
 */
export async function getAircraftByICAO(icao24: string): Promise<{
  time: number
  aircraft: Aircraft
}> {
  return fetchAPI(`/aviation/aircraft/${icao24}`)
}

/**
 * Live aircraft near a named airshow venue.
 * @param venue — "oshkosh" | "reno" | "dayton" | "miramar" | "farnborough" | "paris" | "dubai"
 */
export async function getAircraftAtAirshow(
  venue: 'oshkosh' | 'reno' | 'dayton' | 'miramar' | 'farnborough' | 'paris' | 'dubai'
): Promise<AirshowResponse> {
  return fetchAPI(`/aviation/airshow?venue=${venue}`)
}

/**
 * Flight history for an aircraft.
 * Anonymous OpenSky tier: max ~2 hours of history.
 * @param begin — Unix timestamp
 * @param end   — Unix timestamp
 */
export async function getFlightHistory(
  icao24: string,
  begin: number,
  end: number
): Promise<unknown> {
  return fetchAPI(`/aviation/history/${icao24}?begin=${begin}&end=${end}`)
}

// ═══════════════════════════════════════════════════════════════════════════════
// VENUES — OpenStreetMap Overpass + Nominatim
// ═══════════════════════════════════════════════════════════════════════════════

/** All motor racing circuits in OSM worldwide (600+ entries). */
export async function getRacingCircuits(): Promise<VenueResponse> {
  return fetchAPI('/venues/tracks')
}

/** Go-kart venues worldwide. */
export async function getKartTracks(): Promise<VenueResponse> {
  return fetchAPI('/venues/karts')
}

/** Drag strips worldwide. */
export async function getDragStrips(): Promise<VenueResponse> {
  return fetchAPI('/venues/drag')
}

/**
 * Airfields and aerodromes.
 * @param country — optional ISO 3166-1 alpha-2 code, e.g. "US"
 */
export async function getAirfields(country?: string): Promise<VenueResponse> {
  const q = country ? `?country=${country}` : ''
  return fetchAPI(`/venues/airfields${q}`)
}

/** Marinas and harbours — offshore race venues. */
export async function getMarinas(): Promise<VenueResponse> {
  return fetchAPI('/venues/marinas')
}

/**
 * All motorsport venues within a radius of a point.
 * @param lat    — WGS-84 latitude
 * @param lon    — WGS-84 longitude
 * @param radius — metres, max 200 000 (200 km), default 50 000
 */
export async function getNearbyVenues(
  lat: number,
  lon: number,
  radius = 50000
): Promise<NearbyVenueResponse> {
  return fetchAPI(`/venues/nearby?lat=${lat}&lon=${lon}&radius=${radius}`)
}

/**
 * Search venues by name using OSM Nominatim.
 * @example searchVenuesByName("Monza")
 * @example searchVenuesByName("Daytona")
 */
export async function searchVenuesByName(q: string): Promise<{
  query: string
  results: unknown[]
}> {
  return fetchAPI(`/venues/search?q=${encodeURIComponent(q)}`)
}

// ═══════════════════════════════════════════════════════════════════════════════
// WEBSOCKET HELPER
// ═══════════════════════════════════════════════════════════════════════════════

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

export type WSMessageHandler = (data: unknown) => void

/**
 * connectTelemetryStream opens a WebSocket to /ws/streams and calls
 * onMessage for every incoming frame. Returns a cleanup function.
 *
 * Usage in a React component:
 *
 *   useEffect(() => {
 *     return connectTelemetryStream((data) => {
 *       const msg = data as { type: string; data: F1LivePayload }
 *       if (msg.type === 'f1_live') setTelemetry(msg.data)
 *     })
 *   }, [])
 *
 * The WebSocket sends two types of frames:
 *   { type: 'f1_live', data: F1LivePayload }  — rich F1 telemetry (from PollOpenF1)
 *   StreamTelemetry                            — legacy simulated shape (from SimulateTelemetry)
 */
export function connectTelemetryStream(onMessage: WSMessageHandler): () => void {
  const ws = new WebSocket(`${WS_URL}/ws/streams`)

  ws.onopen = () => console.log('[Nitrous WS] connected')
  ws.onclose = () => console.log('[Nitrous WS] disconnected')
  ws.onerror = (e) => console.error('[Nitrous WS] error', e)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      onMessage(data)
    } catch {
      console.warn('[Nitrous WS] unparseable frame', event.data)
    }
  }

  // Return cleanup — call this in useEffect's return
  return () => ws.close()
}
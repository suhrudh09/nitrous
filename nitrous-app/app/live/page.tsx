'use client'
import Link from 'next/link'
import { useState, useEffect, useRef } from 'react'
import Nav from '@/components/Nav'
import { getEvents, getOpenF1RecentSessions, getOpenF1SessionTelemetry, getStreams } from '@/lib/api'
import type { Event, OpenF1RecentSession, OpenF1SessionTelemetry, Stream, StreamTelemetry } from '@/types'
import styles from './live.module.css'

export default function LivePage() {
  const [streams, setStreams] = useState<Stream[]>([])
  const [recentRaces, setRecentRaces] = useState<OpenF1RecentSession[]>([])
  const [selectedRaceKey, setSelectedRaceKey] = useState<number | null>(null)
  const [selectedRaceTelemetry, setSelectedRaceTelemetry] = useState<OpenF1SessionTelemetry | null>(null)
  const [selectedRaceTelemetryLoading, setSelectedRaceTelemetryLoading] = useState(false)
  const [selectedRaceTelemetryError, setSelectedRaceTelemetryError] = useState('')
  const [upcomingEvents, setUpcomingEvents] = useState<Event[]>([])
  const [active, setActive] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [playbackState, setPlaybackState] = useState<'loading' | 'playing' | 'paused' | 'error'>('loading')
  const [playbackMessage, setPlaybackMessage] = useState('')
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const pickDefaultStreamId = (data: Stream[]): string => {
    if (data.length === 0) return ''
    return data.find((s) => s.id === 'openf1-live')?.id ?? data[0].id
  }

  const mergeIncomingStreams = (current: Stream[], incoming: Stream[]): Stream[] => {
    const byId = new Map(current.map((s) => [s.id, s]))
    return incoming.map((s) => ({
      ...byId.get(s.id),
      ...s,
    }))
  }

  const formatOpenF1Date = (value?: string): string => {
    if (!value) return 'N/A'
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) return value

    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      timeZoneName: 'short',
    }).format(parsed)
  }

  const formatRecentRaceDate = (value: string): string => {
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) return value

    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(parsed)
  }

  const loadSessionTelemetry = async (sessionKey: number) => {
    setSelectedRaceKey(sessionKey)
    setSelectedRaceTelemetryLoading(true)
    setSelectedRaceTelemetryError('')
    setSelectedRaceTelemetry(null)

    try {
      const telemetry = await getOpenF1SessionTelemetry(sessionKey)
      setSelectedRaceTelemetry(telemetry)
    } catch {
      setSelectedRaceTelemetry(null)
      setSelectedRaceTelemetryError('Unable to load telemetry for this race')
    } finally {
      setSelectedRaceTelemetryLoading(false)
    }
  }

  const resetPastTelemetrySelection = () => {
    setSelectedRaceKey(null)
    setSelectedRaceTelemetry(null)
    setSelectedRaceTelemetryError('')
    setSelectedRaceTelemetryLoading(false)
  }

  const parseEventDateTime = (event: Event): Date | null => {
    if (!event.date) return null

    const dateOnly = event.date.includes('T') ? event.date.split('T')[0] : event.date
    if (!event.time) {
      const parsedDateOnly = new Date(`${dateOnly}T00:00:00Z`)
      return Number.isNaN(parsedDateOnly.getTime()) ? null : parsedDateOnly
    }

    const normalizedTime = event.time.trim()

    const utcClockMatch = normalizedTime.match(/^(\d{1,2}):(\d{2})\s*UTC$/i)
    if (utcClockMatch) {
      const hours = utcClockMatch[1].padStart(2, '0')
      const minutes = utcClockMatch[2]
      const parsedUTCClock = new Date(`${dateOnly}T${hours}:${minutes}:00Z`)
      return Number.isNaN(parsedUTCClock.getTime()) ? null : parsedUTCClock
    }

    // Treat only ISO-like datetime strings as full timestamps.
    if (/^\d{4}-\d{2}-\d{2}T/.test(normalizedTime)) {
      const parsedFullTime = new Date(normalizedTime)
      return Number.isNaN(parsedFullTime.getTime()) ? null : parsedFullTime
    }

    const hasTimezone = /Z$|[+-]\d{2}:\d{2}$/.test(normalizedTime)
    const isoCandidate = `${dateOnly}T${normalizedTime}${hasTimezone ? '' : 'Z'}`
    const parsed = new Date(isoCandidate)
    return Number.isNaN(parsed.getTime()) ? null : parsed
  }

  const formatUpcomingTime = (event: Event): string => {
    const eventDate = parseEventDateTime(event)
    if (!eventDate) return 'TBA'

    const diffMs = eventDate.getTime() - Date.now()
    if (diffMs > 0 && diffMs < 24 * 60 * 60 * 1000) {
      const totalMinutes = Math.max(1, Math.floor(diffMs / 60000))
      const hours = Math.floor(totalMinutes / 60)
      const minutes = totalMinutes % 60
      if (hours > 0 && minutes > 0) return `In ${hours}h ${minutes}m`
      if (hours > 0) return `In ${hours}h`
      return `In ${minutes}m`
    }

    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    }).format(eventDate)
  }

  // ── Fetch streams from API on mount ────────────────────────────────────────
  useEffect(() => {
    Promise.all([getStreams(), getOpenF1RecentSessions(8), getEvents()])
      .then(([streamData, recentData, eventData]) => {
        setStreams(streamData)
        setRecentRaces(recentData)
        const now = Date.now()
        const upcoming = eventData
          .filter((event) => {
            if (event.isLive) return false
            const parsed = parseEventDateTime(event)
            return parsed ? parsed.getTime() >= now : false
          })
          .sort((a, b) => {
            const aTime = parseEventDateTime(a)?.getTime() ?? Number.MAX_SAFE_INTEGER
            const bTime = parseEventDateTime(b)?.getTime() ?? Number.MAX_SAFE_INTEGER
            return aTime - bTime
          })
          .slice(0, 6)
        setUpcomingEvents(upcoming)
        setActive((prev) => prev || pickDefaultStreamId(streamData))
      })
      .catch(() => setError('Could not load streams'))
      .finally(() => setLoading(false))
  }, [])

  // Keep stream metadata fresh (leader/session/viewers) even when websocket messages are partial.
  useEffect(() => {
    const intervalId = window.setInterval(() => {
      getStreams()
        .then((data) => {
          setStreams((prev) => mergeIncomingStreams(prev, data))
          setActive((prev) => {
            if (!prev) return pickDefaultStreamId(data)
            const stillExists = data.some((s) => s.id === prev)
            return stillExists ? prev : pickDefaultStreamId(data)
          })
        })
        .catch(() => {
          // Keep existing UI state when background refresh fails.
        })
    }, 8000)

    return () => {
      window.clearInterval(intervalId)
    }
  }, [])

  useEffect(() => {
    if (!featured?.playbackUrl) {
      setPlaybackState('error')
      setPlaybackMessage('No playback source available')
      return
    }

    const video = videoRef.current
    if (!video) return

    setPlaybackState('paused')
    setPlaybackMessage('Selected feed. Press play to start playback.')
  }, [active, streams])

  // ── WebSocket — real-time telemetry updates ────────────────────────────────
  useEffect(() => {
    const wsUrl =
      process.env.NEXT_PUBLIC_WS_URL || '<iframe width="560" height="315" src="https://www.youtube.com/embed/YA1ruaeoZZY?si=8sEKuWUP94CoIG4h" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>ws://localhost:8080/api/streams/ws'
    const ws = new WebSocket(wsUrl)

    ws.onmessage = (e) => {
      try {
        const telemetry: StreamTelemetry = JSON.parse(e.data)
        setStreams((prev) =>
          prev.map((s) =>
            s.id === telemetry.streamId
              ? {
                  ...s,
                      currentSpeed:
                        typeof (telemetry as any).speedKph === 'number'
                          ? `${(telemetry as any).speedKph} km/h`
                          : typeof (telemetry as any).currentSpeed === 'string'
                          ? (telemetry as any).currentSpeed
                          : s.currentSpeed,
                  viewers:
                    typeof (telemetry as any).viewers === 'number'
                      ? (telemetry as any).viewers
                      : s.viewers,
                  currentLeader:
                    typeof (telemetry as any).currentLeader === 'string'
                      ? (telemetry as any).currentLeader
                      : s.currentLeader,
                  subtitle:
                    typeof (telemetry as any).subtitle === 'string'
                      ? (telemetry as any).subtitle
                      : s.subtitle,
                }
              : s
          )
        )
      } catch {
        // ignore malformed messages
      }
    }

    ws.onerror = () => {
      // WebSocket unavailable — REST data still shows, silently degrade
    }

    return () => {
      ws.close()
    }
  }, [])

  const featured = streams.find((s) => s.id === active) ?? streams[0]
  const isOpenF1 = featured?.id === 'openf1-live'
  const selectedRaceSession =
    selectedRaceKey != null
      ? recentRaces.find((race) => race.session_key === selectedRaceKey) ?? null
      : null
  const loadingPastTelemetry = isOpenF1 && selectedRaceTelemetryLoading && selectedRaceKey != null
  const showingPastTelemetry = isOpenF1 && selectedRaceTelemetry != null && selectedRaceSession != null
  const pastModeActive = showingPastTelemetry || loadingPastTelemetry
  const tone = featured?.color
    ? `${featured.color.charAt(0).toUpperCase()}${featured.color.slice(1)}`
    : 'Cyan'

  const handleTogglePlayback = async () => {
    const video = videoRef.current
    if (!video) return

    try {
      if (video.paused) {
        await video.play()
        setPlaybackState('playing')
        setPlaybackMessage('Live playback active')
      } else {
        video.pause()
        setPlaybackState('paused')
        setPlaybackMessage('Playback paused')
      }
    } catch {
      setPlaybackState('error')
      setPlaybackMessage('Playback blocked by the browser. Use the player controls to start the feed.')
    }
  }

  const totalViewers = streams.reduce((acc, s) => acc + s.viewers, 0)
  const viewerDisplay =
    totalViewers >= 1_000_000
      ? `${(totalViewers / 1_000_000).toFixed(1)}M`
      : totalViewers >= 1_000
      ? `${(totalViewers / 1_000).toFixed(0)}K`
      : String(totalViewers)

  if (loading) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--muted)', fontFamily: 'var(--font-mono)' }}>
            LOADING STREAMS...
          </div>
        </main>
      </>
    )
  }

  if (error || streams.length === 0) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--muted)', fontFamily: 'var(--font-mono)' }}>
            {error || 'NO LIVE STREAMS AT THIS TIME'}
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Page Header */}
        <div className={styles.pageHeader}>
          <div className={styles.headerLeft}>
            <div className={styles.liveChip}>
              <span className={styles.liveDot}></span>
              <span>LIVE NOW</span>
            </div>
            <h1 className={styles.pageTitle}>STREAMS</h1>
            <p className={styles.pageSubtitle}>
              {streams.length} channel{streams.length !== 1 ? 's' : ''} broadcasting · {viewerDisplay}+ watching
            </p>
          </div>
          <div className={styles.headerStats}>
            <div className={styles.stat}>
              <span className={styles.statVal}>{streams.length}</span>
              <span className={styles.statLabel}>LIVE</span>
            </div>
            <div className={styles.statDivider}></div>
            <div className={styles.stat}>
              <span className={styles.statVal}>{viewerDisplay}</span>
              <span className={styles.statLabel}>VIEWERS</span>
            </div>
            <div className={styles.statDivider}></div>
            <div className={styles.stat}>
              <span className={styles.statVal}>{upcomingEvents.length}</span>
              <span className={styles.statLabel}>UPCOMING</span>
            </div>
          </div>
        </div>

        <div className={styles.grid}>
          {/* Main Player */}
          <div className={styles.playerWrap}>
            <div
              className={`${styles.player} ${
                styles[
                  `player${tone}`
                ]
              }`}
            >
              {!isOpenF1 && featured.playbackUrl ? (
                <video
                  key={featured.id}
                  ref={videoRef}
                  className={styles.playerVideo}
                  src={featured.playbackUrl}
                  controls
                  autoPlay
                  muted
                  playsInline
                  preload="auto"
                  onLoadedData={() => {
                    setPlaybackState('paused')
                    setPlaybackMessage('Selected feed. Press play to start playback.')
                  }}
                  onWaiting={() => {
                    setPlaybackState('loading')
                    setPlaybackMessage('Buffering live feed...')
                  }}
                  onPause={() => {
                    setPlaybackState('paused')
                  }}
                  onPlay={() => {
                    setPlaybackState('playing')
                    setPlaybackMessage('Live playback active')
                  }}
                  onError={() => {
                    setPlaybackState('error')
                    setPlaybackMessage('Unable to load video source')
                  }}
                />
              ) : isOpenF1 ? (
                <div className={styles.telemetryPanel}>
                  <div className={styles.telemetryTitle}>
                    {pastModeActive ? 'OPENF1 TELEMETRY' : 'OPENF1 LIVE TELEMETRY'}
                  </div>
                  {loadingPastTelemetry ? (
                    <div className={styles.telemetryCard}>
                      <div className={styles.telemetryLabel}>STATUS</div>
                      <div className={styles.telemetryValue}>FETCHING SELECTED RACE TELEMETRY...</div>
                    </div>
                  ) : showingPastTelemetry ? (
                    <div className={styles.telemetryGrid}>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>SESSION</div>
                        <div className={styles.telemetryValue}>
                          {`${selectedRaceSession.session_name} - ${selectedRaceSession.circuit_short_name}`}
                        </div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>START TIME</div>
                        <div className={styles.telemetryValue}>{formatOpenF1Date(selectedRaceSession.date_start)}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>END TIME</div>
                        <div className={styles.telemetryValue}>{formatOpenF1Date(selectedRaceSession.date_end)}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>country_name</div>
                        <div className={styles.telemetryValue}>{selectedRaceSession.country_name}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>LEADER</div>
                        <div className={styles.telemetryValue}>{selectedRaceTelemetry.current_leader || 'Standby'}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>CURRENT SPEED</div>
                        <div className={styles.telemetryValue}>{`${selectedRaceTelemetry.speed_kph} km/h`}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>SNAPSHOT</div>
                        <div className={styles.telemetryValue}>{formatOpenF1Date(selectedRaceTelemetry.captured_at)}</div>
                      </div>
                    </div>
                  ) : (
                    <div className={styles.telemetryGrid}>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>SESSION</div>
                        <div className={styles.telemetryValue}>{featured.subtitle || 'Standby'}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>START TIME</div>
                        <div className={styles.telemetryValue}>{formatOpenF1Date(featured.date_start)}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>END TIME</div>
                        <div className={styles.telemetryValue}>{formatOpenF1Date(featured.date_end)}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>country_name</div>
                        <div className={styles.telemetryValue}>{featured.country_name || 'N/A'}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>LEADER</div>
                        <div className={styles.telemetryValue}>{featured.currentLeader || 'Standby'}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>CURRENT SPEED</div>
                        <div className={styles.telemetryValue}>{featured.currentSpeed || '0 km/h'}</div>
                      </div>
                      <div className={styles.telemetryCard}>
                        <div className={styles.telemetryLabel}>VIEWERS</div>
                        <div className={styles.telemetryValue}>{featured.viewers.toLocaleString()}</div>
                      </div>
                    </div>
                  )}
                  {!pastModeActive && featured.externalWatch && featured.externalWatch.length > 0 ? (
                    <div className={styles.watchOptions}>
                      <div className={styles.watchOptionsLabel}>WATCH LIVE RACES</div>
                      <div className={styles.watchButtons}>
                        {featured.externalWatch.map((w) => (
                          <a
                            key={`${w.platform}-${w.url}`}
                            href={w.url}
                            target="_blank"
                            rel="noreferrer"
                            className={styles.watchButton}
                          >
                            {w.label}
                          </a>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </div>
              ) : (
                <div className={styles.noPlayback}>No video source available</div>
              )}

              {/* Fake video overlay */}
              <div className={styles.playerNoise}></div>
              <div className={styles.playerScanlines}></div>

              {/* Top HUD */}
              <div className={styles.playerHudTop}>
                {!pastModeActive ? (
                  <div className={styles.playerLiveBadge}>
                    <span className={styles.liveDot}></span>LIVE
                  </div>
                ) : null}
                <span className={styles.playerCat}>{featured.category}</span>
                <div className={styles.playerQuality}>{featured.quality}</div>
                <div className={styles.playerStatus}>
                  {playbackState === 'playing'
                    ? 'PLAYING'
                    : playbackState === 'loading'
                    ? 'BUFFERING'
                    : playbackState === 'paused'
                    ? 'PAUSED'
                    : 'ERROR'}
                </div>
              </div>

              {playbackMessage ? (
                <div className={styles.playerMessage}>{playbackMessage}</div>
              ) : null}

              {/* Center play */}
              {!isOpenF1 ? (
                <div className={styles.playerCenter}>
                  <button className={styles.playBtn} type="button" onClick={handleTogglePlayback}>
                    <span>▶</span>
                  </button>
                </div>
              ) : null}

              {/* Bottom HUD */}
              <div className={styles.playerHudBot}>
                <div className={styles.playerInfo}>
                  <div className={styles.playerTitle}>{featured.title}</div>
                  <div className={styles.playerSub}>{featured.subtitle}</div>
                </div>
                <div className={styles.playerMeta}>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>LEADER</span>
                    <span className={styles.metaVal}>{featured.currentLeader}</span>
                  </div>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>SPEED</span>
                    <span className={styles.metaVal}>{featured.currentSpeed}</span>
                  </div>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>VIEWERS</span>
                    <span className={styles.metaVal}>
                      {featured.viewers >= 1_000_000
                        ? `${(featured.viewers / 1_000_000).toFixed(1)}M`
                        : featured.viewers >= 1_000
                        ? `${(featured.viewers / 1_000).toFixed(0)}K`
                        : featured.viewers}
                    </span>
                  </div>
                </div>
              </div>

              {/* Progress bar */}
              <div className={styles.playerProgress}>
                <div className={styles.playerProgressFill}></div>
              </div>
            </div>

            {/* Stream controls */}
            <div className={styles.controls}>
              <button className={styles.controlBtn}>⏮</button>
              <button className={styles.controlBtn}>⏸</button>
              <button className={styles.controlBtn}>⏭</button>
              <div className={styles.controlSep}></div>
              <button className={styles.controlBtn}>🔇</button>
              <div className={styles.volumeBar}>
                <div className={styles.volumeFill}></div>
              </div>
              <div className={styles.controlSep}></div>
              <button className={`${styles.controlBtn} ${styles.controlBtnRight}`}>⛶</button>
            </div>
          </div>

          {/* Sidebar */}
          <div className={styles.sidebar}>
            <div className={styles.sideSection}>
              <div className={styles.sideLabel}>LIVE CHANNELS</div>
              <div className={styles.channelList}>
                {streams.map((stream) => (
                  <div
                    key={stream.id}
                    className={`${styles.channelCard} ${
                      active === stream.id ? styles.channelCardActive : ''
                    } ${
                      styles[
                        `channelCard${
                          stream.color.charAt(0).toUpperCase() + stream.color.slice(1)
                        }`
                      ]
                    }`}
                    onClick={() => {
                      setActive(stream.id)
                      resetPastTelemetrySelection()
                    }}
                  >
                    <div className={styles.channelTop}>
                      <div className={styles.channelLive}>
                        <span className={styles.liveDotSm}></span>
                        {stream.category}
                      </div>
                      <div className={styles.channelViewers}>
                        👁{' '}
                        {stream.viewers >= 1_000_000
                          ? `${(stream.viewers / 1_000_000).toFixed(1)}M`
                          : stream.viewers >= 1_000
                          ? `${(stream.viewers / 1_000).toFixed(0)}K`
                          : stream.viewers}
                      </div>
                    </div>
                    <div className={styles.channelTitle}>{stream.title}</div>
                    <div className={styles.channelSub}>
                      {stream.subtitle} · {stream.location}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className={styles.sideSection}>
              <div className={styles.sideLabel}>UP NEXT</div>
              <div className={styles.upcomingList}>
                {upcomingEvents.length === 0 ? (
                  <div className={styles.upcomingItem}>
                    <div className={styles.upcomingTitle}>No upcoming events scheduled</div>
                  </div>
                ) : (
                  upcomingEvents.map((event) => (
                    <div key={event.id} className={styles.upcomingItem}>
                      <div className={styles.upcomingTime}>{formatUpcomingTime(event)}</div>
                      <div className={styles.upcomingTitle}>{event.title}</div>
                      <div className={styles.upcomingCat}>{event.category.toUpperCase()}</div>
                    </div>
                  ))
                )}
              </div>
            </div>

            <div className={styles.sideSection}>
              <div className={styles.sideLabel}>RECENT F1 RACES</div>
              <div className={styles.upcomingList}>
                {recentRaces.length === 0 ? (
                  <div className={styles.upcomingItem}>
                    <div className={styles.upcomingTitle}>No recent races available</div>
                  </div>
                ) : (
                  recentRaces.map((race) => (
                    <button
                      key={race.session_key}
                      type="button"
                      className={`${styles.recentRaceButton} ${selectedRaceKey === race.session_key ? styles.recentRaceButtonActive : ''}`}
                      onClick={() => {
                        setActive('openf1-live')
                        void loadSessionTelemetry(race.session_key)
                      }}
                    >
                      <div className={styles.upcomingTime}>{formatRecentRaceDate(race.date_start)}</div>
                      <div className={styles.upcomingTitle}>{race.session_name} - {race.circuit_short_name}</div>
                      <div className={styles.upcomingCat}>{race.country_name}</div>
                    </button>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}
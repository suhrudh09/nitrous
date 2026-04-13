'use client'
import Link from 'next/link'
import { useState, useEffect, useRef } from 'react'
import Nav from '@/components/Nav'
import { getStreams } from '@/lib/api'
import type { Stream, StreamTelemetry } from '@/types'
import styles from './live.module.css'

const upcomingStreams = [
  { title: 'Baja 1000 — Stage 4', time: 'In 2h 14m', category: 'OFF-ROAD' },
  { title: 'Crop Duster Air Racing', time: 'In 5h 30m', category: 'AIR' },
  { title: 'Jet Ski Championship — Finals', time: 'Tomorrow 14:00', category: 'WATER' },
  { title: 'Formula E Round 6 — Monaco', time: 'Tomorrow 18:00', category: 'MOTORSPORT' },
]

export default function LivePage() {
  const [streams, setStreams] = useState<Stream[]>([])
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

  // ── Fetch streams from API on mount ────────────────────────────────────────
  useEffect(() => {
    getStreams()
      .then((data) => {
        setStreams(data)
        setActive((prev) => prev || pickDefaultStreamId(data))
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
      process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/streams/ws'
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
              <span className={styles.statVal}>{upcomingStreams.length}</span>
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
                  <div className={styles.telemetryTitle}>OPENF1 LIVE TELEMETRY</div>
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
                  {featured.externalWatch && featured.externalWatch.length > 0 ? (
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
                <div className={styles.playerLiveBadge}>
                  <span className={styles.liveDot}></span>LIVE
                </div>
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
                    onClick={() => setActive(stream.id)}
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
                {upcomingStreams.map((s, i) => (
                  <div key={i} className={styles.upcomingItem}>
                    <div className={styles.upcomingTime}>{s.time}</div>
                    <div className={styles.upcomingTitle}>{s.title}</div>
                    <div className={styles.upcomingCat}>{s.category}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}
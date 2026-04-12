'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getJourneys, bookJourney } from '@/lib/api'
import type { Journey } from '@/types'
import styles from './journeys.module.css'

const badgeColors: Record<string, string> = {
  EXCLUSIVE: 'var(--cyan)',
  'MEMBERS ONLY': '#facc15',
  LIMITED: 'var(--red)',
}

const colorMap: Record<string, string> = {
  red: 'var(--red)',
  cyan: 'var(--cyan)',
  orange: '#fb923c',
  blue: '#60a5fa',
  purple: '#a78bfa',
  gold: '#facc15',
}

export default function JourneysPage() {
  const [journeys, setJourneys] = useState<Journey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [booking, setBooking] = useState<string | null>(null)
  const [booked, setBooked] = useState<string | null>(null)

  useEffect(() => {
    getJourneys()
      .then(setJourneys)
      .catch(() => setError('Could not load journeys'))
      .finally(() => setLoading(false))
  }, [])

  async function handleBook(journeyId: string) {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }
    setBooking(journeyId)
    try {
      const result = await bookJourney(journeyId, token)
      // Update slots locally so UI responds immediately
      setJourneys((prev) =>
        prev.map((j) => (j.id === result.journey.id ? result.journey : j))
      )
      setBooked(journeyId)
      setTimeout(() => setBooked(null), 2000)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Booking failed'
      alert(message)
    } finally {
      setBooking(null)
    }
  }

  if (loading) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--muted)', fontFamily: 'var(--font-mono)' }}>
            LOADING JOURNEYS...
          </div>
        </main>
      </>
    )
  }

  if (error) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--red)', fontFamily: 'var(--font-mono)' }}>
            {error}
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Hero Header */}
        <div className={styles.pageHero}>
          <div className={styles.heroLines}>
            {[...Array(5)].map((_, i) => (
              <div key={i} className={styles.heroLine}></div>
            ))}
          </div>
          <div className={styles.heroContent}>
            <div className={styles.headerTag}>/ JOURNEYS</div>
            <h1 className={styles.pageTitle}>
              LIVE THE
              <br />
              <span className={styles.titleAccent}>EXPERIENCE</span>
            </h1>
            <p className={styles.pageSubtitle}>
              Exclusive access journeys for those who want more than just a stream. Be there.
              Feel it. Live it.
            </p>
          </div>
          <div className={styles.heroStats}>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>{journeys.length}</div>
              <div className={styles.hStatL}>JOURNEYS</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>18</div>
              <div className={styles.hStatL}>COUNTRIES</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>{journeys.filter((j) => j.slotsLeft > 0).length}</div>
              <div className={styles.hStatL}>AVAILABLE</div>
            </div>
          </div>
        </div>

        {/* Journeys List */}
        <div className={styles.journeysList}>
          {journeys.map((journey, i) => {
            // Derive a color key from category keywords for consistent accent coloring
            const colorKey = journey.category.toLowerCase().includes('air')
              ? 'cyan'
              : journey.category.toLowerCase().includes('rally') ||
                journey.category.toLowerCase().includes('offroad') ||
                journey.category.toLowerCase().includes('desert')
              ? 'orange'
              : journey.category.toLowerCase().includes('water')
              ? 'blue'
              : 'red'
            const accent = colorMap[colorKey] ?? 'var(--cyan)'
            const isUrgent = journey.slotsLeft <= 4

            return (
              <div key={journey.id} className={styles.journeyCard}>
                {/* Left: number */}
                <div className={styles.journeyNum} style={{ color: accent }}>
                  {String(i + 1).padStart(2, '0')}
                </div>

                {/* Main content */}
                <div className={styles.journeyMain}>
                  <div className={styles.journeyHead}>
                    <div>
                      <div className={styles.journeyCat}>{journey.category}</div>
                      <div className={styles.journeyTitle}>{journey.title}</div>
                    </div>
                    <div
                      className={styles.journeyBadge}
                      style={{
                        color: badgeColors[journey.badge],
                        borderColor: badgeColors[journey.badge],
                      }}
                    >
                      {journey.badge}
                    </div>
                  </div>

                  <p className={styles.journeyDesc}>{journey.description}</p>

                  <div className={styles.journeyMeta}>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>📅</span>
                      <span>{new Date(journey.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
                    </div>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>💰</span>
                      <span>${journey.price.toLocaleString()} per person</span>
                    </div>
                  </div>
                </div>

                {/* Right: price + slots + CTA */}
                <div className={styles.journeyAside}>
                  <div className={styles.journeyPrice}>
                    <span className={styles.priceFrom}>FROM</span>
                    <span className={styles.priceVal} style={{ color: accent }}>
                      ${journey.price.toLocaleString()}
                    </span>
                    <span className={styles.priceLabel}>per person</span>
                  </div>

                  <div className={styles.slotsWrap}>
                    <div className={styles.slotsTop}>
                      <span className={`${styles.slotsLabel} ${isUrgent ? styles.slotsUrgent : ''}`}>
                        {isUrgent ? '🔥 ' : ''}
                        {journey.slotsLeft} SLOTS LEFT
                      </span>
                    </div>
                    <div className={styles.slotsBar}>
                      <div
                        className={styles.slotsFill}
                        style={{
                          width: `${Math.max((journey.slotsLeft / 20) * 100, 5)}%`,
                          background: isUrgent ? 'var(--red)' : accent,
                        }}
                      ></div>
                    </div>
                  </div>

                  <button
                    className={styles.bookBtn}
                    style={{
                      background: `linear-gradient(135deg, ${accent}22, ${accent}11)`,
                      borderColor: `${accent}66`,
                      color: accent,
                      opacity: journey.slotsLeft <= 0 ? 0.4 : 1,
                      cursor: journey.slotsLeft <= 0 ? 'not-allowed' : 'pointer',
                    }}
                    onClick={() => handleBook(journey.id)}
                    disabled={booking === journey.id || journey.slotsLeft <= 0}
                  >
                    {booked === journey.id
                      ? '✓ BOOKED'
                      : booking === journey.id
                      ? 'BOOKING...'
                      : journey.slotsLeft <= 0
                      ? 'SOLD OUT'
                      : 'BOOK JOURNEY →'}
                  </button>
                </div>

                {/* Color accent left border */}
                <div
                  className={styles.journeyBorder}
                  style={{ background: accent, boxShadow: `0 0 12px ${accent}` }}
                ></div>
              </div>
            )
          })}
        </div>
      </main>
    </>
  )
}
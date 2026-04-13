'use client'
import Link from 'next/link'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getEvents, setReminder } from '@/lib/api'
import styles from './events.module.css'

const cats = ['all', 'motorsport', 'offroad', 'water', 'air']

const catColors: Record<string, string> = {
  motorsport: 'cyan',
  offroad: 'orange',
  water: 'blue',
  air: 'purple',
}

const catIcons: Record<string, string> = {
  motorsport: '🏎️',
  offroad: '🏔️',
  water: '🌊',
  air: '🪂',
}

export default function EventsPage() {
  const [allEvents, setAllEvents] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('all')
  const [liveOnly, setLiveOnly] = useState(false)
  const [remindingId, setRemindingId] = useState<string | null>(null)

  const formatEventDate = (value: string): string => {
    const dateOnly = value.includes('T') ? value.split('T')[0] : value
    const parsed = new Date(`${dateOnly}T00:00:00Z`)
    if (Number.isNaN(parsed.getTime())) return value

    return new Intl.DateTimeFormat('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(parsed)
  }

  const formatEventTime = (value: string): string => {
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) return value

    const offsetMatch = value.match(/([+-])(\d{2}):(\d{2})$/)
    const offsetMinutes = offsetMatch
      ? (offsetMatch[1] === '-' ? -1 : 1) * (Number(offsetMatch[2]) * 60 + Number(offsetMatch[3]))
      : 0

    const adjusted = new Date(parsed.getTime() + offsetMinutes * 60_000)
    const timePart = new Intl.DateTimeFormat('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      timeZone: 'UTC',
    }).format(adjusted)

    if (!offsetMatch) return timePart

    const sign = offsetMatch[1]
    const hours = offsetMatch[2]
    const minutes = offsetMatch[3]

    return `${timePart} (UTC${sign}${hours}${minutes === '00' ? '' : `:${minutes}`})`
  }

  useEffect(() => {
    async function fetchData() {
      try {
        const events = await getEvents()
        setAllEvents(events)
      } catch (error) {
        console.error('Failed to fetch events:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  const handleSetReminder = async (eventId: string) => {
    const token = localStorage.getItem('authToken')
    if (!token) {
      window.location.href = '/login'
      return
    }

    setRemindingId(eventId)
    try {
      await setReminder(eventId, token)
      console.log('Reminder set for event:', eventId)
      // Optionally show a success message
    } catch (error) {
      console.error('Failed to set reminder:', error)
    } finally {
      setRemindingId(null)
    }
  }

  const filtered = allEvents.filter(e => {
    if (liveOnly && !e.isLive) return false
    if (filter === 'all') return true
    return e.category === filter
  })

  return (
    <>
      <Nav />
      <main className={styles.page}>
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ EVENTS</div>
            <h1 className={styles.pageTitle}>ALL EVENTS</h1>
            <p className={styles.pageSubtitle}>Upcoming races, qualifiers, and championships worldwide</p>
          </div>
        </div>

        {/* Filters */}
        <div className={styles.filterBar}>
          <div className={styles.catFilters}>
            {cats.map(cat => (
              <button
                key={cat}
                className={`${styles.catBtn} ${filter === cat ? styles.catBtnActive : ''}`}
                onClick={() => setFilter(cat)}
              >
                {cat === 'all' ? 'ALL' : `${catIcons[cat]} ${cat.toUpperCase()}`}
              </button>
            ))}
          </div>
          <button
            className={`${styles.liveToggle} ${liveOnly ? styles.liveToggleActive : ''}`}
            onClick={() => setLiveOnly(!liveOnly)}
          >
            <span className={styles.liveDot}></span>
            LIVE ONLY
          </button>
        </div>

        {/* Count */}
        <div className={styles.countBar}>
          <span className={styles.countTxt}>
            {filtered.length} event{filtered.length !== 1 ? 's' : ''}
          </span>
        </div>

        {/* Events Grid */}
        <div className={styles.eventsGrid}>
          {filtered.map((event) => (
            <div key={event.id} className={`${styles.eventCard} ${event.isLive ? styles.eventCardLive : ''}`}>
              {/* Color accent top */}
              <div className={`${styles.cardAccent} ${styles[`cardAccent${(catColors[event.category] || 'cyan').charAt(0).toUpperCase() + (catColors[event.category] || 'cyan').slice(1)}`]}`}></div>

              <div className={styles.cardTop}>
                <div className={styles.cardCat}>
                  <span>{catIcons[event.category]}</span>
                  <span>{event.category.toUpperCase()}</span>
                </div>
                {event.isLive ? (
                  <div className={styles.liveBadge}>
                    <span className={styles.liveDot}></span>LIVE
                  </div>
                ) : (
                  <div className={styles.priceBadge}>{event.price}</div>
                )}
              </div>

              <div className={styles.cardTitle}>{event.title}</div>
              <div className={styles.cardLocation}>📍 {event.location}</div>

              <div className={styles.cardBottom}>
                <div className={styles.cardDate}>
                  <span className={styles.cardDateLabel}>DATE</span>
                  <span className={styles.cardDateVal}>{formatEventDate(event.date)}</span>
                </div>
                <div className={styles.cardDate}>
                  <span className={styles.cardDateLabel}>TIME</span>
                  <span className={`${styles.cardDateVal} ${event.isLive ? styles.cardTimeLive : ''}`}>
                    {event.time ? formatEventTime(event.time) : 'N/A'}
                  </span>
                </div>
                {event.viewers && (
                  <div className={styles.cardDate}>
                    <span className={styles.cardDateLabel}>WATCHING</span>
                    <span className={styles.cardDateVal}>{event.viewers}</span>
                  </div>
                )}
              </div>

              <div className={styles.cardActions}>
                {event.isLive ? (
                  <Link href="/live" className={styles.btnWatch}>▶ Watch Live</Link>
                ) : (
                  <button 
                    className={styles.btnRemind}
                    onClick={() => handleSetReminder(event.id)}
                    disabled={remindingId === event.id}
                  >
                    {remindingId === event.id ? '⏳ Setting...' : 'Set Reminder'}
                  </button>
                )}
                <button className={styles.btnMore}>···</button>
              </div>
            </div>
          ))}
        </div>
      </main>
    </>
  )
}

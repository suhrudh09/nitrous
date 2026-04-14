'use client'

import Link from 'next/link'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getEvents, setReminder } from '@/lib/api'
import styles from './events.module.css'

// Configuration
const cats = ['all', 'motorsport', 'offroad', 'water', 'air'] as const
type Category = (typeof cats)[number]

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

// Interfaces
interface Event {
  id: string
  title: string
  category: string
  location: string
  date: string
  time?: string
  isLive: boolean
  price?: string
  viewers?: string
}

const getAccentColor = (category: string): string => {
  const color = catColors[category.toLowerCase()] || 'cyan'
  return color.charAt(0).toUpperCase() + color.slice(1)
}

export default function EventsPage() {
  // State
  const [allEvents, setAllEvents] = useState<Event[]>([])
  const [filter, setFilter] = useState<Category>('all')
  const [liveOnly, setLiveOnly] = useState(false)
  const [remindingId, setRemindingId] = useState<string | null>(null)

  // Fetch Data
  useEffect(() => {
    const fetchData = async () => {
      try {
        const events = await getEvents()
        setAllEvents(events)
      } catch (error) {
        console.error('Failed to fetch events:', error)
      }
    }
    fetchData()
  }, [])

  // Helpers
  const formatEventDate = (value: string): string => {
    const dateOnly = value.includes('T') ? value.split('T')[0] : value
    const parsed = new Date(`${dateOnly}T00:00:00Z`)
    if (isNaN(parsed.getTime())) return value

    return new Intl.DateTimeFormat('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(parsed)
  }

  const formatEventTime = (value: string): string => {
    const parsed = new Date(value)
    if (isNaN(parsed.getTime())) return value

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

  const handleSetReminder = async (eventId: string) => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }

    setRemindingId(eventId)
    try {
      await setReminder(eventId, token)
    } catch (error) {
      console.error('Failed to set reminder:', error)
    } finally {
      setRemindingId(null)
    }
  }

  // Filtered Results
  const filtered = allEvents.filter((e) => {
    const matchesLive = liveOnly ? e.isLive : true
    const matchesCat = filter === 'all' ? true : e.category.toLowerCase() === filter.toLowerCase()
    return matchesLive && matchesCat
  })

  return (
    <>
      <Nav />
      <main className={styles.page}>
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ EVENTS</div>
            <h1 className={styles.pageTitle}>ALL EVENTS</h1>
            <p className={styles.pageSubtitle}>Upcoming races and championships worldwide</p>
          </div>
        </div>

        <div className={styles.filterBar}>
          <div className={styles.catFilters}>
            {cats.map((cat) => (
              <button
                key={cat}
                className={`${styles.catBtn} ${filter === cat ? styles.catBtnActive : ''}`}
                onClick={() => setFilter(cat)}
              >
                {cat === 'all' ? 'ALL' : `${catIcons[cat] || ''} ${cat.toUpperCase()}`}
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

        <div className={styles.countBar}>
          <span className={styles.countTxt}>
            {filtered.length} event{filtered.length === 1 ? '' : 's'}
          </span>
        </div>

        <div className={styles.eventsGrid}>
          {filtered.map((event) => (
            <div key={event.id} className={`${styles.eventCard} ${event.isLive ? styles.eventCardLive : ''}`}>
              <div className={`${styles.cardAccent} ${styles[`cardAccent${getAccentColor(event.category)}`]}`}></div>

              <div className={styles.cardTop}>
                <div className={styles.cardCat}>
                  <span>{catIcons[event.category.toLowerCase()]}</span>
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
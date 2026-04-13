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

const getAccentColor = (category: string): string => {
  const color = catColors[category] || 'cyan'
  return color.charAt(0).toUpperCase() + color.slice(1)
}

export default function EventsPage() {
  const [allEvents, setAllEvents] = useState<any[]>([])
  const [filter, setFilter] = useState('all')
  const [liveOnly, setLiveOnly] = useState(false)
  const [remindingId, setRemindingId] = useState<string | null>(null)

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

  const handleSetReminder = async (eventId: string) => {
    const token = localStorage.getItem('authToken')
    if (!token) {
      globalThis.location.href = '/login'
      return
    }

    setRemindingId(eventId)
    try {
      await setReminder(eventId, token)
      console.log('Reminder set for event:', eventId)
    } catch (error) {
      console.error('Failed to set reminder:', error)
    } finally {
      setRemindingId(null)
    }
  }

  const filtered = allEvents.filter((e) => {
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
            {cats.map((cat) => (
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
            {filtered.length} event{filtered.length === 1 ? '' : 's'}
          </span>
        </div>

        {/* Events Grid */}
        <div className={styles.eventsGrid}>
          {filtered.map((event) => (
            <div key={event.id} className={`${styles.eventCard} ${event.isLive ? styles.eventCardLive : ''}`}>
              {/* Color accent top */}
              <div className={`${styles.cardAccent} ${styles[`cardAccent${getAccentColor(event.category)}`]}`}></div>

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
                  <span className={styles.cardDateVal}>{event.date}</span>
                </div>
                <div className={styles.cardDate}>
                  <span className={styles.cardDateLabel}>TIME</span>
                  <span className={`${styles.cardDateVal} ${event.isLive ? styles.cardTimeLive : ''}`}>{event.time}</span>
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

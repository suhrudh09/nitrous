'use client'
import Link from 'next/link'
import { useState } from 'react'
import Nav from '@/components/Nav'
import styles from './events.module.css'

const allEvents = [
  { id: '1', title: 'NASCAR Daytona 500', location: 'Daytona International Speedway · FL', date: 'Feb 16, 2026', time: 'LIVE NOW', isLive: true, category: 'motorsport', viewers: '1.2M', price: null },
  { id: '2', title: 'World Dirt Track Championship', location: 'Knob Noster · Missouri, USA', date: 'Feb 21, 2026', time: '20:00 UTC', isLive: true, category: 'motorsport', viewers: '340K', price: null },
  { id: '3', title: 'Dakar Rally — Stage 9', location: 'Al Ula → Ha\'il · Saudi Arabia', date: 'Feb 10, 2026', time: '09:00 UTC', isLive: false, category: 'offroad', viewers: null, price: 'Free' },
  { id: '4', title: 'Speed Boat Cup — Finals', location: 'Lake Como · Italy', date: 'Mar 2, 2026', time: '14:00 UTC', isLive: false, category: 'water', viewers: null, price: '$12' },
  { id: '5', title: 'Red Bull Skydive Series — Rd. 3', location: 'Interlaken Drop Zone · Switzerland', date: 'Mar 8, 2026', time: '11:30 UTC', isLive: false, category: 'air', viewers: null, price: 'Free' },
  { id: '6', title: 'Crop Duster Air Racing', location: 'Bakersfield Airfield · California', date: 'Mar 14, 2026', time: '16:00 UTC', isLive: false, category: 'air', viewers: null, price: '$8' },
  { id: '7', title: 'Baja 1000 Extreme', location: 'Ensenada · Baja California, Mexico', date: 'Mar 20, 2026', time: '08:00 UTC', isLive: false, category: 'offroad', viewers: null, price: 'Free' },
  { id: '8', title: 'Formula E — Monaco Round', location: 'Circuit de Monaco · Monte Carlo', date: 'Apr 5, 2026', time: '15:00 UTC', isLive: false, category: 'motorsport', viewers: null, price: '$15' },
  { id: '9', title: 'Jet Ski World Championship', location: 'Miami Beach · Florida', date: 'Apr 12, 2026', time: '12:00 UTC', isLive: false, category: 'water', viewers: null, price: '$10' },
]

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
  const [filter, setFilter] = useState('all')
  const [liveOnly, setLiveOnly] = useState(false)

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
                  <button className={styles.btnRemind}>Set Reminder</button>
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

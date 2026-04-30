'use client'

import Link from 'next/link'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getEvents, setReminder, getMyReminders, deleteReminder } from '@/lib/api'
import type { Reminder } from '@/types'
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
  const [remindersByEvent, setRemindersByEvent] = useState<Record<string, Reminder>>({})
  const [modalEvent, setModalEvent] = useState<Event | null>(null)
  const [modalMessage, setModalMessage] = useState('')
  const [modalRemindDate, setModalRemindDate] = useState('')
  const [modalRemindTime, setModalRemindTime] = useState('')
  const [modalSaving, setModalSaving] = useState(false)

  // Fetch Data
  useEffect(() => {
    let timer: ReturnType<typeof setInterval> | undefined
    let cancelled = false

    const syncReminders = async () => {
      const token = localStorage.getItem('nitrous_token')
      if (!token) return

      try {
        const reminders = await getMyReminders(token)
        if (cancelled) return

        const map: Record<string, Reminder> = {}
        reminders.forEach((reminder) => {
          map[reminder.eventId] = reminder
        })
        setRemindersByEvent(map)
      } catch {
        // ignore periodic reminder sync failures
      }
    }

    const fetchData = async () => {
      try {
        const events = await getEvents()
        setAllEvents(events)

        await syncReminders()

        timer = setInterval(syncReminders, 15000)
      } catch (error) {
        console.error('Failed to fetch events:', error)
      }
    }
    fetchData()

    return () => {
      cancelled = true
      if (timer) clearInterval(timer)
    }
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
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'UTC',
    }).format(adjusted)

    if (!offsetMatch) return `${timePart} UTC`

    const sign = offsetMatch[1]
    const hours = offsetMatch[2]
    const minutes = offsetMatch[3]

    return `${timePart} (UTC${sign}${hours}${minutes === '00' ? '' : `:${minutes}`})`
  }

  const toDateTimeUtcInput = (value: string): string => {
    const d = new Date(value)
    if (isNaN(d.getTime())) return ''
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())}T${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}`
  }

  const utcInputToIso = (value: string): string => {
    const match = value.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/)
    if (!match) return new Date(value).toISOString()

    const [, y, m, d, hh, mm] = match
    return new Date(Date.UTC(Number(y), Number(m) - 1, Number(d), Number(hh), Number(mm))).toISOString()
  }

  const splitUtcInput = (value: string): { date: string; time: string } => {
    const match = value.match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2})$/)
    if (!match) {
      return { date: '', time: '' }
    }

    return { date: match[1], time: match[2] }
  }

  const getTodayUtcDate = (): string => {
    const now = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${now.getUTCFullYear()}-${pad(now.getUTCMonth() + 1)}-${pad(now.getUTCDate())}`
  }

  const getTodayLocalDate = (): string => {
    const now = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
  }

  const toUtcInputFromParts = (date: string, time: string): string => {
    return `${date}T${time}`
  }

  const getEventStart = (event: Event): Date | null => {
    if (event.time) {
      const fromTime = new Date(event.time)
      if (!isNaN(fromTime.getTime())) return fromTime
    }

    if (event.date.includes('T')) {
      const fromDateTime = new Date(event.date)
      if (!isNaN(fromDateTime.getTime())) return fromDateTime
    }

    return null
  }

  const getDefaultReminderDateTimeUtcInput = (event: Event): string => {
    const now = Date.now()
    const minAllowed = new Date(now + 5 * 60 * 1000)
    const eventStart = getEventStart(event)

    if (!eventStart) {
      return toDateTimeUtcInput(new Date(now + 60 * 60 * 1000).toISOString())
    }

    const oneHourBefore = new Date(eventStart.getTime() - 60 * 60 * 1000)
    return toDateTimeUtcInput((oneHourBefore.getTime() > minAllowed.getTime() ? oneHourBefore : minAllowed).toISOString())
  }

  const normalizeDatePickerValueToUtc = (value: string): string => {
    const localToday = getTodayLocalDate()
    const utcToday = getTodayUtcDate()

    // Native picker "Today" follows local timezone. In UTC mode, remap to UTC today when they differ.
    if (value === localToday && localToday !== utcToday) {
      return utcToday
    }

    return value
  }

  const isPastEvent = (event: Event): boolean => {
    const parsed = new Date(event.date)
    if (isNaN(parsed.getTime())) return false
    return parsed.getTime() <= Date.now()
  }

  const openReminderModal = (event: Event) => {
    const existing = remindersByEvent[event.id]
    const initialValue = existing ? toDateTimeUtcInput(existing.remindAt) : getDefaultReminderDateTimeUtcInput(event)
    const split = splitUtcInput(initialValue)
    setModalEvent(event)
    setModalMessage(existing?.message || '')
    setModalRemindDate(split.date)
    setModalRemindTime(split.time)
  }

  const closeReminderModal = () => {
    setModalEvent(null)
    setModalMessage('')
    setModalRemindDate('')
    setModalRemindTime('')
    setModalSaving(false)
  }

  const handleSaveReminder = async () => {
    if (!modalEvent || !modalRemindDate || !modalRemindTime) return
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }

    setModalSaving(true)
    try {
      const existing = remindersByEvent[modalEvent.id]
      if (existing) {
        await deleteReminder(existing.id, token)
      }

      const saved = await setReminder(
        {
          eventId: modalEvent.id,
          message: modalMessage.trim(),
          remindAt: utcInputToIso(toUtcInputFromParts(modalRemindDate, modalRemindTime)),
        },
        token
      )

      setRemindersByEvent((prev) => ({ ...prev, [modalEvent.id]: saved }))
      closeReminderModal()
    } catch (error) {
      console.error('Failed to set reminder:', error)
    } finally {
      setModalSaving(false)
    }
  }

  const handleDeleteReminder = async (eventId: string) => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }
    const reminder = remindersByEvent[eventId]
    if (!reminder) return

    setRemindingId(eventId)
    try {
      await deleteReminder(reminder.id, token)
      setRemindersByEvent((prev) => {
        const next = { ...prev }
        delete next[eventId]
        return next
      })
    } catch (error) {
      console.error('Failed to delete reminder:', error)
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
                  (() => {
                    const pastEvent = isPastEvent(event)
                    return (
                  <button 
                    className={styles.btnRemind}
                    onClick={() => openReminderModal(event)}
                    disabled={remindingId === event.id || pastEvent}
                  >
                    {remindingId === event.id
                      ? '⏳ Saving...'
                      : pastEvent
                        ? 'Event Ended'
                        : remindersByEvent[event.id]
                          ? 'Edit Reminder'
                          : 'Set Reminder'}
                  </button>
                    )
                  })()
                )}
                {!event.isLive && remindersByEvent[event.id] ? (
                  <button
                    className={styles.btnDelete}
                    onClick={() => handleDeleteReminder(event.id)}
                    disabled={remindingId === event.id}
                    aria-label="Delete reminder"
                  >
                    🗑️
                  </button>
                ) : null}
              </div>
            </div>
          ))}
        </div>

        {modalEvent ? (
          <div className={styles.reminderModalOverlay} onClick={closeReminderModal}>
            <div className={styles.reminderModal} onClick={(e) => e.stopPropagation()}>
              <div className={styles.reminderModalTitle}>Reminder for {modalEvent.title}</div>
              <label className={styles.reminderFieldLabel}>Reminder Date (UTC)</label>
              <input
                className={styles.reminderInput}
                type="date"
                value={modalRemindDate}
                onChange={(e) => setModalRemindDate(normalizeDatePickerValueToUtc(e.target.value))}
              />

              <label className={styles.reminderFieldLabel}>Reminder Time (UTC)</label>
              <input
                className={styles.reminderInput}
                type="time"
                value={modalRemindTime}
                onChange={(e) => setModalRemindTime(e.target.value)}
              />

              <label className={styles.reminderFieldLabel}>Message (optional)</label>
              <textarea
                className={styles.reminderTextarea}
                value={modalMessage}
                onChange={(e) => setModalMessage(e.target.value)}
                placeholder="Ex: Leave for track at 6:30 PM"
              />

              <div className={styles.reminderActions}>
                <button className={styles.btnCancel} onClick={closeReminderModal} disabled={modalSaving}>Cancel</button>
                <button className={styles.btnSave} onClick={handleSaveReminder} disabled={modalSaving || !modalRemindDate || !modalRemindTime}>
                  {modalSaving ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </main>
    </>
  )
}
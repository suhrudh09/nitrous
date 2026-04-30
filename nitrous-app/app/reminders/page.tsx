'use client'

import { useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import Nav from '@/components/Nav'
import { deleteReminder, getEvents, getMyReminders } from '@/lib/api'
import type { Reminder, Event } from '@/types'
import styles from './reminders.module.css'

export default function RemindersPage() {
  const [reminders, setReminders] = useState<Reminder[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(true)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }

    let timer: ReturnType<typeof setInterval> | undefined
    let cancelled = false

    const syncReminders = async () => {
      try {
        const reminderData = await getMyReminders(token)
        if (cancelled) return
        setReminders(reminderData)
      } catch {
        // ignore periodic reminder sync failures
      }
    }

    Promise.all([getMyReminders(token), getEvents().catch(() => [])])
      .then(([reminderData, eventData]) => {
        setReminders(reminderData)
        setEvents(eventData)
        timer = setInterval(syncReminders, 15000)
      })
      .finally(() => setLoading(false))

    return () => {
      cancelled = true
      if (timer) clearInterval(timer)
    }
  }, [])

  const eventById = useMemo(() => {
    const map: Record<string, Event> = {}
    events.forEach((event) => {
      map[event.id] = event
    })
    return map
  }, [events])

  const formatDateTime = (value: string) => {
    const d = new Date(value)
    if (isNaN(d.getTime())) return value
    return d.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'UTC',
      timeZoneName: 'short',
    })
  }

  const formatUtcTime = (value?: string) => {
    if (!value) return 'TBA'
    const parsed = new Date(value)
    if (isNaN(parsed.getTime())) return `${value} UTC`
    return parsed.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'UTC',
      timeZoneName: 'short',
    })
  }

  const formatEventSchedule = (event?: Event) => {
    if (!event) return 'Date/time unavailable'

    const rawDate = event.date || ''
    const dateOnly = rawDate.includes('T') ? rawDate.split('T')[0] : rawDate
    const parsedDate = new Date(`${dateOnly}T00:00:00Z`)
    const dateLabel = isNaN(parsedDate.getTime())
      ? rawDate || 'Date TBA'
      : parsedDate.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
          year: 'numeric',
          timeZone: 'UTC',
        })

    return event.time ? `${dateLabel} • ${formatUtcTime(event.time)}` : `${dateLabel} • TBA`
  }

  const handleDelete = async (id: string) => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) return

    setDeletingId(id)
    try {
      await deleteReminder(id, token)
      setReminders((prev) => prev.filter((reminder) => reminder.id !== id))
    } finally {
      setDeletingId(null)
    }
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>
        <div className={styles.pageHeader}>
          <div className={styles.headerTag}>/ REMINDERS</div>
          <h1 className={styles.pageTitle}>MY REMINDERS</h1>
          <p className={styles.pageSubtitle}>Manage your saved event reminders.</p>
        </div>

        {loading ? (
          <div className={styles.loading}>Loading reminders...</div>
        ) : reminders.length === 0 ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>⏰</div>
            <h2>No reminders set</h2>
            <p>Set reminders on the events page and they will show up here.</p>
            <Link href="/events" className={styles.ctaBtn}>Browse Events</Link>
          </div>
        ) : (
          <div className={styles.list}>
            {reminders.map((reminder) => {
              const event = eventById[reminder.eventId]
              return (
                <div key={reminder.id} className={styles.card}>
                  <div className={styles.cardHead}>
                    <div>
                      <div className={styles.eventTitle}>{event?.title || 'Event'}</div>
                      <div className={styles.eventMeta}>{event?.location || 'Location TBA'}</div>
                    </div>
                    <button
                      className={styles.deleteBtn}
                      onClick={() => handleDelete(reminder.id)}
                      disabled={deletingId === reminder.id}
                    >
                      {deletingId === reminder.id ? 'Deleting...' : '🗑️'}
                    </button>
                  </div>

                  <div className={styles.remindAt}>Event: {formatEventSchedule(event)}</div>
                  <div className={styles.remindAt}>Reminder at: {formatDateTime(reminder.remindAt)}</div>
                  {reminder.message ? <div className={styles.message}>{reminder.message}</div> : null}
                  <div className={styles.cardActions}>
                    <Link href="/events" className={styles.editLink}>Edit on Events</Link>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </main>
    </>
  )
}

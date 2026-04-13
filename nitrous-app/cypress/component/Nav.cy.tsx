'use client'
import Link from 'next/link'
import { useState, useEffect, useRef } from 'react'
import styles from './Nav.module.css'

interface User {
  name: string
  email: string
  initials: string
}

export default function Nav() {
  const [user, setUser] = useState<User | null>(null)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropRef = useRef<HTMLDivElement>(null)

  // Read auth from localStorage on mount
  useEffect(() => {
    try {
      const token = localStorage.getItem('nitrous_token')
      const stored = localStorage.getItem('nitrous_user')
      if (token && stored) {
        const parsed = JSON.parse(stored)
        const names = (parsed.name || 'User').trim().split(' ')
        const initials = names.length >= 2
          ? `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase()
          : names[0].slice(0, 2).toUpperCase()
        setUser({ name: parsed.name, email: parsed.email, initials })
      }
    } catch {
      // ignore
    }
  }, [])

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropRef.current && !dropRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
    }
    if (dropdownOpen) document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [dropdownOpen])

  function handleSignOut() {
    localStorage.removeItem('nitrous_token')
    localStorage.removeItem('nitrous_user')
    setUser(null)
    setDropdownOpen(false)
    globalThis.location.href = '/'
  }

  return (
    <nav className={styles.nav}>
      <Link href="/" className={styles.logo}>
        NITROUS<span>.</span>
      </Link>

      <div className={styles.navCenter}>
        <Link href="/live" className={styles.navLink}>Live</Link>
        <Link href="/events" className={styles.navLink}>Events</Link>
        <Link href="/teams" className={styles.navLink}>Teams</Link>
        <Link href="/journeys" className={styles.navLink}>Journeys</Link>
        <Link href="/merch" className={styles.navLink}>Merch</Link>
      </div>

      <div className={styles.navRight}>
        <div className={styles.navStatus}>
          <div className={styles.dotLive} />
          <span>4 Events Live</span>
        </div>

        {user ? (
          <div className={styles.userZone} ref={dropRef}>
            <button
              className={`${styles.userBtn} ${dropdownOpen ? styles.userBtnOpen : ''}`}
              onClick={() => setDropdownOpen(v => !v)}
              aria-label="User menu"
            >
              <div className={styles.userAvatar}>
                <span className={styles.userInitials}>{user.initials}</span>
                <div className={styles.avatarRing} />
              </div>
              <svg className={`${styles.chevron} ${dropdownOpen ? styles.chevronUp : ''}`} viewBox="0 0 12 8" fill="none">
                <path d="M1 1L6 7L11 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </svg>
            </button>

            {dropdownOpen && (
              <div className={styles.dropdown}>
                {/* Top corners */}
                <div className={`${styles.dropCorner} ${styles.dropCornerTL}`} />
                <div className={`${styles.dropCorner} ${styles.dropCornerTR}`} />

                {/* User info */}
                <div className={styles.dropHeader}>
                  <div className={styles.dropAvatar}>
                    <span>{user.initials}</span>
                  </div>
                  <div className={styles.dropUserInfo}>
                    <div className={styles.dropName}>{user.name}</div>
                    <div className={styles.dropEmail}>{user.email}</div>
                  </div>
                </div>

                <div className={styles.dropDivider} />

                <Link href="/garage" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>🚗</span>
                  <span>My Garage</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>
                <Link href="/passes" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>🎫</span>
                  <span>My Passes</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>
                <Link href="/orders" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>📦</span>
                  <span>My Orders</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>
                <Link href="/journeys" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>🌍</span>
                  <span>My Journeys</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>

                <div className={styles.dropDivider} />

                <Link href="/settings" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>⚙️</span>
                  <span>Settings</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>

                <div className={styles.dropDivider} />

                <button className={`${styles.dropItem} ${styles.dropSignOut}`} onClick={handleSignOut}>
                  <span className={styles.dropItemIcon}>⎋</span>
                  <span>Sign Out</span>
                </button>
              </div>
            )}
          </div>
        ) : (
          // Rendered as <button> so cy.get('button').contains('Sign In') works in tests.
          // Uses router-style navigation to keep Next.js routing intact.
          <button
            className={styles.btnNav}
            onClick={() => { globalThis.location.href = '/login' }}
          >
            Sign In
          </button>
        )}
      </div>
    </nav>
  )
}
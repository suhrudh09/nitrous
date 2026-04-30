'use client'
import Link from 'next/link'
import { useState, useEffect, useRef } from 'react'
import { getCart, getLiveEvents, getNotifications, markNotificationRead } from '@/lib/api'
import type { Notification } from '@/types'
import styles from './Nav.module.css'

interface User {
  name: string
  email: string
  initials: string
  role: 'viewer' | 'participant' | 'manager' | 'sponsor' | 'admin'
}

interface CartItem {
  item: {
    id: string
    name: string
    icon: string
    price: number
    category: string
  }
  quantity: number
  size?: string
}

const CART_STORAGE_KEY = 'nitrous_cart_v1'
const CART_UPDATED_EVENT = 'nitrous-cart-updated'

export default function Nav() {
  const [user, setUser] = useState<User | null>(null)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [cartOpen, setCartOpen] = useState(false)
  const [notificationsOpen, setNotificationsOpen] = useState(false)
  const [cart, setCart] = useState<CartItem[]>([])
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [liveEventCount, setLiveEventCount] = useState(0)
  const [showNewNotificationPopup, setShowNewNotificationPopup] = useState(false)
  const dropRef = useRef<HTMLDivElement>(null)
  const cartRef = useRef<HTMLDivElement>(null)
  const notificationsRef = useRef<HTMLDivElement>(null)
  const prevUnreadRef = useRef(0)

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
        setUser({ name: parsed.name, email: parsed.email, initials, role: parsed.role || 'viewer' })
      }
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setInterval> | undefined

    const fetchLiveEvents = async () => {
      try {
        const liveEvents = await getLiveEvents()
        if (cancelled) return
        setLiveEventCount(liveEvents.length)
      } catch {
        // keep the previous live count on transient failures
      }
    }

    fetchLiveEvents()
    timer = setInterval(fetchLiveEvents, 30000)

    return () => {
      cancelled = true
      if (timer) clearInterval(timer)
    }
  }, [])

  // Load cart from localStorage
  useEffect(() => {
    let cancelled = false

    const loadLocalCart = () => {
      try {
        const raw = localStorage.getItem(CART_STORAGE_KEY)
        if (!raw) {
          setCart([])
          return
        }
        const parsed = JSON.parse(raw)
        if (Array.isArray(parsed)) {
          setCart(parsed)
        }
      } catch {
        setCart([])
      }
    }

    const loadCart = async () => {
      loadLocalCart()

      const token = localStorage.getItem('nitrous_token')
      if (!token) return

      try {
        const remote = await getCart(token)
        if (cancelled) return

        const normalized = remote.map((entry) => ({
          item: {
            id: entry.merchId,
            name: entry.name,
            icon: entry.icon,
            price: entry.price,
            category: entry.category,
          },
          quantity: entry.quantity,
          size: entry.size || undefined,
        }))
        setCart(normalized)
      } catch {
        // Keep local cart when backend sync fails
      }
    }

    const onStorage = (event: StorageEvent) => {
      if (event.key === CART_STORAGE_KEY) {
        loadLocalCart()
      }
    }

    const onCartUpdated = () => {
      loadLocalCart()
    }

    loadCart()
    window.addEventListener('storage', onStorage)
    window.addEventListener(CART_UPDATED_EVENT, onCartUpdated)

    return () => {
      cancelled = true
      window.removeEventListener('storage', onStorage)
      window.removeEventListener(CART_UPDATED_EVENT, onCartUpdated)
    }
  }, [])

  // Close dropdowns on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropRef.current && !dropRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
      if (cartRef.current && !cartRef.current.contains(e.target as Node)) {
        setCartOpen(false)
      }
      if (notificationsRef.current && !notificationsRef.current.contains(e.target as Node)) {
        setNotificationsOpen(false)
      }
    }
    if (dropdownOpen || cartOpen || notificationsOpen) document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [dropdownOpen, cartOpen, notificationsOpen])

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setInterval> | undefined
    const token = localStorage.getItem('nitrous_token')
    if (!token) return

    const fetchNotifications = async () => {
      try {
        const data = await getNotifications(token)
        if (cancelled) return

        setNotifications(data.notifications)
        setUnreadCount(data.unread)

        if (data.unread > prevUnreadRef.current) {
          setShowNewNotificationPopup(true)
          window.setTimeout(() => setShowNewNotificationPopup(false), 4000)
        }
        prevUnreadRef.current = data.unread
      } catch {
        // ignore notification polling failures
      }
    }

    fetchNotifications()
    timer = setInterval(fetchNotifications, 15000)

    return () => {
      cancelled = true
      if (timer) clearInterval(timer)
    }
  }, [])

  const onMarkNotificationRead = async (notificationId: string) => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) return

    try {
      await markNotificationRead(notificationId, token)
      setNotifications((prev) =>
        prev.map((notification) =>
          notification.id === notificationId
            ? { ...notification, readAt: new Date().toISOString() }
            : notification
        )
      )
      setUnreadCount((prev) => Math.max(0, prev - 1))
      prevUnreadRef.current = Math.max(0, prevUnreadRef.current - 1)
    } catch {
      // ignore mark read errors
    }
  }

  function handleSignOut() {
    localStorage.removeItem('nitrous_token')
    localStorage.removeItem('nitrous_user')
    localStorage.removeItem(CART_STORAGE_KEY)
    window.dispatchEvent(new Event(CART_UPDATED_EVENT))
    setCart([])
    setUser(null)
    setDropdownOpen(false)
    globalThis.location.href = '/'
  }

  const cartCount = cart.reduce((sum, item) => sum + item.quantity, 0)
  const cartTotal = cart.reduce((sum, item) => sum + item.item.price * item.quantity, 0)

  return (
    <nav className={styles.nav}>
      {showNewNotificationPopup ? (
        <div className={styles.notificationPopup}>New notification</div>
      ) : null}

      <Link href="/" className={styles.logo}>
        NITROUS<span>.</span>
      </Link>

      <div className={styles.navCenter}>
        <Link href="/garage" className={styles.navLink}>Garage</Link>
        <Link href="/passes" className={styles.navLink}>Access Passes</Link>
        <Link href="/live" className={styles.navLink}>Live</Link>
        <Link href="/events" className={styles.navLink}>Events</Link>
        <Link href="/teams" className={styles.navLink}>Teams</Link>
        <Link href="/journeys" className={styles.navLink}>Journeys</Link>
        <Link href="/merch" className={styles.navLink}>Merch</Link>
      </div>

      <div className={styles.navRight}>
        <div className={styles.navStatus}>
          <div className={styles.dotLive} />
          <span>{liveEventCount} {liveEventCount === 1 ? 'Event' : 'Events'} Live</span>
        </div>

        <div className={styles.notificationZone} ref={notificationsRef}>
          <button
            className={`${styles.notificationBtn} ${notificationsOpen ? styles.notificationBtnOpen : ''}`}
            onClick={() => setNotificationsOpen((v) => !v)}
            aria-label="Notifications"
          >
            🔔
            {unreadCount > 0 ? <span className={styles.notificationBadge}>{unreadCount}</span> : null}
          </button>

          {notificationsOpen ? (
            <div className={styles.notificationDropdown}>
              <div className={`${styles.dropCorner} ${styles.dropCornerTL}`} />
              <div className={`${styles.dropCorner} ${styles.dropCornerTR}`} />
              <div className={styles.notificationHeader}>
                <span>Notifications</span>
                <span>{unreadCount} new</span>
              </div>
              <div className={styles.cartDivider} />

              <div className={styles.notificationList}>
                {notifications.length === 0 ? (
                  <div className={styles.notificationEmpty}>No notifications</div>
                ) : (
                  notifications.map((notification) => (
                    <div key={notification.id} className={styles.notificationItem}>
                      <div className={styles.notificationTitle}>{notification.title}</div>
                      <div className={styles.notificationBody}>{notification.body}</div>
                      {notification.readAt ? (
                        <div className={styles.notificationRead}>Read</div>
                      ) : (
                        <button
                          className={styles.notificationMarkRead}
                          onClick={() => onMarkNotificationRead(notification.id)}
                        >
                          Mark read
                        </button>
                      )}
                    </div>
                  ))
                )}
              </div>
            </div>
          ) : null}
        </div>

        {/* Cart Icon */}
        <div className={styles.cartZone} ref={cartRef}>
          <button
            className={`${styles.cartBtn} ${cartOpen ? styles.cartBtnOpen : ''}`}
            onClick={() => setCartOpen(v => !v)}
            aria-label="Shopping cart"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M6 6h15l-1.5 9h-12L6 6z" />
              <path d="M6 6L5 3H2" />
              <circle cx="9" cy="20" r="1.5" />
              <circle cx="18" cy="20" r="1.5" />
            </svg>
            {cartCount > 0 && (
              <span className={styles.cartBadge}>{cartCount}</span>
            )}
          </button>

          {cartOpen && (
            <div className={styles.cartDropdown}>
              <div className={`${styles.dropCorner} ${styles.dropCornerTL}`} />
              <div className={`${styles.dropCorner} ${styles.dropCornerTR}`} />

              <div className={styles.cartHeader}>
                <span>Your Cart</span>
                <span className={styles.cartCount}>{cartCount} items</span>
              </div>

              <div className={styles.cartDivider} />

              {cart.length === 0 ? (
                <div className={styles.cartEmpty}>
                  <span>🛒</span>
                  <p>Your cart is empty</p>
                </div>
              ) : (
                <>
                  <div className={styles.cartItems}>
                    {cart.slice(0, 4).map((entry, idx) => (
                      <div key={idx} className={styles.cartItem}>
                        <span className={styles.cartItemIcon}>{entry.item.icon}</span>
                        <div className={styles.cartItemInfo}>
                          <span className={styles.cartItemName}>{entry.item.name}</span>
                          <span className={styles.cartItemQty}>x{entry.quantity}</span>
                        </div>
                        <span className={styles.cartItemPrice}>${entry.item.price * entry.quantity}</span>
                      </div>
                    ))}
                    {cart.length > 4 && (
                      <div className={styles.cartMore}>+{cart.length - 4} more items</div>
                    )}
                  </div>

                  <div className={styles.cartDivider} />

                  <div className={styles.cartFooter}>
                    <div className={styles.cartTotal}>
                      <span>Total</span>
                      <span>${cartTotal.toFixed(2)}</span>
                    </div>
                    <Link
                      href="/cart"
                      className={styles.cartCta}
                      onClick={() => setCartOpen(false)}
                    >
                      View Cart →
                    </Link>
                  </div>
                </>
              )}
            </div>
          )}
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
                <Link href="/reminders" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>⏰</span>
                  <span>My Reminders</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>
                <Link href="/journeys" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                  <span className={styles.dropItemIcon}>🌍</span>
                  <span>My Journeys</span>
                  <span className={styles.dropItemArrow}>→</span>
                </Link>

                <div className={styles.dropDivider} />

                {user?.role === 'admin' && (
                  <Link href="/admin" className={styles.dropItem} onClick={() => setDropdownOpen(false)}>
                    <span className={styles.dropItemIcon}>⚡</span>
                    <span>Admin Panel</span>
                    <span className={styles.dropItemArrow}>→</span>
                  </Link>
                )}

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
          <Link href="/login" className={styles.btnNav}>Sign In</Link>
        )}
      </div>
    </nav>
  )
}
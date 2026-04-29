'use client'
import Link from 'next/link'
import { useState, useEffect, useRef } from 'react'
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

export default function Nav() {
  const [user, setUser] = useState<User | null>(null)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [cartOpen, setCartOpen] = useState(false)
  const [cart, setCart] = useState<CartItem[]>([])
  const dropRef = useRef<HTMLDivElement>(null)
  const cartRef = useRef<HTMLDivElement>(null)

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

  // Load cart from localStorage
  useEffect(() => {
    try {
      const raw = localStorage.getItem(CART_STORAGE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw)
        if (Array.isArray(parsed)) {
          setCart(parsed)
        }
      }
    } catch {
      // ignore
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
    }
    if (dropdownOpen || cartOpen) document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [dropdownOpen, cartOpen])

  function handleSignOut() {
    localStorage.removeItem('nitrous_token')
    localStorage.removeItem('nitrous_user')
    setUser(null)
    setDropdownOpen(false)
    globalThis.location.href = '/'
  }

  const cartCount = cart.reduce((sum, item) => sum + item.quantity, 0)
  const cartTotal = cart.reduce((sum, item) => sum + item.item.price * item.quantity, 0)

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
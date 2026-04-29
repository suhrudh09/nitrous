'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getMerchItems, createOrder } from '@/lib/api'
import type { MerchItem, OrderItem } from '@/types'
import styles from './merch.module.css'

const cats = ['all', 'apparel', 'accessories', 'collectibles']

const tagColors: Record<string, string> = {
  BESTSELLER: 'var(--cyan)',
  NEW: '#4ade80',
  LIMITED: 'var(--red)',
}

// Local cart item tracks quantity + optional size alongside the merch item
interface CartEntry {
  item: MerchItem
  quantity: number
  size?: string
}

const CART_STORAGE_KEY = 'nitrous_cart_v1'

export default function MerchPage() {
  const [products, setProducts] = useState<MerchItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')

  // cart stores CartEntry objects so we can pass correctly-typed OrderItem[] to createOrder
  const [cart, setCart] = useState<CartEntry[]>([])
  const [added, setAdded] = useState<string | null>(null)
  const [checkoutLoading, setCheckoutLoading] = useState(false)
  const [checkoutMsg, setCheckoutMsg] = useState('')
  const [showSizeModal, setShowSizeModal] = useState(false)
  const [selectedItemForSize, setSelectedItemForSize] = useState<MerchItem | null>(null)

  const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL']

  const persistCart = (nextCart: CartEntry[]) => {
    try {
      localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(nextCart))
    } catch {
      // Ignore persistence errors (quota/private mode), keep cart in memory.
    }
  }

  useEffect(() => {
    getMerchItems()
      .then(setProducts)
      .catch(() => setError('Could not load products'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    try {
      const raw = localStorage.getItem(CART_STORAGE_KEY)
      if (!raw) return

      const parsed = JSON.parse(raw)
      if (!Array.isArray(parsed)) return

      const hydrated: CartEntry[] = parsed
        .filter((entry) => entry && entry.item && typeof entry.item.id === 'string')
        .map((entry) => ({
          item: entry.item as MerchItem,
          quantity:
            typeof entry.quantity === 'number' && entry.quantity > 0
              ? Math.floor(entry.quantity)
              : 1,
          size: typeof entry.size === 'string' ? entry.size : undefined,
        }))

      setCart(hydrated)
    } catch {
      localStorage.removeItem(CART_STORAGE_KEY)
    }
  }, [])

  useEffect(() => {
    if (products.length === 0) return

    setCart((prev) => {
      if (prev.length === 0) return prev

      const byId = new Map(products.map((p) => [p.id, p]))
      const byName = new Map(products.map((p) => [p.name.toLowerCase(), p]))

      const reconciled = prev
        .map((entry) => {
          const direct = byId.get(entry.item.id)
          if (direct) {
            return { ...entry, item: direct }
          }

          const fallback = byName.get(entry.item.name.toLowerCase())
          if (!fallback) return null

          return { ...entry, item: fallback }
        })
        .filter((entry): entry is CartEntry => entry !== null)

      persistCart(reconciled)
      return reconciled
    })
  }, [products])

  const filtered =
    filter === 'all' ? products : products.filter((p) => p.category === filter)

  function addToCart(item: MerchItem) {
    // Show size modal for apparel items
    if (item.category === 'apparel') {
      setSelectedItemForSize(item)
      setShowSizeModal(true)
      return
    }

    // For non-apparel items, add directly
    addToCartWithSize(item, undefined)
  }

  function addToCartWithSize(item: MerchItem, size?: string) {
    setCart((prev) => {
      const existing = prev.find((e) => e.item.id === item.id && e.size === size)
      let next: CartEntry[]
      if (existing) {
        next = prev.map((e) =>
          e.item.id === item.id && e.size === size ? { ...e, quantity: e.quantity + 1 } : e
        )
      } else {
        next = [...prev, { item, quantity: 1, size }]
      }
      persistCart(next)
      return next
    })
    setAdded(item.id)
    setTimeout(() => setAdded(null), 1200)
  }

  function handleSizeSelect(size: string) {
    if (selectedItemForSize) {
      addToCartWithSize(selectedItemForSize, size)
      setShowSizeModal(false)
      setSelectedItemForSize(null)
    }
  }

  async function handleCheckout() {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      globalThis.location.href = '/login'
      return
    }
    if (cart.length === 0) return

    // Build correctly-typed OrderItem[] for the API
    const orderItems: OrderItem[] = cart.map((entry) => ({
      merchId: entry.item.id,
      name: entry.item.name,
      price: entry.item.price,
      quantity: entry.quantity,
      size: entry.size,
    }))

    setCheckoutLoading(true)
    setCheckoutMsg('')
    try {
      const result = await createOrder(orderItems, token)
      setCart([])
      persistCart([])
      setCheckoutMsg(`✓ Order #${result.order.id.slice(0, 8).toUpperCase()} placed — total $${result.order.total}`)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Checkout failed'
      setCheckoutMsg(`✗ ${message}`)
    } finally {
      setCheckoutLoading(false)
    }
  }

  const totalItems = cart.reduce((acc, e) => acc + e.quantity, 0)

  if (loading) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--muted)', fontFamily: 'var(--font-mono)' }}>
            LOADING PRODUCTS...
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
        {/* Header */}
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ MERCH</div>
            <h1 className={styles.pageTitle}>TEAM STORE</h1>
            <p className={styles.pageSubtitle}>Official gear, limited drops, and collector pieces</p>
          </div>
          <div className={styles.cartBadge} onClick={handleCheckout} style={{ cursor: totalItems > 0 ? 'pointer' : 'default' }}>
            <span className={styles.cartIcon}>🛒</span>
            <span className={styles.cartCount}>{totalItems}</span>
            <span className={styles.cartLabel}>{checkoutLoading ? 'PLACING...' : 'CART'}</span>
          </div>
        </div>

        {checkoutMsg && (
          <div
            style={{
              padding: '12px 48px',
              fontFamily: 'var(--font-mono)',
              fontSize: '13px',
              color: checkoutMsg.startsWith('✓') ? 'var(--cyan)' : 'var(--red)',
            }}
          >
            {checkoutMsg}
          </div>
        )}

        {/* Filter tabs */}
        <div className={styles.filterBar}>
          {cats.map((cat) => (
            <button
              key={cat}
              className={`${styles.catBtn} ${filter === cat ? styles.catBtnActive : ''}`}
              onClick={() => setFilter(cat)}
            >
              {cat === 'all' ? 'ALL PRODUCTS' : cat.toUpperCase()}
            </button>
          ))}
          <div className={styles.filterRight}>
            <span className={styles.countTxt}>{filtered.length} items</span>
          </div>
        </div>

        {/* Products Grid */}
        <div className={styles.productsGrid}>
          {filtered.map((product) => (
            <div key={product.id} className={styles.productCard}>
              {/* Product visual */}
              <div className={styles.productVisual}>
                <div className={styles.productIcon}>{product.icon}</div>
              </div>

              <div className={styles.productInfo}>
                <div className={styles.productCat}>{product.category}</div>
                <div className={styles.productName}>{product.name}</div>
              </div>

              <div className={styles.productBottom}>
                <div className={styles.productPrice}>${product.price}</div>
                <button
                  className={`${styles.addBtn} ${added === product.id ? styles.addBtnSuccess : ''}`}
                  onClick={() => addToCart(product)}
                >
                  {added === product.id ? '✓ ADDED' : '+ ADD TO CART'}
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Bottom banner */}
        <div className={styles.promoBanner}>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>MEMBERS GET 15% OFF</span>
            <span className={styles.promoCta}>Sign in or create an account →</span>
          </div>
          <div className={styles.promoDivider}></div>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>FREE SHIPPING OVER $100</span>
            <span className={styles.promoCta}>Worldwide delivery available</span>
          </div>
          <div className={styles.promoDivider}></div>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>LIMITED DROPS</span>
            <span className={styles.promoCta}>Subscribe for early access</span>
          </div>
        </div>

        {/* Size Selection Modal */}
        {showSizeModal && selectedItemForSize && (
          <div className={styles.sizeModalOverlay} onClick={() => setShowSizeModal(false)}>
            <div className={styles.sizeModal} onClick={(e) => e.stopPropagation()}>
              <button className={styles.sizeModalClose} onClick={() => setShowSizeModal(false)}>×</button>
              <div className={styles.sizeModalHeader}>
                <span className={styles.sizeModalIcon}>{selectedItemForSize.icon}</span>
                <div>
                  <div className={styles.sizeModalName}>{selectedItemForSize.name}</div>
                  <div className={styles.sizeModalPrice}>${selectedItemForSize.price}</div>
                </div>
              </div>
              <div className={styles.sizeModalBody}>
                <div className={styles.sizeModalLabel}>Select Size</div>
                <div className={styles.sizeGrid}>
                  {SIZES.map((size) => (
                    <button
                      key={size}
                      className={styles.sizeBtn}
                      onClick={() => handleSizeSelect(size)}
                    >
                      {size}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </>
  )
}

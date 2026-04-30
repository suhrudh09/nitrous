'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Nav from '@/components/Nav'
import { getMerchItems, getCart, saveCart } from '@/lib/api'
import type { MerchItem, CartItem as APICartItem } from '@/types'
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
const CART_UPDATED_EVENT = 'nitrous-cart-updated'

export default function MerchPage() {
  const router = useRouter()
  const [products, setProducts] = useState<MerchItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')

  // cart stores CartEntry objects so we can pass correctly-typed OrderItem[] to createOrder
  const [cart, setCart] = useState<CartEntry[]>([])
  const [added, setAdded] = useState<string | null>(null)
  const [selectedSizes, setSelectedSizes] = useState<Record<string, string>>({})

  const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL']

  const toAPICartItems = (entries: CartEntry[]): APICartItem[] =>
    entries.map((entry) => ({
      merchId: entry.item.id,
      name: entry.item.name,
      icon: entry.item.icon,
      price: entry.item.price,
      category: entry.item.category,
      quantity: entry.quantity,
      size: entry.size,
    }))

  const persistCart = (nextCart: CartEntry[]) => {
    const token = localStorage.getItem('nitrous_token')

    try {
      // Guest carts persist in localStorage; authenticated carts live on the backend only.
      if (!token) {
        localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(nextCart))
      }
      window.dispatchEvent(new Event(CART_UPDATED_EVENT))
    } catch {
      // Ignore persistence errors (quota/private mode), keep cart in memory.
    }

    if (!token) return
    saveCart(toAPICartItems(nextCart), token).catch(() => {
      // Keep local state if backend cart sync fails.
    })
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

    const token = localStorage.getItem('nitrous_token')
    if (!token) return

    getCart(token)
      .then((items) => {
        const remoteCart: CartEntry[] = items.map((entry) => ({
          item: {
            id: entry.merchId,
            name: entry.name,
            icon: entry.icon,
            price: entry.price,
            category: entry.category as MerchItem['category'],
          },
          quantity: entry.quantity,
          size: entry.size || undefined,
        }))
        setCart(remoteCart)
      })
      .catch(() => {
        // Keep local cart when backend cart load fails.
      })
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

  function getSelectedSize(item: MerchItem): string | undefined {
    if (item.category !== 'apparel') return undefined
    return selectedSizes[item.id] ?? SIZES[0]
  }

  function addToCartWithSize(item: MerchItem, size?: string, quantityToAdd = 1) {
    const safeQuantity = Math.max(1, Math.floor(quantityToAdd))
    const existing = cart.find((e) => e.item.id === item.id && e.size === size)
    let next: CartEntry[]
    if (existing) {
      next = cart.map((e) =>
        e.item.id === item.id && e.size === size ? { ...e, quantity: e.quantity + safeQuantity } : e
      )
    } else {
      next = [...cart, { item, quantity: safeQuantity, size }]
    }

    setCart(next)
    persistCart(next)
    setAdded(item.id)
    setTimeout(() => setAdded(null), 1200)
  }

  function updateCartQuantity(item: MerchItem, size: string | undefined, delta: number) {
    const existing = cart.find((e) => e.item.id === item.id && e.size === size)
    if (!existing) return

    const nextQuantity = existing.quantity + delta
    let next: CartEntry[]
    if (nextQuantity <= 0) {
      next = cart.filter((e) => !(e.item.id === item.id && e.size === size))
    } else {
      next = cart.map((e) =>
        e.item.id === item.id && e.size === size ? { ...e, quantity: nextQuantity } : e
      )
    }

    setCart(next)
    persistCart(next)
  }

  function setCardSize(productId: string, size: string) {
    setSelectedSizes((prev) => ({ ...prev, [productId]: size }))
  }

  function getCartQuantity(item: MerchItem, size: string | undefined): number {
    const entry = cart.find((e) => e.item.id === item.id && e.size === size)
    return entry?.quantity ?? 0
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
          <div className={styles.cartBadge} onClick={() => router.push('/cart')} style={{ cursor: 'pointer' }}>
            <span className={styles.cartIcon}>🛒</span>
            <span className={styles.cartCount}>{totalItems}</span>
            <span className={styles.cartLabel}>PLACE ORDER</span>
          </div>
        </div>

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
          {filtered.map((product) => {
            const selectedSize = getSelectedSize(product)
            const quantityInCart = getCartQuantity(product, selectedSize)

            return <div key={product.id} className={styles.productCard}>
              {/* Product visual */}
              <div className={styles.productVisual}>
                <div className={styles.productIcon}>{product.icon}</div>
              </div>

              <div className={styles.productInfo}>
                <div className={styles.productCat}>{product.category}</div>
                <div className={styles.productName}>{product.name}</div>
                {product.category === 'apparel' ? (
                  <div className={styles.sizes}>
                    {SIZES.map((size) => {
                      const active = (selectedSizes[product.id] ?? SIZES[0]) === size
                      return (
                        <button
                          key={size}
                          type="button"
                          className={`${styles.sizeChip} ${active ? styles.sizeChipActive : ''}`}
                          onClick={() => setCardSize(product.id, size)}
                        >
                          {size}
                        </button>
                      )
                    })}
                  </div>
                ) : null}
              </div>

              <div className={styles.productBottom}>
                <div className={styles.productPrice}>${product.price}</div>
                {quantityInCart > 0 ? (
                  <div className={styles.itemQuantity}>
                    <button
                      type="button"
                      className={styles.qtyBtn}
                      onClick={() => updateCartQuantity(product, selectedSize, -1)}
                    >
                      -
                    </button>
                    <span className={styles.qtyValue}>{quantityInCart}</span>
                    <button
                      type="button"
                      className={styles.qtyBtn}
                      onClick={() => updateCartQuantity(product, selectedSize, 1)}
                    >
                      +
                    </button>
                  </div>
                ) : (
                  <button
                    className={`${styles.addBtn} ${added === product.id ? styles.addBtnSuccess : ''}`}
                    onClick={() => addToCartWithSize(product, selectedSize, 1)}
                  >
                    {added === product.id ? '✓ ADDED' : '+ ADD TO CART'}
                  </button>
                )}
              </div>
            </div>
          })}
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
      </main>
    </>
  )
}

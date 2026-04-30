'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import Nav from '@/components/Nav'
import { createOrder, getMerchItems, getCart, saveCart, clearCart as clearCartAPI } from '@/lib/api'
import type { MerchItem, OrderItem, CartItem as APICartItem } from '@/types'
import styles from './cart.module.css'

const CART_STORAGE_KEY = 'nitrous_cart_v1'
const CART_UPDATED_EVENT = 'nitrous-cart-updated'

interface CartEntry {
  item: MerchItem
  quantity: number
  size?: string
}

export default function CartPage() {
  const router = useRouter()
  const [cart, setCart] = useState<CartEntry[]>([])
  const [products, setProducts] = useState<MerchItem[]>([])
  const [loading, setLoading] = useState(true)
  const [checkoutLoading, setCheckoutLoading] = useState(false)
  const [checkoutMsg, setCheckoutMsg] = useState('')

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

  const toCartEntries = (items: APICartItem[]): CartEntry[] =>
    items.map((entry) => ({
      item: {
        id: entry.merchId,
        name: entry.name,
        icon: entry.icon,
        price: entry.price,
        category: entry.category as MerchItem['category'],
      },
      quantity: entry.quantity,
      size: entry.size,
    }))

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

    const token = localStorage.getItem('nitrous_token')
    if (token) {
      getCart(token)
        .then((items) => {
          const remoteCart = toCartEntries(items)
          setCart(remoteCart)
        })
        .catch(() => {
          // Keep local cart fallback when backend fetch fails.
        })
    }

    setLoading(false)
  }, [])

  // Load products for reconciliation
  useEffect(() => {
    getMerchItems()
      .then(setProducts)
      .catch(() => {})
  }, [])

  const persistCart = (nextCart: CartEntry[]) => {
    const token = localStorage.getItem('nitrous_token')

    try {
      // Guest carts persist in localStorage; authenticated carts live on the backend only.
      if (!token) {
        localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(nextCart))
      }
      window.dispatchEvent(new Event(CART_UPDATED_EVENT))
    } catch {
      // ignore
    }

    if (!token) return
    saveCart(toAPICartItems(nextCart), token).catch(() => {
      // Keep local cart if backend save fails.
    })
  }

  function updateQuantity(index: number, delta: number) {
    const next = [...cart]
    const newQty = next[index].quantity + delta
    if (newQty <= 0) {
      next.splice(index, 1)
    } else {
      next[index] = { ...next[index], quantity: newQty }
    }
    setCart(next)
    persistCart(next)
  }

  function removeItem(index: number) {
    const next = cart.filter((_, i) => i !== index)
    setCart(next)
    persistCart(next)
  }

  function clearCart() {
    setCart([])
    persistCart([])

    const token = localStorage.getItem('nitrous_token')
    if (token) {
      clearCartAPI(token).catch(() => {
        // Ignore backend clear failure; local cart is still cleared.
      })
    }
  }

  async function handleCheckout() {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      router.push('/login')
      return
    }
    if (cart.length === 0) return

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
      // Store order info for payment and redirect to payment page
      localStorage.setItem('pending_order', JSON.stringify({
        orderId: result.order.id,
        total: result.order.total,
        items: orderItems,
      }))
      router.push('/payment')
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Checkout failed'
      setCheckoutMsg(msg)
    } finally {
      setCheckoutLoading(false)
    }
  }

  const cartTotal = cart.reduce((sum, entry) => sum + entry.item.price * entry.quantity, 0)
  const cartCount = cart.reduce((sum, entry) => sum + entry.quantity, 0)

  if (loading) {
    return (
      <div className={styles.container}>
        <Nav />
        <div className={styles.loading}>
          <div className={styles.spinner} />
          <span>Loading cart...</span>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Nav />
      <main className={styles.main}>
        <header className={styles.header}>
          <h1 className={styles.title}>Shopping Cart</h1>
          <p className={styles.subtitle}>
            {cartCount === 0
              ? 'Your cart is empty'
              : `${cartCount} item${cartCount === 1 ? '' : 's'} in your cart`}
          </p>
        </header>

        {checkoutMsg && (
          <div
            className={styles.message}
            style={{ color: checkoutMsg.startsWith('✓') ? 'var(--cyan)' : 'var(--red)' }}
          >
            {checkoutMsg}
          </div>
        )}

        {cart.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>🛒</div>
            <h3>Your cart is empty</h3>
            <p>Browse our merch store to find exclusive items</p>
            <Link href="/merch" className={styles.ctaButton}>
              Browse Merch
            </Link>
          </div>
        ) : (
          <div className={styles.content}>
            <div className={styles.cartList}>
              {cart.map((entry, index) => (
                <div key={`${entry.item.id}-${index}`} className={styles.cartItem}>
                  <div className={styles.itemVisual}>
                    <span className={styles.itemIcon}>{entry.item.icon}</span>
                  </div>
                  <div className={styles.itemInfo}>
                    <div className={styles.itemName}>{entry.item.name}</div>
                    <div className={styles.itemCategory}>{entry.item.category}</div>
                    {entry.size && (
                      <div className={styles.itemSize}>Size: {entry.size}</div>
                    )}
                  </div>
                  <div className={styles.itemQuantity}>
                    <button
                      className={styles.qtyBtn}
                      onClick={() => updateQuantity(index, -1)}
                      disabled={entry.quantity <= 1}
                    >
                      −
                    </button>
                    <span className={styles.qtyValue}>{entry.quantity}</span>
                    <button
                      className={styles.qtyBtn}
                      onClick={() => updateQuantity(index, 1)}
                    >
                      +
                    </button>
                  </div>
                  <div className={styles.itemPrice}>
                    ${(entry.item.price * entry.quantity).toFixed(2)}
                  </div>
                  <button
                    className={styles.removeBtn}
                    onClick={() => removeItem(index)}
                    aria-label="Remove item"
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>

            <div className={styles.summary}>
              <h3 className={styles.summaryTitle}>Order Summary</h3>
              <div className={styles.summaryRow}>
                <span>Subtotal</span>
                <span>${cartTotal.toFixed(2)}</span>
              </div>
              <div className={styles.summaryRow}>
                <span>Shipping</span>
                <span className={styles.freeShipping}>FREE</span>
              </div>
              <div className={styles.summaryDivider} />
              <div className={styles.summaryRow}>
                <span className={styles.totalLabel}>Total</span>
                <span className={styles.totalValue}>${cartTotal.toFixed(2)}</span>
              </div>
              <button
                className={styles.checkoutBtn}
                onClick={handleCheckout}
                disabled={checkoutLoading || cart.length === 0}
              >
                {checkoutLoading ? 'Processing...' : 'Checkout'}
              </button>
              <button className={styles.clearBtn} onClick={clearCart}>
                Clear Cart
              </button>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
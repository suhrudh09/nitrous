'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import Nav from '@/components/Nav'
import { getOrderById, getMerchItems, saveCart } from '@/lib/api'
import type { CartItem, MerchItem, Order } from '@/types'
import styles from '../orders.module.css'

const CART_STORAGE_KEY = 'nitrous_cart_v1'
const CART_UPDATED_EVENT = 'nitrous-cart-updated'

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const [order, setOrder] = useState<Order | null>(null)
  const [merchItems, setMerchItems] = useState<MerchItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const orderId = useMemo(() => String(params?.id || ''), [params])

  useEffect(() => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      router.push('/login')
      return
    }

    Promise.all([
      getOrderById(orderId, token),
      getMerchItems().catch(() => [] as MerchItem[]),
    ])
      .then(([orderData, merchData]) => {
        setOrder(orderData)
        setMerchItems(merchData)
      })
      .catch(() => setError('Failed to load order'))
      .finally(() => setLoading(false))
  }, [orderId, router])

  const handleRepeatOrder = async () => {
    if (!order) return
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      router.push('/login')
      return
    }

    const merchByID = new Map(merchItems.map((item) => [item.id, item]))
    const nextCart: CartItem[] = order.items
      .filter((item) => item.quantity > 0)
      .map((item) => {
        const merch = merchByID.get(item.merchId)
        return {
          merchId: item.merchId,
          name: item.name || merch?.name || item.merchId,
          icon: merch?.icon || '🛍️',
          price: item.price > 0 ? item.price : merch?.price || 0,
          category: merch?.category || 'collectibles',
          quantity: item.quantity,
          size: item.size,
        }
      })
      .filter((item) => item.price > 0)

    await saveCart(nextCart, token)

    const localCart = nextCart.map((item) => ({
      item: {
        id: item.merchId,
        name: item.name,
        icon: item.icon,
        price: item.price,
        category: item.category,
      },
      quantity: item.quantity,
      size: item.size,
    }))

    localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(localCart))
    window.dispatchEvent(new Event(CART_UPDATED_EVENT))
    router.push('/cart')
  }

  const formatPrice = (price: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price)

  if (loading) {
    return (
      <div className={styles.container}>
        <Nav />
        <div className={styles.loading}>
          <div className={styles.spinner} />
          <span>Loading order...</span>
        </div>
      </div>
    )
  }

  if (error || !order) {
    return (
      <div className={styles.container}>
        <Nav />
        <main className={styles.main}>
          <div className={styles.error}>{error || 'Order not found'}</div>
          <Link href="/orders" className={styles.ctaButton}>Back to Orders</Link>
        </main>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Nav />
      <main className={styles.main}>
        <header className={styles.header}>
          <h1 className={styles.title}>Order #{order.id.slice(0, 8).toUpperCase()}</h1>
          <p className={styles.subtitle}>Status: {order.status.toUpperCase()}</p>
        </header>

        <div className={styles.orderCard}>
          <div className={styles.orderItems}>
            {order.items.map((item, idx) => (
              <div key={idx} className={styles.orderItem}>
                <span className={styles.itemName}>{item.name || item.merchId}</span>
                <span className={styles.itemQty}>x{item.quantity}</span>
                <span className={styles.itemPrice}>{formatPrice(item.price * item.quantity)}</span>
              </div>
            ))}
          </div>

          <div className={styles.orderFooter}>
            <span className={styles.totalLabel}>Total</span>
            <div>
              <span className={styles.totalPrice}>{formatPrice(order.total)}</span>
              <button className={styles.ctaButton} style={{ marginLeft: '12px' }} onClick={handleRepeatOrder}>
                Repeat
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}

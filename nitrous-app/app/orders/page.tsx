'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Nav from '@/components/Nav'
import { getMyOrders, getMyPasses, type UserPass } from '@/lib/api'
import type { Order } from '@/types'
import styles from './orders.module.css'

type Tab = 'passes' | 'merch'

export default function OrdersPage() {
  const router = useRouter()
  const [activeTab, setActiveTab] = useState<Tab>('passes')
  const [passes, setPasses] = useState<UserPass[]>([])
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      router.push('/login')
      return
    }

    Promise.all([
      getMyPasses(token).catch(() => []),
      getMyOrders(token).catch(() => []),
    ])
      .then(([passData, orderData]) => {
        setPasses(passData)
        setOrders(orderData)
      })
      .catch(() => setError('Failed to load orders'))
      .finally(() => setLoading(false))
  }, [router])

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      })
    } catch {
      return dateStr
    }
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'var(--orange)'
      case 'confirmed':
      case 'shipped':
        return 'var(--cyan)'
      case 'cancelled':
        return 'var(--red)'
      default:
        return 'var(--grey)'
    }
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <Nav />
        <div className={styles.loading}>
          <div className={styles.spinner} />
          <span>Loading your orders...</span>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Nav />
      <main className={styles.main}>
        <header className={styles.header}>
          <h1 className={styles.title}>My Orders</h1>
          <p className={styles.subtitle}>
            View your event passes and merchandise orders
          </p>
        </header>

        <div className={styles.tabs}>
          <button
            className={`${styles.tab} ${activeTab === 'passes' ? styles.tabActive : ''}`}
            onClick={() => setActiveTab('passes')}
          >
            Event Passes
            {passes.length > 0 && (
              <span className={styles.badge}>{passes.length}</span>
            )}
          </button>
          <button
            className={`${styles.tab} ${activeTab === 'merch' ? styles.tabActive : ''}`}
            onClick={() => setActiveTab('merch')}
          >
            Merchandise
            {orders.length > 0 && (
              <span className={styles.badge}>{orders.length}</span>
            )}
          </button>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        {activeTab === 'passes' ? (
          <div className={styles.content}>
            {passes.length === 0 ? (
              <div className={styles.empty}>
                <div className={styles.emptyIcon}>🎟️</div>
                <h3>No event passes yet</h3>
                <p>Browse our events and purchase passes to see them here</p>
                <button
                  className={styles.ctaButton}
                  onClick={() => router.push('/passes')}
                >
                  Browse Passes
                </button>
              </div>
            ) : (
              <div className={styles.grid}>
                {passes.map((pass) => (
                  <div key={pass.purchaseId} className={styles.card}>
                    <div
                      className={styles.cardHeader}
                      style={{ borderColor: `var(--${pass.tierColor})` }}
                    >
                      <span
                        className={styles.tier}
                        style={{ color: `var(--${pass.tierColor})` }}
                      >
                        {pass.tier}
                      </span>
                      {pass.badge && (
                        <span className={styles.cardBadge}>{pass.badge}</span>
                      )}
                    </div>
                    <div className={styles.cardBody}>
                      <h3 className={styles.eventName}>{pass.event}</h3>
                      <div className={styles.cardDetails}>
                        <div className={styles.detailRow}>
                          <span className={styles.detailLabel}>📍</span>
                          <span>{pass.location}</span>
                        </div>
                        <div className={styles.detailRow}>
                          <span className={styles.detailLabel}>📅</span>
                          <span>{pass.date}</span>
                        </div>
                        <div className={styles.detailRow}>
                          <span className={styles.detailLabel}>🏷️</span>
                          <span>{pass.category}</span>
                        </div>
                      </div>
                      {pass.perks.length > 0 && (
                        <div className={styles.perks}>
                          <span className={styles.perksLabel}>Includes:</span>
                          <ul className={styles.perksList}>
                            {pass.perks.slice(0, 3).map((perk, i) => (
                              <li key={i}>{perk}</li>
                            ))}
                            {pass.perks.length > 3 && (
                              <li>+{pass.perks.length - 3} more</li>
                            )}
                          </ul>
                        </div>
                      )}
                      <div className={styles.cardFooter}>
                        <span className={styles.price}>
                          {formatPrice(pass.price)}
                        </span>
                        <span className={styles.purchaseDate}>
                          Purchased {formatDate(pass.createdAt)}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className={styles.content}>
            {orders.length === 0 ? (
              <div className={styles.empty}>
                <div className={styles.emptyIcon}>🛍️</div>
                <h3>No merchandise orders yet</h3>
                <p>Check out our merch store to find exclusive items</p>
                <button
                  className={styles.ctaButton}
                  onClick={() => router.push('/merch')}
                >
                  Browse Merch
                </button>
              </div>
            ) : (
              <div className={styles.list}>
                {orders.map((order) => (
                  <div key={order.id} className={styles.orderCard}>
                    <div className={styles.orderHeader}>
                      <div>
                        <span className={styles.orderId}>
                          Order #{order.id.slice(0, 8).toUpperCase()}
                        </span>
                        <span className={styles.orderDate}>
                          {formatDate(order.createdAt)}
                        </span>
                      </div>
                      <span
                        className={styles.orderStatus}
                        style={{ color: getStatusColor(order.status) }}
                      >
                        {order.status.charAt(0).toUpperCase() +
                          order.status.slice(1)}
                      </span>
                    </div>
                    <div className={styles.orderItems}>
                      {order.items.map((item, i) => (
                        <div key={i} className={styles.orderItem}>
                          <span className={styles.itemName}>
                            {item.name || item.merchId}
                          </span>
                          <span className={styles.itemQty}>
                            x{item.quantity}
                          </span>
                          <span className={styles.itemPrice}>
                            {formatPrice(item.price * item.quantity)}
                          </span>
                        </div>
                      ))}
                    </div>
                    <div className={styles.orderFooter}>
                      <span className={styles.totalLabel}>Total</span>
                      <span className={styles.totalPrice}>
                        {formatPrice(order.total)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  )
}
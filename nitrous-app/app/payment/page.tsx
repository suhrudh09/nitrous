'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import Nav from '@/components/Nav'
import { createPaymentIntent, confirmPayment } from '@/lib/api'
import type { OrderItem, PaymentState } from '@/types'
import styles from './payment.module.css'

interface PendingOrder {
  orderId: string
  total: number
  items: OrderItem[]
}

export default function PaymentPage() {
  const router = useRouter()
  const [pendingOrder, setPendingOrder] = useState<PendingOrder | null>(null)
  const [loading, setLoading] = useState(true)
  const [paymentState, setPaymentState] = useState<PaymentState>({ status: 'idle' })
  
  // Card form state
  const [cardNumber, setCardNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvc, setCvc] = useState('')
  const [name, setName] = useState('')

  useEffect(() => {
    // Load pending order from localStorage
    try {
      const stored = localStorage.getItem('pending_order')
      if (stored) {
        const order = JSON.parse(stored) as PendingOrder
        setPendingOrder(order)
      } else {
        // No pending order, redirect to cart
        router.push('/cart')
      }
    } catch {
      router.push('/cart')
    }
    setLoading(false)
  }, [router])

  const formatCardNumber = (value: string) => {
    const digits = value.replace(/\D/g, '').slice(0, 16)
    return digits.replace(/(\d{4})(?=\d)/g, '$1 ')
  }

  const formatExpiry = (value: string) => {
    const digits = value.replace(/\D/g, '').slice(0, 4)
    if (digits.length >= 2) {
      return digits.slice(0, 2) + '/' + digits.slice(2)
    }
    return digits
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    const token = localStorage.getItem('nitrous_token')
    if (!token || !pendingOrder) {
      router.push('/login')
      return
    }

    setPaymentState({ status: 'processing' })

    try {
      // Create payment intent
      const intent = await createPaymentIntent(
        pendingOrder.total,
        'order',
        pendingOrder.orderId,
        token
      )

      // Simulate payment processing (in production, this would use Stripe Elements)
      await new Promise(resolve => setTimeout(resolve, 1500))

      // Confirm payment
      await confirmPayment(intent.paymentId, token)

      // Clear pending order
      localStorage.removeItem('pending_order')

      setPaymentState({
        status: 'success',
        paymentId: intent.paymentId,
      })
    } catch (err) {
      setPaymentState({
        status: 'failed',
        error: err instanceof Error ? err.message : 'Payment failed',
      })
    }
  }

  const handleRetry = () => {
    setPaymentState({ status: 'idle' })
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <Nav />
        <div className={styles.main}>
          <div className={styles.header}>
            <h1 className={styles.title}>Payment</h1>
          </div>
          <div style={{ textAlign: 'center', color: 'var(--muted)' }}>
            Loading...
          </div>
        </div>
      </div>
    )
  }

  // Success state
  if (paymentState.status === 'success') {
    return (
      <div className={styles.container}>
        <Nav />
        <main className={styles.main}>
          <div className={styles.successState}>
            <div className={styles.successIcon}>✓</div>
            <h2 className={styles.successTitle}>Payment Successful!</h2>
            <p className={styles.successMessage}>
              Your order has been confirmed. You will receive a confirmation email shortly.
            </p>
            <Link href="/orders" className={styles.submitBtn}>
              View Orders
            </Link>
          </div>
        </main>
      </div>
    )
  }

  // Failed state
  if (paymentState.status === 'failed') {
    return (
      <div className={styles.container}>
        <Nav />
        <main className={styles.main}>
          <div className={styles.errorState}>
            <div className={styles.errorIcon}>✕</div>
            <h2 className={styles.errorTitle}>Payment Failed</h2>
            <p className={styles.errorMessage}>
              {paymentState.error || 'There was an issue processing your payment.'}
            </p>
            <button onClick={handleRetry} className={styles.retryBtn}>
              Try Again
            </button>
          </div>
        </main>
      </div>
    )
  }

  // No pending order
  if (!pendingOrder) {
    return (
      <div className={styles.container}>
        <Nav />
        <main className={styles.main}>
          <div style={{ textAlign: 'center', color: 'var(--muted)' }}>
            No pending order found.
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Nav />
      <main className={styles.main}>
        <Link href="/cart" className={styles.backLink}>
          ← Back to Cart
        </Link>

        <header className={styles.header}>
          <h1 className={styles.title}>Payment</h1>
          <p className={styles.subtitle}>Complete your purchase</p>
        </header>

        {/* Order Summary */}
        <div className={styles.orderSummary}>
          <h2 className={styles.orderTitle}>Order Summary</h2>
          <div className={styles.orderItems}>
            {pendingOrder.items.map((item, index) => (
              <div key={index} className={styles.orderItem}>
                <div>
                  <div className={styles.itemName}>{item.name}</div>
                  <div className={styles.itemQty}>Qty: {item.quantity}</div>
                </div>
                <div className={styles.itemPrice}>
                  ${(item.price * item.quantity).toFixed(2)}
                </div>
              </div>
            ))}
          </div>
          <div className={styles.orderTotal}>
            <span className={styles.totalLabel}>Total</span>
            <span className={styles.totalAmount}>${pendingOrder.total.toFixed(2)}</span>
          </div>
        </div>

        {/* Card Visual */}
        <div className={styles.cardVisual}>
          <div className={styles.cardNumber}>
            {cardNumber || '•••• •••• •••• ••••'}
          </div>
          <div className={styles.cardDetails}>
            <div>
              <div className={styles.cardLabel}>Card Holder</div>
              <div className={styles.cardValue}>{name || 'YOUR NAME'}</div>
            </div>
            <div>
              <div className={styles.cardLabel}>Expires</div>
              <div className={styles.cardValue}>{expiry || 'MM/YY'}</div>
            </div>
          </div>
        </div>

        {/* Payment Form */}
        <form onSubmit={handleSubmit} className={styles.paymentForm}>
          <h2 className={styles.formTitle}>Card Details</h2>
          
          <div className={styles.formGroup}>
            <label className={styles.label}>Card Number</label>
            <input
              type="text"
              className={styles.input}
              placeholder="1234 5678 9012 3456"
              value={cardNumber}
              onChange={(e) => setCardNumber(formatCardNumber(e.target.value))}
              required
            />
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>Card Holder Name</label>
            <input
              type="text"
              className={styles.input}
              placeholder="John Doe"
              value={name}
              onChange={(e) => setName(e.target.value.toUpperCase())}
              required
            />
          </div>

          <div className={styles.row}>
            <div className={styles.formGroup}>
              <label className={styles.label}>Expiry Date</label>
              <input
                type="text"
                className={styles.input}
                placeholder="MM/YY"
                value={expiry}
                onChange={(e) => setExpiry(formatExpiry(e.target.value))}
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label className={styles.label}>CVC</label>
              <input
                type="text"
                className={styles.input}
                placeholder="123"
                value={cvc}
                onChange={(e) => setCvc(e.target.value.replace(/\D/g, '').slice(0, 4))}
                required
              />
            </div>
          </div>

          <button
            type="submit"
            className={`${styles.submitBtn} ${paymentState.status === 'processing' ? styles.processing : ''}`}
            disabled={paymentState.status === 'processing'}
          >
            {paymentState.status === 'processing' ? 'Processing...' : `Pay $${pendingOrder.total.toFixed(2)}`}
          </button>
        </form>
      </main>
    </div>
  )
}
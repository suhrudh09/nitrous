'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getJourneys, bookJourney, createPaymentIntent, confirmPayment } from '@/lib/api'
import { useCanBookJourneys, useCanRegisterOthers, useUser } from '@/hooks/usePermission'
import type { Journey } from '@/types'
import styles from './journeys.module.css'

const badgeColors: Record<string, string> = {
  EXCLUSIVE: 'var(--cyan)',
  'MEMBERS ONLY': '#facc15',
  LIMITED: 'var(--red)',
}

const colorMap: Record<string, string> = {
  red: 'var(--red)',
  cyan: 'var(--cyan)',
  orange: '#fb923c',
  blue: '#60a5fa',
  purple: '#a78bfa',
  gold: '#facc15',
}

interface PaymentModal {
  journeyId: string
  journeyTitle: string
  pricePerPerson: number
  quantity: number
  targetUserId?: string
}

function formatCardNumber(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 16)
  return digits.replace(/(\d{4})(?=\d)/g, '$1 ')
}

function formatExpiry(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 4)
  if (digits.length >= 2) return digits.slice(0, 2) + '/' + digits.slice(2)
  return digits
}

export default function JourneysPage() {
  const [journeys, setJourneys] = useState<Journey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [booked, setBooked] = useState<Set<string>>(new Set())
  const [registerForUser, setRegisterForUser] = useState<string>('')
  const [quantities, setQuantities] = useState<Record<string, number>>({})
  const [paymentModal, setPaymentModal] = useState<PaymentModal | null>(null)
  const [payStatus, setPayStatus] = useState<'idle' | 'processing' | 'done' | 'error'>('idle')
  const [payError, setPayError] = useState('')
  const [cardNumber, setCardNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvc, setCvc] = useState('')
  const [cardName, setCardName] = useState('')

  // Permission hooks
  const canBook = useCanBookJourneys()
  const canRegisterOthers = useCanRegisterOthers()
  const { user } = useUser()

  const getQty = (id: string) => quantities[id] ?? 1
  const setQty = (id: string, v: number) =>
    setQuantities(prev => ({ ...prev, [id]: Math.max(1, Math.min(10, v)) }))

  useEffect(() => {
    getJourneys()
      .then(setJourneys)
      .catch(() => setError('Could not load journeys'))
      .finally(() => setLoading(false))
  }, [])

  function openPayment(journey: Journey) {
    const token = localStorage.getItem('nitrous_token')
    if (!token) { globalThis.location.href = '/login'; return }
    if (!canBook && user?.role === 'participant') {
      alert('Participants cannot register for journeys. Contact your team manager to register.')
      return
    }
    const targetUserId = canRegisterOthers && registerForUser ? registerForUser : undefined
    setPaymentModal({ journeyId: journey.id, journeyTitle: journey.title, pricePerPerson: journey.price, quantity: getQty(journey.id), targetUserId })
    setPayStatus('idle')
    setPayError('')
    setCardNumber('')
    setExpiry('')
    setCvc('')
    setCardName('')
  }

  async function handlePaySubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!paymentModal) return
    const token = localStorage.getItem('nitrous_token')
    if (!token) { globalThis.location.href = '/login'; return }

    setPayStatus('processing')
    setPayError('')
    try {
      const total = Math.round(paymentModal.pricePerPerson * paymentModal.quantity * 100)
      const intent = await createPaymentIntent(total, 'journey', paymentModal.journeyId, token)
      await confirmPayment(intent.paymentId, token)
      const result = await bookJourney(paymentModal.journeyId, token, paymentModal.targetUserId, paymentModal.quantity)
      setJourneys(prev => prev.map(j => j.id === result.journey.id ? result.journey : j))
      setBooked(prev => new Set([...prev, paymentModal.journeyId]))
      setPayStatus('done')
      setTimeout(() => setPaymentModal(null), 1800)
    } catch (err: unknown) {
      setPayStatus('error')
      setPayError(err instanceof Error ? err.message : 'Payment failed')
    }
  }

  if (loading) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div style={{ padding: '120px 48px', color: 'var(--muted)', fontFamily: 'var(--font-mono)' }}>
            LOADING JOURNEYS...
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
        {/* Hero Header */}
        <div className={styles.pageHero}>
          <div className={styles.heroLines}>
            {[...new Array(5)].map((_, i) => (
              <div key={`line-${i}`} className={styles.heroLine}></div>
            ))}
          </div>
          <div className={styles.heroContent}>
            <div className={styles.headerTag}>/ JOURNEYS</div>
            <h1 className={styles.pageTitle}>
              LIVE THE
              <br />
              <span className={styles.titleAccent}>EXPERIENCE</span>
            </h1>
            <p className={styles.pageSubtitle}>
              Exclusive access journeys for those who want more than just a stream. Be there.
              Feel it. Live it.
            </p>
          </div>
          <div className={styles.heroStats}>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>{journeys.length}</div>
              <div className={styles.hStatL}>JOURNEYS</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>18</div>
              <div className={styles.hStatL}>COUNTRIES</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>{journeys.filter((j) => j.slotsLeft > 0).length}</div>
              <div className={styles.hStatL}>AVAILABLE</div>
            </div>
          </div>
        </div>

        {/* Journeys List */}
        <div className={styles.journeysList}>
          {journeys.map((journey, i) => {
            // Derive a color key from category keywords for consistent accent coloring
            const colorKey = journey.category.toLowerCase().includes('air')
              ? 'cyan'
              : journey.category.toLowerCase().includes('rally') ||
                journey.category.toLowerCase().includes('offroad') ||
                journey.category.toLowerCase().includes('desert')
              ? 'orange'
              : journey.category.toLowerCase().includes('water')
              ? 'blue'
              : 'red'
            const accent = colorMap[colorKey] ?? 'var(--cyan)'
            const isUrgent = journey.slotsLeft <= 4

            return (
              <div key={journey.id} className={styles.journeyCard}>
                {/* Left: number */}
                <div className={styles.journeyNum} style={{ color: accent }}>
                  {String(i + 1).padStart(2, '0')}
                </div>

                {/* Main content */}
                <div className={styles.journeyMain}>
                  <div className={styles.journeyHead}>
                    <div>
                      <div className={styles.journeyCat}>{journey.category}</div>
                      <div className={styles.journeyTitle}>{journey.title}</div>
                    </div>
                    <div
                      className={styles.journeyBadge}
                      style={{
                        color: badgeColors[journey.badge],
                        borderColor: badgeColors[journey.badge],
                      }}
                    >
                      {journey.badge}
                    </div>
                  </div>

                  <p className={styles.journeyDesc}>{journey.description}</p>

                  <div className={styles.journeyMeta}>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>📅</span>
                      <span>{new Date(journey.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
                    </div>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>💰</span>
                      <span>${journey.price.toLocaleString()} per person</span>
                    </div>
                  </div>
                </div>

                {/* Right: price + slots + CTA */}
                <div className={styles.journeyAside}>
                  <div className={styles.journeyPrice}>
                    <span className={styles.priceFrom}>FROM</span>
                    <span className={styles.priceVal} style={{ color: accent }}>
                      ${journey.price.toLocaleString()}
                    </span>
                    <span className={styles.priceLabel}>per person</span>
                  </div>

                  <div className={styles.slotsWrap}>
                    <div className={styles.slotsTop}>
                      <span className={`${styles.slotsLabel} ${isUrgent ? styles.slotsUrgent : ''}`}>
                        {isUrgent ? '🔥 ' : ''}
                        {journey.slotsLeft} SLOTS LEFT
                      </span>
                    </div>
                    <div className={styles.slotsBar}>
                      <div
                        className={styles.slotsFill}
                        style={{
                          width: `${Math.max((journey.slotsLeft / 20) * 100, 5)}%`,
                          background: isUrgent ? 'var(--red)' : accent,
                        }}
                      ></div>
                    </div>
                  </div>

                  {!booked.has(journey.id) && journey.slotsLeft > 0 && canBook && (
                    <div className={styles.peopleSelector}>
                      <button className={styles.stepBtn} onClick={() => setQty(journey.id, getQty(journey.id) - 1)}>−</button>
                      <span className={styles.stepCount}>{getQty(journey.id)}</span>
                      <button className={styles.stepBtn} onClick={() => setQty(journey.id, getQty(journey.id) + 1)}>+</button>
                      <span className={styles.stepLabel}>{getQty(journey.id) === 1 ? 'person' : 'people'}</span>
                    </div>
                  )}

                  <button
                    className={styles.bookBtn}
                    style={{
                      background: booked.has(journey.id)
                        ? 'rgba(74,222,128,0.1)'
                        : `linear-gradient(135deg, ${accent}22, ${accent}11)`,
                      borderColor: booked.has(journey.id) ? '#4ade80' : `${accent}66`,
                      color: booked.has(journey.id) ? '#4ade80' : accent,
                      opacity: journey.slotsLeft <= 0 ? 0.4 : 1,
                      cursor: journey.slotsLeft <= 0 ? 'not-allowed' : 'pointer',
                    }}
                    onClick={() => !booked.has(journey.id) && openPayment(journey)}
                    disabled={journey.slotsLeft <= 0 || (!canBook && user?.role === 'participant') || booked.has(journey.id)}
                  >
                    {booked.has(journey.id)
                      ? '✓ BOOKED'
                      : journey.slotsLeft <= 0
                      ? 'SOLD OUT'
                      : !canBook && user?.role === 'participant'
                      ? 'CONTACT MANAGER'
                      : `BOOK · $${(journey.price * getQty(journey.id)).toLocaleString()} →`}
                  </button>
                </div>

                {/* Color accent left border */}
                <div
                  className={styles.journeyBorder}
                  style={{ background: accent, boxShadow: `0 0 12px ${accent}` }}
                ></div>
              </div>
            )
          })}
        </div>
      </main>

      {/* Payment Modal */}
      {paymentModal && (
        <div className={styles.payModalOverlay} onClick={() => payStatus !== 'processing' && setPaymentModal(null)}>
          <div className={styles.payModal} onClick={e => e.stopPropagation()}>
            {payStatus === 'done' ? (
              <div className={styles.payDone}>
                <div className={styles.payDoneIcon}>✓</div>
                <div className={styles.payDoneTitle}>Journey Booked!</div>
                <div className={styles.payDoneSub}>{paymentModal.journeyTitle} × {paymentModal.quantity}</div>
              </div>
            ) : (
              <>
                <div className={styles.payModalHeader}>
                  <div className={styles.payModalTitle}>BOOK JOURNEY</div>
                  <button className={styles.payModalClose} onClick={() => setPaymentModal(null)} disabled={payStatus === 'processing'}>✕</button>
                </div>
                <div className={styles.payOrderSummary}>
                  <div className={styles.payOrderRow}>
                    <span>{paymentModal.journeyTitle}</span>
                    <span>${paymentModal.pricePerPerson.toLocaleString()} × {paymentModal.quantity}</span>
                  </div>
                  <div className={styles.payOrderTotal}>
                    <span>TOTAL</span>
                    <span>${(paymentModal.pricePerPerson * paymentModal.quantity).toLocaleString()}</span>
                  </div>
                </div>
                <form className={styles.payForm} onSubmit={handlePaySubmit}>
                  <div className={styles.payFormField}>
                    <label className={styles.payLabel}>CARDHOLDER NAME</label>
                    <input className={styles.payInput} placeholder="Full Name" value={cardName} onChange={e => setCardName(e.target.value)} required />
                  </div>
                  <div className={styles.payFormField}>
                    <label className={styles.payLabel}>CARD NUMBER</label>
                    <input className={styles.payInput} placeholder="1234 5678 9012 3456" value={cardNumber} onChange={e => setCardNumber(formatCardNumber(e.target.value))} maxLength={19} required />
                  </div>
                  <div className={styles.payFormRow}>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>EXPIRY</label>
                      <input className={styles.payInput} placeholder="MM/YY" value={expiry} onChange={e => setExpiry(formatExpiry(e.target.value))} maxLength={5} required />
                    </div>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>CVC</label>
                      <input className={styles.payInput} placeholder="123" value={cvc} onChange={e => setCvc(e.target.value.replace(/\D/g, '').slice(0, 4))} maxLength={4} required />
                    </div>
                  </div>
                  {payError && <div className={styles.payError}>{payError}</div>}
                  <button className={styles.paySubmitBtn} type="submit" disabled={payStatus === 'processing'}>
                    {payStatus === 'processing' ? 'PROCESSING...' : `PAY $${(paymentModal.pricePerPerson * paymentModal.quantity).toLocaleString()}`}
                  </button>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </>
  )
}
'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import styles from './passes.module.css'
import { purchasePass, createPaymentIntent, confirmPayment, getAvailablePasses, type AvailablePass } from '@/lib/api'

const colorMap: Record<string, string> = {
  gold: '#facc15',
  GOLD: '#facc15',
  cyan: 'var(--cyan)',
  CYAN: 'var(--cyan)',
  red: 'var(--red)',
  RED: 'var(--red)',
  muted: 'var(--muted)',
  orange: '#fb923c',
  blue: '#60a5fa',
  BLUE: '#60a5fa',
  purple: '#a78bfa',
}

const catIcons: Record<string, string> = {
  MOTORSPORT: '🏎️',
  'OFF-ROAD': '🏔️',
  WATER: '🌊',
  AIR: '🪂',
}

type Tier = 'all' | 'PLATINUM' | 'VIP' | 'GENERAL'
type Category = 'all' | 'MOTORSPORT' | 'OFF-ROAD' | 'WATER' | 'AIR'

interface PaymentModal {
  passId: string
  passName: string
  pricePerPerson: number
  quantity: number
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

export default function PassesPage() {
  const [tierFilter, setTierFilter] = useState<Tier>('all')
  const [catFilter, setCatFilter] = useState<Category>('all')
  const [passes, setPasses] = useState<AvailablePass[]>([])
  const [loadingPasses, setLoadingPasses] = useState(true)
  const [quantities, setQuantities] = useState<Record<string, number>>({})
  const [purchased, setPurchased] = useState<Set<string>>(new Set())
  const [paymentModal, setPaymentModal] = useState<PaymentModal | null>(null)
  const [payStatus, setPayStatus] = useState<'idle' | 'processing' | 'done' | 'error'>('idle')
  const [payError, setPayError] = useState('')
  const [cardNumber, setCardNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvc, setCvc] = useState('')
  const [cardName, setCardName] = useState('')

  const getQty = (id: string) => quantities[id] ?? 1
  const setQty = (id: string, v: number) =>
    setQuantities(prev => ({ ...prev, [id]: Math.max(1, Math.min(10, v)) }))

  useEffect(() => {
    getAvailablePasses()
      .then(setPasses)
      .catch(() => {/* silently show empty state */})
      .finally(() => setLoadingPasses(false))
  }, [])

  const filtered = passes.filter(p => {
    if (tierFilter !== 'all' && p.tier !== tierFilter) return false
    if (catFilter !== 'all' && p.category !== catFilter) return false
    return true
  })

  function openPayment(passId: string, passName: string, pricePerPerson: number) {
    const token = localStorage.getItem('nitrous_token')
    if (!token) { globalThis.location.href = '/login'; return }
    setPaymentModal({ passId, passName, pricePerPerson, quantity: getQty(passId) })
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
      const intent = await createPaymentIntent(total, 'pass', paymentModal.passId, token)
      await confirmPayment(intent.paymentId, token)
      await purchasePass(paymentModal.passId, token, paymentModal.quantity)
      setPasses(prev =>
        prev.map(pass =>
          pass.id === paymentModal.passId
            ? { ...pass, spotsLeft: Math.max(0, pass.spotsLeft - paymentModal.quantity) }
            : pass
        )
      )
      setQuantities(prev => ({ ...prev, [paymentModal.passId]: 1 }))
      setPurchased(prev => new Set([...prev, paymentModal.passId]))
      setPayStatus('done')
      setTimeout(() => setPaymentModal(null), 1800)
    } catch (err: unknown) {
      setPayStatus('error')
      setPayError(err instanceof Error ? err.message : 'Payment failed')
    }
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>

        {/* Hero Header */}
        <div className={styles.pageHero}>
          <div className={styles.heroLines}>
            {[...Array(5)].map((_, i) => <div key={i} className={styles.heroLine} />)}
          </div>
          <div className={styles.heroContent}>
            <div className={styles.headerTag}>/ EVENT PASSES</div>
            <h1 className={styles.pageTitle}>
              ACCESS<br />
              <span className={styles.titleAccent}>PASSES</span>
            </h1>
            <p className={styles.pageSubtitle}>
              Exclusive credentials for those who need more than a ticket. Get inside the action.
            </p>
          </div>
          <div className={styles.heroStats}>
            <div className={styles.hStat}><div className={styles.hStatN}>{passes.length}</div><div className={styles.hStatL}>PASSES</div></div>
            <div className={styles.hStat}><div className={styles.hStatN} style={{ color: '#facc15' }}>{passes.filter(p => p.tier.toUpperCase().includes('PLATINUM')).length}</div><div className={styles.hStatL}>PLATINUM</div></div>
            <div className={styles.hStat}><div className={styles.hStatN} style={{ color: 'var(--red)' }}>{passes.filter(p => p.spotsLeft <= 5).length}</div><div className={styles.hStatL}>URGENT</div></div>
          </div>
        </div>

        {/* Filters */}
        <div className={styles.filterBar}>
          <div className={styles.filterGroup}>
            <span className={styles.filterGroupLabel}>TIER</span>
            {(['all', 'PLATINUM', 'VIP', 'GENERAL'] as Tier[]).map(t => (
              <button
                key={t}
                className={`${styles.filterBtn} ${tierFilter === t ? styles.filterBtnActive : ''}`}
                onClick={() => setTierFilter(t)}
              >
                {t === 'all' ? 'ALL' : t}
              </button>
            ))}
          </div>
          <div className={styles.filterGroup}>
            <span className={styles.filterGroupLabel}>CATEGORY</span>
            {(['all', 'MOTORSPORT', 'OFF-ROAD', 'WATER', 'AIR'] as Category[]).map(c => (
              <button
                key={c}
                className={`${styles.filterBtn} ${catFilter === c ? styles.filterBtnActive : ''}`}
                onClick={() => setCatFilter(c)}
              >
                {c === 'all' ? 'ALL' : `${catIcons[c]} ${c}`}
              </button>
            ))}
          </div>
          <span className={styles.countTxt}>{filtered.length} pass{filtered.length !== 1 ? 'es' : ''}</span>
        </div>

        {/* Pass Cards */}
        <div className={styles.passGrid}>
          {loadingPasses ? (
            <div style={{ gridColumn: '1/-1', padding: '60px', color: 'var(--muted)', fontFamily: 'var(--font-mono)', textAlign: 'center' }}>
              LOADING PASSES...
            </div>
          ) : filtered.length === 0 ? (
            <div style={{ gridColumn: '1/-1', padding: '60px', color: 'var(--muted)', fontFamily: 'var(--font-mono)', textAlign: 'center' }}>
              NO PASSES MATCH YOUR FILTERS
            </div>
          ) : filtered.map(pass => {
            const accent = colorMap[pass.tierColor] ?? 'var(--cyan)'
            const isUrgent = pass.spotsLeft <= 5
            const isPurchased = purchased.has(pass.id)
            const slotPct = Math.max((pass.spotsLeft / pass.totalSpots) * 100, 4)

            return (
              <div key={pass.id} className={styles.passCard}>
                {/* Top accent bar */}
                <div className={styles.cardAccentBar} style={{ background: accent, boxShadow: `0 0 10px ${accent}80` }} />

                {/* Badge */}
                {pass.badge && (
                  <div className={styles.urgencyBadge} style={{
                    background: isUrgent ? 'rgba(255,42,42,0.12)' : 'rgba(250,204,21,0.1)',
                    borderColor: isUrgent ? 'rgba(255,42,42,0.4)' : 'rgba(250,204,21,0.4)',
                    color: isUrgent ? 'var(--red)' : '#facc15',
                  }}>
                    {isUrgent && <span className={styles.urgencyDot} />}
                    {pass.badge}
                  </div>
                )}

                {/* Tier + Category */}
                <div className={styles.cardTopRow}>
                  <div className={styles.tierBadge} style={{ color: accent, borderColor: `${accent}66` }}>
                    {pass.tier}
                  </div>
                  <div className={styles.catChip}>
                    {catIcons[pass.category.toUpperCase()] ?? '🏁'} {pass.category.toUpperCase()}
                  </div>
                </div>

                {/* Event info */}
                <div className={styles.cardEvent}>{pass.event}</div>
                <div className={styles.cardLocation}>📍 {pass.location}</div>
                <div className={styles.cardDate}>📅 {(() => { try { return new Date(pass.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) } catch { return pass.date } })()}</div>

                {/* Perks */}
                <div className={styles.perksSection}>
                  <div className={styles.perksLabel}>INCLUDED</div>
                  <div className={styles.perksList}>
                    {pass.perks.map((perk, i) => (
                      <div key={i} className={styles.perkItem}>
                        <span className={styles.perkDot} style={{ background: accent }} />
                        <span>{perk}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Availability bar */}
                <div className={styles.availSection}>
                  <div className={styles.availTop}>
                    <span className={`${styles.availLabel} ${isUrgent ? styles.availLabelUrgent : ''}`}>
                      {pass.spotsLeft} / {pass.totalSpots} spots left
                    </span>
                  </div>
                  <div className={styles.availBar}>
                    <div className={styles.availFill} style={{ width: `${slotPct}%`, background: isUrgent ? 'var(--red)' : accent }} />
                  </div>
                </div>

                {/* Price + people selector + CTA */}
                <div className={styles.cardBottom}>
                  <div className={styles.priceBlock}>
                    <span className={styles.priceFrom}>FROM</span>
                    <span className={styles.priceVal} style={{ color: accent }}>${pass.price.toLocaleString()}</span>
                  </div>

                  {!isPurchased && (
                    <div className={styles.peopleSelector}>
                      <button className={styles.stepBtn} onClick={() => setQty(pass.id, getQty(pass.id) - 1)}>−</button>
                      <span className={styles.stepCount}>{getQty(pass.id)}</span>
                      <button className={styles.stepBtn} onClick={() => setQty(pass.id, getQty(pass.id) + 1)}>+</button>
                      <span className={styles.stepLabel}>
                        {getQty(pass.id) === 1 ? 'person' : 'people'}
                      </span>
                    </div>
                  )}

                  <button
                    className={`${styles.purchaseBtn} ${isPurchased ? styles.purchaseBtnDone : ''}`}
                    style={!isPurchased ? { borderColor: accent, color: accent, background: `${accent}0d` } : {}}
                    onClick={() => !isPurchased && openPayment(pass.id, pass.event, pass.price)}
                    disabled={isPurchased}
                  >
                    {isPurchased ? '✓ PASS SECURED' : `SECURE PASS · $${(pass.price * getQty(pass.id)).toLocaleString()} →`}
                  </button>
                </div>
              </div>
            )
          })}
        </div>

        {/* Bottom CTA */}
        <div className={styles.bottomBanner}>
          <div className={styles.bannerContent}>
            <div className={styles.bannerTitle}>NEED A CUSTOM PACKAGE?</div>
            <div className={styles.bannerSub}>Corporate groups, team access, and bespoke event credentials available.</div>
          </div>
          <button className={styles.bannerBtn}>Contact Our Team →</button>
        </div>
      </main>

      {/* Payment Modal */}
      {paymentModal && (
        <div className={styles.payModalOverlay} onClick={() => payStatus !== 'processing' && setPaymentModal(null)}>
          <div className={styles.payModal} onClick={e => e.stopPropagation()}>
            {payStatus === 'done' ? (
              <div className={styles.payDone}>
                <div className={styles.payDoneIcon}>✓</div>
                <div className={styles.payDoneTitle}>Pass Secured!</div>
                <div className={styles.payDoneSub}>{paymentModal.passName} × {paymentModal.quantity}</div>
              </div>
            ) : (
              <>
                <div className={styles.payModalHeader}>
                  <div className={styles.payModalTitle}>SECURE YOUR PASS</div>
                  <button className={styles.payModalClose} onClick={() => setPaymentModal(null)} disabled={payStatus === 'processing'}>✕</button>
                </div>
                <div className={styles.payOrderSummary}>
                  <div className={styles.payOrderRow}>
                    <span>{paymentModal.passName}</span>
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
                    <input
                      className={styles.payInput}
                      placeholder="Full Name"
                      value={cardName}
                      onChange={e => setCardName(e.target.value)}
                      required
                    />
                  </div>
                  <div className={styles.payFormField}>
                    <label className={styles.payLabel}>CARD NUMBER</label>
                    <input
                      className={styles.payInput}
                      placeholder="1234 5678 9012 3456"
                      value={cardNumber}
                      onChange={e => setCardNumber(formatCardNumber(e.target.value))}
                      maxLength={19}
                      required
                    />
                  </div>
                  <div className={styles.payFormRow}>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>EXPIRY</label>
                      <input
                        className={styles.payInput}
                        placeholder="MM/YY"
                        value={expiry}
                        onChange={e => setExpiry(formatExpiry(e.target.value))}
                        maxLength={5}
                        required
                      />
                    </div>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>CVC</label>
                      <input
                        className={styles.payInput}
                        placeholder="123"
                        value={cvc}
                        onChange={e => setCvc(e.target.value.replace(/\D/g, '').slice(0, 4))}
                        maxLength={4}
                        required
                      />
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
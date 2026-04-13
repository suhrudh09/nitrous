'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import styles from './passes.module.css'

// ── Static pass data (would come from API in production) ──────────────────────
const PASSES = [
  {
    id: 'p1',
    tier: 'PLATINUM',
    tierColor: 'gold',
    event: 'NASCAR Daytona 500',
    location: 'Daytona International Speedway · Florida',
    date: 'Feb 16, 2026',
    category: 'MOTORSPORT',
    price: 1200,
    perks: ['Pit Lane Access', 'Meet & Greet Drivers', 'Hospitality Suite', 'Paddock Tour', 'Official Credential'],
    spotsLeft: 4,
    totalSpots: 20,
    badge: 'SOLD OUT RISK',
  },
  {
    id: 'p2',
    tier: 'VIP',
    tierColor: 'cyan',
    event: 'Dakar Rally — Stage 9',
    location: 'Al Ula → Hail · Saudi Arabia',
    date: 'Feb 10, 2026',
    category: 'OFF-ROAD',
    price: 850,
    perks: ['Service Park Access', 'Co-Driver Briefing', 'Bivouac Dinner', 'Official Rally Kit'],
    spotsLeft: 12,
    totalSpots: 50,
    badge: null,
  },
  {
    id: 'p3',
    tier: 'GENERAL',
    tierColor: 'muted',
    event: 'Speed Boat Cup — Finals',
    location: 'Lake Como · Italy',
    date: 'Mar 2, 2026',
    category: 'WATER',
    price: 240,
    perks: ['Grandstand Seating', 'Race Programme', 'Boat Dock Access'],
    spotsLeft: 89,
    totalSpots: 200,
    badge: null,
  },
  {
    id: 'p4',
    tier: 'PLATINUM',
    tierColor: 'gold',
    event: 'Red Bull Skydive Series — Rd. 3',
    location: 'Interlaken Drop Zone · Switzerland',
    date: 'Mar 8, 2026',
    category: 'AIR',
    price: 680,
    perks: ['Dropzone Access', 'Pilot Q&A', 'Skydiving Demo Flight', 'Red Bull Lounge', 'Souvenir Pack'],
    spotsLeft: 8,
    totalSpots: 30,
    badge: 'HOT',
  },
  {
    id: 'p5',
    tier: 'VIP',
    tierColor: 'cyan',
    event: 'Crop Duster Air Racing',
    location: 'Bakersfield Airfield · California',
    date: 'Mar 14, 2026',
    category: 'AIR',
    price: 420,
    perks: ['Airside Access', 'Cockpit Photo Session', 'Pilots Lounge Entry', 'VIP Viewing Deck'],
    spotsLeft: 22,
    totalSpots: 60,
    badge: null,
  },
  {
    id: 'p6',
    tier: 'PLATINUM',
    tierColor: 'red',
    event: 'World Dirt Track Championship',
    location: 'Knob Noster · Missouri, USA',
    date: 'Feb 21, 2026',
    category: 'MOTORSPORT',
    price: 960,
    perks: ['Pre-Race Grid Walk', 'Team Garage Access', 'Podium Ceremony Attendance', 'Signed Merchandise', 'Premium Seating'],
    spotsLeft: 2,
    totalSpots: 15,
    badge: 'ALMOST GONE',
  },
]

const colorMap: Record<string, string> = {
  gold: '#facc15',
  cyan: 'var(--cyan)',
  red: 'var(--red)',
  muted: 'var(--muted)',
  orange: '#fb923c',
}

const catIcons: Record<string, string> = {
  MOTORSPORT: '🏎️',
  'OFF-ROAD': '🏔️',
  WATER: '🌊',
  AIR: '🪂',
}

type Tier = 'all' | 'PLATINUM' | 'VIP' | 'GENERAL'
type Category = 'all' | 'MOTORSPORT' | 'OFF-ROAD' | 'WATER' | 'AIR'

export default function PassesPage() {
  const [tierFilter, setTierFilter] = useState<Tier>('all')
  const [catFilter, setCatFilter] = useState<Category>('all')
  const [purchasing, setPurchasing] = useState<string | null>(null)
  const [purchased, setPurchased] = useState<Set<string>>(new Set())

  const filtered = PASSES.filter(p => {
    if (tierFilter !== 'all' && p.tier !== tierFilter) return false
    if (catFilter !== 'all' && p.category !== catFilter) return false
    return true
  })

  async function handlePurchase(passId: string) {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      globalThis.location.href = '/login'
      return
    }
    setPurchasing(passId)
    await new Promise(r => setTimeout(r, 1200))
    setPurchased(prev => new Set([...prev, passId]))
    setPurchasing(null)
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
            <div className={styles.hStat}><div className={styles.hStatN}>{PASSES.length}</div><div className={styles.hStatL}>PASSES</div></div>
            <div className={styles.hStat}><div className={styles.hStatN} style={{ color: '#facc15' }}>{PASSES.filter(p => p.tier === 'PLATINUM').length}</div><div className={styles.hStatL}>PLATINUM</div></div>
            <div className={styles.hStat}><div className={styles.hStatN} style={{ color: 'var(--red)' }}>{PASSES.filter(p => p.spotsLeft <= 5).length}</div><div className={styles.hStatL}>URGENT</div></div>
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
          {filtered.map(pass => {
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
                    {catIcons[pass.category]} {pass.category}
                  </div>
                </div>

                {/* Event info */}
                <div className={styles.cardEvent}>{pass.event}</div>
                <div className={styles.cardLocation}>📍 {pass.location}</div>
                <div className={styles.cardDate}>📅 {pass.date}</div>

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

                {/* Price + CTA */}
                <div className={styles.cardBottom}>
                  <div className={styles.priceBlock}>
                    <span className={styles.priceFrom}>FROM</span>
                    <span className={styles.priceVal} style={{ color: accent }}>${pass.price.toLocaleString()}</span>
                  </div>
                  <button
                    className={`${styles.purchaseBtn} ${isPurchased ? styles.purchaseBtnDone : ''}`}
                    style={!isPurchased ? { borderColor: accent, color: accent, background: `${accent}0d` } : {}}
                    onClick={() => handlePurchase(pass.id)}
                    disabled={purchasing === pass.id || isPurchased}
                  >
                    {isPurchased ? '✓ PASS SECURED' : purchasing === pass.id ? 'PROCESSING...' : 'SECURE PASS →'}
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
    </>
  )
}
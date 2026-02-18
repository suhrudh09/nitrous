import Nav from '@/components/Nav'
import styles from './journeys.module.css'

const journeys = [
  {
    id: '1',
    title: 'DAYTONA PIT CREW EXPERIENCE',
    category: 'MOTORSPORT · BEHIND THE SCENES',
    description: 'Go behind the wall at Daytona 500. Watch pit stops up close, meet the crew chiefs, and ride the pace car on track.',
    badge: 'EXCLUSIVE',
    slotsLeft: 12,
    slotsTotal: 20,
    date: 'Feb 16, 2026',
    duration: '3 Days',
    location: 'Daytona Beach, Florida',
    price: 2400,
    color: 'red',
    includes: ['Pit wall access', 'Team debrief', 'Pace car ride', 'Hospitality suite'],
  },
  {
    id: '2',
    title: 'DAKAR DESERT CONVOY',
    category: 'RALLY · DESERT EXPEDITION',
    description: 'Ride a support vehicle through the Dakar stages. Sleep under the stars, eat with the team, and feel the dust.',
    badge: 'MEMBERS ONLY',
    slotsLeft: 6,
    slotsTotal: 10,
    date: 'Jan 18, 2027',
    duration: '7 Days',
    location: 'Saudi Arabia',
    price: 5800,
    color: 'orange',
    includes: ['Support convoy access', 'Desert camp stays', 'Team meals', 'Recovery vehicle'],
  },
  {
    id: '3',
    title: 'RED BULL TANDEM SKYDIVE',
    category: 'AIR · EXTREME SPORT',
    description: 'Jump with a Red Bull certified instructor at 15,000ft. Camera-equipped, full debrief, and a story you\'ll never forget.',
    badge: 'LIMITED',
    slotsLeft: 3,
    slotsTotal: 8,
    date: 'Mar 8, 2026',
    duration: '2 Days',
    location: 'Interlaken, Switzerland',
    price: 1200,
    color: 'cyan',
    includes: ['Tandem jump', 'Camera footage', 'Team debrief', 'Red Bull gear'],
  },
  {
    id: '4',
    title: 'FORMULA E GARAGE PASS',
    category: 'MOTORSPORT · TECH ACCESS',
    description: 'Step inside a Formula E garage during race weekend. Engineers will walk you through the EV powertrain and strategy.',
    badge: 'EXCLUSIVE',
    slotsLeft: 8,
    slotsTotal: 15,
    date: 'Apr 5, 2026',
    duration: '2 Days',
    location: 'Monaco',
    price: 3200,
    color: 'blue',
    includes: ['Garage access', 'Tech briefing', 'Pit lane walk', 'Paddock hospitality'],
  },
  {
    id: '5',
    title: 'BAJA 1000 CO-PILOT SEAT',
    category: 'OFF-ROAD · DESERT RACE',
    description: 'Ride as co-pilot through a non-competitive Baja stage. Experience the brutality of off-road racing firsthand.',
    badge: 'LIMITED',
    slotsLeft: 4,
    slotsTotal: 6,
    date: 'Nov 14, 2026',
    duration: '4 Days',
    location: 'Baja California, Mexico',
    price: 4200,
    color: 'gold',
    includes: ['Co-pilot ride', 'Safety briefing', 'HANS device & gear', 'Base camp stay'],
  },
  {
    id: '6',
    title: 'LAKE COMO SPEED BOAT CHARTER',
    category: 'WATER · VIP EXPERIENCE',
    description: 'Follow the speed boat races from a luxury chase boat on Lake Como. Champagne and front-row spray guaranteed.',
    badge: 'EXCLUSIVE',
    slotsLeft: 10,
    slotsTotal: 12,
    date: 'Mar 2, 2026',
    duration: '1 Day',
    location: 'Lake Como, Italy',
    price: 890,
    color: 'purple',
    includes: ['Chase boat charter', 'Champagne service', 'Race commentary', 'Captain briefing'],
  },
]

const badgeColors: Record<string, string> = {
  'EXCLUSIVE': 'var(--cyan)',
  'MEMBERS ONLY': '#facc15',
  'LIMITED': 'var(--red)',
}

const colorMap: Record<string, string> = {
  red: 'var(--red)',
  cyan: 'var(--cyan)',
  orange: '#fb923c',
  blue: '#60a5fa',
  purple: '#a78bfa',
  gold: '#facc15',
}

export default function JourneysPage() {
  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Hero Header */}
        <div className={styles.pageHero}>
          <div className={styles.heroLines}>
            {[...Array(5)].map((_, i) => (
              <div key={i} className={styles.heroLine}></div>
            ))}
          </div>
          <div className={styles.heroContent}>
            <div className={styles.headerTag}>/ JOURNEYS</div>
            <h1 className={styles.pageTitle}>LIVE THE<br /><span className={styles.titleAccent}>EXPERIENCE</span></h1>
            <p className={styles.pageSubtitle}>Exclusive access journeys for those who want more than just a stream. Be there. Feel it. Live it.</p>
          </div>
          <div className={styles.heroStats}>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>24</div>
              <div className={styles.hStatL}>JOURNEYS</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>18</div>
              <div className={styles.hStatL}>COUNTRIES</div>
            </div>
            <div className={styles.hStat}>
              <div className={styles.hStatN}>3</div>
              <div className={styles.hStatL}>AVAILABLE</div>
            </div>
          </div>
        </div>

        {/* Journeys List */}
        <div className={styles.journeysList}>
          {journeys.map((journey, i) => {
            const accent = colorMap[journey.color]
            const slotPct = (journey.slotsLeft / journey.slotsTotal) * 100
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
                    <div className={styles.journeyBadge} style={{ color: badgeColors[journey.badge], borderColor: badgeColors[journey.badge] }}>
                      {journey.badge}
                    </div>
                  </div>

                  <p className={styles.journeyDesc}>{journey.description}</p>

                  <div className={styles.journeyIncludes}>
                    {journey.includes.map((item, j) => (
                      <span key={j} className={styles.includeTag} style={{ borderColor: `${accent}44`, color: accent }}>
                        ✓ {item}
                      </span>
                    ))}
                  </div>

                  <div className={styles.journeyMeta}>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>📅</span>
                      <span>{journey.date}</span>
                    </div>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>⏱</span>
                      <span>{journey.duration}</span>
                    </div>
                    <div className={styles.metaItem}>
                      <span className={styles.metaLabel}>📍</span>
                      <span>{journey.location}</span>
                    </div>
                  </div>
                </div>

                {/* Right: price + slots + CTA */}
                <div className={styles.journeyAside}>
                  <div className={styles.journeyPrice}>
                    <span className={styles.priceFrom}>FROM</span>
                    <span className={styles.priceVal} style={{ color: accent }}>${journey.price.toLocaleString()}</span>
                    <span className={styles.priceLabel}>per person</span>
                  </div>

                  <div className={styles.slotsWrap}>
                    <div className={styles.slotsTop}>
                      <span className={`${styles.slotsLabel} ${isUrgent ? styles.slotsUrgent : ''}`}>
                        {isUrgent ? '🔥 ' : ''}{journey.slotsLeft} SLOTS LEFT
                      </span>
                      <span className={styles.slotsOf}>/ {journey.slotsTotal}</span>
                    </div>
                    <div className={styles.slotsBar}>
                      <div
                        className={styles.slotsFill}
                        style={{ width: `${slotPct}%`, background: isUrgent ? 'var(--red)' : accent }}
                      ></div>
                    </div>
                  </div>

                  <button className={styles.bookBtn} style={{ background: `linear-gradient(135deg, ${accent}22, ${accent}11)`, borderColor: `${accent}66`, color: accent }}>
                    BOOK JOURNEY →
                  </button>
                </div>

                {/* Color accent left border */}
                <div className={styles.journeyBorder} style={{ background: accent, boxShadow: `0 0 12px ${accent}` }}></div>
              </div>
            )
          })}
        </div>
      </main>
    </>
  )
}

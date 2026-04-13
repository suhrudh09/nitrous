import Image from 'next/image'
import Link from 'next/link'
import styles from './Hero.module.css'

const heroNavCards = [
  { label: 'ACCESS\nGARAGE', icon: '🚗', href: '/garage', color: 'grey', progress: 60 },
  { label: 'ACCESS\nEVENT PASSES', icon: '🎫', href: '/passes', color: 'red', progress: 40 },
  { label: 'ACCESS\nLIVE STREAMS', icon: '📺', href: '/live', color: 'cyan', progress: 75 },
  { label: 'TEAMS', icon: '🏆', href: '/teams', color: 'orange', progress: 55 },
  { label: 'JOURNEYS', icon: '🌍', href: '/journeys', color: 'blue', progress: 85 },
  { label: 'MERCH', icon: '👕', href: '/merch', color: 'gold', progress: 45 },
]

export default function Hero() {
  return (
    <section className={styles.hero}>
      {/* Background Image */}
      <div className={styles.heroImgWrap}>
        <Image
          src="/hero-car.png"
          alt="Nitrous wireframe car"
          fill
          priority
          className={styles.heroImg}
        />
      </div>

      {/* Circuit Layer */}
      <div className={styles.circuitLayer}>
        <div className={styles.trace}></div>
        <div className={styles.trace}></div>
        <div className={styles.trace}></div>
        <div className={styles.trace}></div>
        <div className={styles.trace}></div>
        <div className={styles.node}></div>
        <div className={styles.node}></div>
        <div className={styles.node}></div>
        <div className={styles.node}></div>
        <div className={styles.node}></div>
        <div className={styles.energySwirl}></div>
        <div className={styles.energySwirl}></div>
      </div>

      {/* HUD Corners */}
      <div className={`${styles.corner} ${styles.cornerTL}`}>
        <svg viewBox="0 0 34 34" fill="none">
          <path d="M1 33V5L5 1H33" stroke="rgba(0,229,255,0.5)" strokeWidth="1.2" />
        </svg>
      </div>
      <div className={`${styles.corner} ${styles.cornerTR}`}>
        <svg viewBox="0 0 34 34" fill="none" style={{ transform: 'scaleX(-1)' }}>
          <path d="M1 33V5L5 1H33" stroke="rgba(0,229,255,0.5)" strokeWidth="1.2" />
        </svg>
      </div>

      {/* Hero Content */}
      <div className={styles.heroContent}>
        <div className={styles.heroHudLabel}>
          <div className={styles.hudLine}></div>
          <span className={styles.hudDot}></span>
          <span className={styles.hudTxt}>System Online — Daytona 500 Qualifying</span>
        </div>
        
        <h1 className={styles.heroTitle}>
          NITROUS<br />
          <span className={styles.glow}>FUEL</span><br />
          <span className={styles.outline}>YOUR SPEED</span>
        </h1>
        
        <p className={styles.heroSub}>
          Stream every race on the planet. Book VIP passes. Grab team merch. Live exclusive journeys. One platform. Full throttle.
        </p>
        
        <div className={styles.heroActions}>
          <Link href="/live" className={styles.btnNitro}>
            ▶ &nbsp;Ignite Stream
          </Link>
          <Link href="/events" className={styles.btnGhost}>Explore Events</Link>
        </div>
      </div>

      {/* Hero Nav Rail */}
      <div className={styles.heroNavRail}>
        {heroNavCards.map((card, i) => (
          <Link
            key={i}
            href={card.href}
            className={`${styles.hnrCard} ${styles[`hnrCard${card.color.charAt(0).toUpperCase() + card.color.slice(1)}`]}`}
          >
            <div className={styles.hnrIcon}>{card.icon}</div>
            <div className={styles.hnrLabel}>{card.label}</div>
            <div className={styles.hnrBarWrap}>
              <div className={styles.hnrBar} style={{ width: `${card.progress}%` }}></div>
            </div>
          </Link>
        ))}
      </div>
    </section>
  )
}
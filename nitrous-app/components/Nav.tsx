import Link from 'next/link'
import styles from './Nav.module.css'

export default function Nav() {
  return (
    <nav className={styles.nav}>
      <Link href="/" className={styles.logo}>
        NITROUS<span>.</span>
      </Link>
      
      <div className={styles.navCenter}>
        <Link href="/live" className={styles.navLink}>Live</Link>
        <Link href="/events" className={styles.navLink}>Events</Link>
        <Link href="/teams" className={styles.navLink}>Teams</Link>
        <Link href="/journeys" className={styles.navLink}>Journeys</Link>
        <Link href="/merch" className={styles.navLink}>Merch</Link>
      </div>
      
      <div className={styles.navRight}>
        <div className={styles.navStatus}>
          <div className={styles.dotLive} />
          <span>4 Events Live</span>
        </div>
        <Link href="/login" className={styles.btnNav}>Sign In</Link>
      </div>
    </nav>
  )
}

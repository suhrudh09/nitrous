import Nav from '@/components/Nav'
import styles from './teams.module.css'

const teams = [
  {
    id: '1',
    name: 'RED BULL RACING',
    category: 'MOTORSPORT · F1',
    country: '🇦🇹 Austria',
    drivers: ['Max Verstappen', 'Sergio Pérez'],
    wins: 21,
    points: 860,
    color: 'red',
    accentColor: '#1e3a8a',
    rank: 1,
    following: '8.2M',
    founded: 2005,
  },
  {
    id: '2',
    name: 'HENDRICK MOTORSPORTS',
    category: 'MOTORSPORT · NASCAR',
    country: '🇺🇸 USA',
    drivers: ['Kyle Larson', 'Chase Elliott', 'William Byron', 'Alex Bowman'],
    wins: 14,
    points: 2340,
    color: 'cyan',
    accentColor: '#c41e3a',
    rank: 2,
    following: '3.1M',
    founded: 1984,
  },
  {
    id: '3',
    name: 'TOYOTA GAZOO RACING',
    category: 'RALLY · WRC',
    country: '🇯🇵 Japan',
    drivers: ['Sébastien Ogier', 'Elfyn Evans', 'Kalle Rovanperä'],
    wins: 9,
    points: 564,
    color: 'orange',
    accentColor: '#dc2626',
    rank: 3,
    following: '1.8M',
    founded: 1957,
  },
  {
    id: '4',
    name: 'TEAM SEA FORCE',
    category: 'WATER · SPEED BOAT',
    country: '🇮🇹 Italy',
    drivers: ['F. Bertrand', 'L. Capelli'],
    wins: 7,
    points: 320,
    color: 'blue',
    accentColor: '#0e7490',
    rank: 4,
    following: '420K',
    founded: 2010,
  },
  {
    id: '5',
    name: 'FALCON AIR SQUADRON',
    category: 'AIR · AIR RACING',
    country: '🇫🇷 France',
    drivers: ['A. Garnier', 'B. Morin'],
    wins: 5,
    points: 198,
    color: 'purple',
    accentColor: '#7c3aed',
    rank: 5,
    following: '280K',
    founded: 2015,
  },
  {
    id: '6',
    name: 'BAJA IRON SQUAD',
    category: 'OFF-ROAD · TROPHY TRUCK',
    country: '🇲🇽 Mexico',
    drivers: ['C. Wedekin', 'P. McMillin'],
    wins: 12,
    points: 415,
    color: 'gold',
    accentColor: '#92400e',
    rank: 6,
    following: '640K',
    founded: 1998,
  },
]

const colorMap: Record<string, string> = {
  red: 'var(--red)',
  cyan: 'var(--cyan)',
  orange: '#fb923c',
  blue: '#60a5fa',
  purple: '#a78bfa',
  gold: '#facc15',
}

export default function TeamsPage() {
  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Header */}
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ TEAMS</div>
            <h1 className={styles.pageTitle}>COMPETING TEAMS</h1>
            <p className={styles.pageSubtitle}>Follow the world's elite motorsport & racing teams</p>
          </div>
          <div className={styles.headerStats}>
            <div className={styles.stat}><span className={styles.statN}>142</span><span className={styles.statL}>TEAMS</span></div>
            <div className={styles.statDiv}></div>
            <div className={styles.stat}><span className={styles.statN}>28</span><span className={styles.statL}>COUNTRIES</span></div>
            <div className={styles.statDiv}></div>
            <div className={styles.stat}><span className={styles.statN}>6</span><span className={styles.statL}>CATEGORIES</span></div>
          </div>
        </div>

        {/* Ranking Banner */}
        <div className={styles.rankBanner}>
          <span className={styles.rankBannerLabel}>CURRENT SEASON STANDINGS</span>
          <div className={styles.rankBannerDivider}></div>
          {teams.slice(0,3).map(t => (
            <div key={t.id} className={styles.rankBannerItem}>
              <span className={styles.rankBannerPos}>#{t.rank}</span>
              <span className={styles.rankBannerName}>{t.name}</span>
            </div>
          ))}
        </div>

        {/* Teams Grid */}
        <div className={styles.teamsGrid}>
          {teams.map((team) => {
            const accentRgb = colorMap[team.color] || 'var(--cyan)'
            return (
              <div key={team.id} className={styles.teamCard}>
                {/* Top accent */}
                <div className={styles.cardAccent} style={{ background: accentRgb, boxShadow: `0 0 10px ${accentRgb}` }}></div>

                {/* Rank badge */}
                <div className={styles.rankBadge} style={{ color: accentRgb, borderColor: accentRgb }}>
                  #{team.rank}
                </div>

                {/* Team Logo Area */}
                <div className={styles.logoArea} style={{ background: `linear-gradient(135deg, ${team.accentColor}33 0%, transparent 60%)` }}>
                  <div className={styles.logoCircle} style={{ borderColor: accentRgb }}>
                    <span className={styles.logoInitials}>{team.name.split(' ').map(w => w[0]).join('').slice(0,3)}</span>
                  </div>
                </div>

                <div className={styles.teamInfo}>
                  <div className={styles.teamCat} style={{ color: accentRgb }}>{team.category}</div>
                  <div className={styles.teamName}>{team.name}</div>
                  <div className={styles.teamCountry}>{team.country} · Est. {team.founded}</div>
                </div>

                {/* Drivers */}
                <div className={styles.driversSection}>
                  <div className={styles.driversLabel}>ROSTER</div>
                  <div className={styles.driversList}>
                    {team.drivers.map((d, i) => (
                      <div key={i} className={styles.driverChip}>{d}</div>
                    ))}
                  </div>
                </div>

                {/* Stats */}
                <div className={styles.statsRow}>
                  <div className={styles.teamStat}>
                    <span className={styles.teamStatN}>{team.wins}</span>
                    <span className={styles.teamStatL}>WINS</span>
                  </div>
                  <div className={styles.teamStatDiv}></div>
                  <div className={styles.teamStat}>
                    <span className={styles.teamStatN}>{team.points}</span>
                    <span className={styles.teamStatL}>PTS</span>
                  </div>
                  <div className={styles.teamStatDiv}></div>
                  <div className={styles.teamStat}>
                    <span className={styles.teamStatN}>{team.following}</span>
                    <span className={styles.teamStatL}>FOLLOWING</span>
                  </div>
                </div>

                {/* Action */}
                <button className={styles.followBtn} style={{ borderColor: accentRgb, color: accentRgb }}>
                  + FOLLOW TEAM
                </button>
              </div>
            )
          })}
        </div>
      </main>
    </>
  )
}

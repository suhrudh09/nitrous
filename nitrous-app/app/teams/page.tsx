'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import { getTeams, followTeam, unfollowTeam } from '@/lib/api'
import styles from './teams.module.css'

const colorMap: Record<string, string> = {
  red: 'var(--red)',
  cyan: 'var(--cyan)',
  orange: '#fb923c',
  blue: '#60a5fa',
  purple: '#a78bfa',
  gold: '#facc15',
}

type TeamCard = {
  id: string
  name: string
  category: string
  country: string
  founded: string
  rank: number
  wins: number
  points: number
  following: number
  drivers: string[]
  color: string
  accentColor: string
}

function normalizeTeams(input: any[]): TeamCard[] {
  return input.map((team, index) => {
    const color = typeof team?.color === 'string' ? team.color : 'cyan'
    const followersCount =
      typeof team?.following === 'number'
        ? team.following
        : typeof team?.followersCount === 'number'
        ? team.followersCount
        : 0

    return {
      id: String(team?.id ?? `team-${index}`),
      name: typeof team?.name === 'string' && team.name.trim() ? team.name : 'Unknown Team',
      category: typeof team?.category === 'string' ? team.category : 'MOTORSPORT',
      country: typeof team?.country === 'string' && team.country.trim() ? team.country : 'Global',
      founded: typeof team?.founded === 'number' ? String(team.founded) : 'N/A',
      rank: typeof team?.rank === 'number' ? team.rank : index + 1,
      wins: typeof team?.wins === 'number' ? team.wins : 0,
      points: typeof team?.points === 'number' ? team.points : 0,
      following: followersCount,
      drivers: Array.isArray(team?.drivers) ? team.drivers.filter((d: unknown) => typeof d === 'string') : [],
      color,
      accentColor: typeof team?.accentColor === 'string' ? team.accentColor : colorMap[color] || 'var(--cyan)',
    }
  })
}

export default function TeamsPage() {
  const [teams, setTeams] = useState<TeamCard[]>([])
  const [loading, setLoading] = useState(true)
  const [followingIds, setFollowingIds] = useState<Set<string>>(new Set())
  const [processingId, setProcessingId] = useState<string | null>(null)

  useEffect(() => {
    async function fetchData() {
      try {
        const data = await getTeams()
        setTeams(normalizeTeams(data))
      } catch (error) {
        console.error('Failed to fetch teams:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  const handleFollow = async (teamId: string) => {
    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      window.location.href = '/login'
      return
    }

    setProcessingId(teamId)
    try {
      if (followingIds.has(teamId)) {
        await unfollowTeam(teamId, token)
        setFollowingIds(prev => {
          const newSet = new Set(prev)
          newSet.delete(teamId)
          return newSet
        })
      } else {
        await followTeam(teamId, token)
        setFollowingIds(prev => new Set([...prev, teamId]))
      }
    } catch (error) {
      console.error('Failed to toggle follow:', error)
    } finally {
      setProcessingId(null)
    }
  }
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
                    <span className={styles.logoInitials}>{team.name.split(' ').map((w: string) => w[0]).join('').slice(0,3)}</span>
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
                    {team.drivers.length > 0 ? (
                      team.drivers.map((d, i) => (
                        <div key={i} className={styles.driverChip}>{d}</div>
                      ))
                    ) : (
                      <div className={styles.driverChip}>Roster updating</div>
                    )}
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
                <button 
                  className={styles.followBtn} 
                  style={{ borderColor: accentRgb, color: accentRgb }}
                  onClick={() => handleFollow(team.id)}
                  disabled={processingId === team.id}
                >
                  {processingId === team.id 
                    ? '⏳ ...' 
                    : followingIds.has(team.id) 
                    ? '✓ FOLLOWING' 
                    : '+ FOLLOW TEAM'
                  }
                </button>
              </div>
            )
          })}
        </div>
      </main>
    </>
  )
}

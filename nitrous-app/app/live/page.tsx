'use client'
import Link from 'next/link'
import { useState } from 'react'
import Nav from '@/components/Nav'
import styles from './live.module.css'

const liveStreams = [
  {
    id: '1',
    title: 'NASCAR Daytona 500',
    subtitle: 'Lap 87 / 200',
    category: 'MOTORSPORT',
    location: 'Daytona International Speedway · FL',
    viewers: '1.2M',
    quality: '4K',
    color: 'red',
    leader: 'Bubba Wallace #23',
    speed: '198 mph',
    hot: true,
  },
  {
    id: '2',
    title: 'World Dirt Track Championship',
    subtitle: 'Heat 3 — Semi Finals',
    category: 'MOTORSPORT',
    location: 'Knob Noster · Missouri, USA',
    viewers: '340K',
    quality: 'HD',
    color: 'orange',
    leader: 'Kyle Larson #57',
    speed: '142 mph',
    hot: false,
  },
  {
    id: '3',
    title: 'Lake Como Speed Boat Qualifier',
    subtitle: 'Qualifying Round 2',
    category: 'WATER',
    location: 'Lake Como · Italy',
    viewers: '89K',
    quality: 'HD',
    color: 'cyan',
    leader: 'F. Bertrand #9',
    speed: '87 knots',
    hot: false,
  },
  {
    id: '4',
    title: 'Red Bull Skydive Series',
    subtitle: 'Live Drop — 14,800ft',
    category: 'AIR',
    location: 'Interlaken Drop Zone · Switzerland',
    viewers: '220K',
    quality: 'HD',
    color: 'purple',
    leader: 'A. Garnier',
    speed: '120 mph',
    hot: false,
  },
]

const upcomingStreams = [
  { title: 'Baja 1000 — Stage 4', time: 'In 2h 14m', category: 'OFF-ROAD' },
  { title: 'Crop Duster Air Racing', time: 'In 5h 30m', category: 'AIR' },
  { title: 'Jet Ski Championship — Finals', time: 'Tomorrow 14:00', category: 'WATER' },
  { title: 'Formula E Round 6 — Monaco', time: 'Tomorrow 18:00', category: 'MOTORSPORT' },
]

export default function LivePage() {
  const [active, setActive] = useState('1')
  const featured = liveStreams.find(s => s.id === active) || liveStreams[0]

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Page Header */}
        <div className={styles.pageHeader}>
          <div className={styles.headerLeft}>
            <div className={styles.liveChip}>
              <span className={styles.liveDot}></span>
              <span>LIVE NOW</span>
            </div>
            <h1 className={styles.pageTitle}>STREAMS</h1>
            <p className={styles.pageSubtitle}>
              {liveStreams.length} channels broadcasting · {liveStreams.reduce((a, s) => a + parseFloat(s.viewers), 0).toFixed(1)}M+ watching
            </p>
          </div>
          <div className={styles.headerStats}>
            <div className={styles.stat}>
              <span className={styles.statVal}>4</span>
              <span className={styles.statLabel}>LIVE</span>
            </div>
            <div className={styles.statDivider}></div>
            <div className={styles.stat}>
              <span className={styles.statVal}>1.8M</span>
              <span className={styles.statLabel}>VIEWERS</span>
            </div>
            <div className={styles.statDivider}></div>
            <div className={styles.stat}>
              <span className={styles.statVal}>3</span>
              <span className={styles.statLabel}>UPCOMING</span>
            </div>
          </div>
        </div>

        <div className={styles.grid}>
          {/* Main Player */}
          <div className={styles.playerWrap}>
            <div className={`${styles.player} ${styles[`player${featured.color.charAt(0).toUpperCase() + featured.color.slice(1)}`]}`}>
              {/* Fake video overlay */}
              <div className={styles.playerNoise}></div>
              <div className={styles.playerScanlines}></div>

              {/* Top HUD */}
              <div className={styles.playerHudTop}>
                <div className={styles.playerLiveBadge}>
                  <span className={styles.liveDot}></span>LIVE
                </div>
                <span className={styles.playerCat}>{featured.category}</span>
                <div className={styles.playerQuality}>{featured.quality}</div>
              </div>

              {/* Center play */}
              <div className={styles.playerCenter}>
                <div className={styles.playBtn}>
                  <span>▶</span>
                </div>
              </div>

              {/* Bottom HUD */}
              <div className={styles.playerHudBot}>
                <div className={styles.playerInfo}>
                  <div className={styles.playerTitle}>{featured.title}</div>
                  <div className={styles.playerSub}>{featured.subtitle}</div>
                </div>
                <div className={styles.playerMeta}>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>LEADER</span>
                    <span className={styles.metaVal}>{featured.leader}</span>
                  </div>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>SPEED</span>
                    <span className={styles.metaVal}>{featured.speed}</span>
                  </div>
                  <div className={styles.metaItem}>
                    <span className={styles.metaLabel}>VIEWERS</span>
                    <span className={styles.metaVal}>{featured.viewers}</span>
                  </div>
                </div>
              </div>

              {/* Progress bar */}
              <div className={styles.playerProgress}>
                <div className={styles.playerProgressFill}></div>
              </div>
            </div>

            {/* Stream controls */}
            <div className={styles.controls}>
              <button className={styles.controlBtn}>⏮</button>
              <button className={styles.controlBtn}>⏸</button>
              <button className={styles.controlBtn}>⏭</button>
              <div className={styles.controlSep}></div>
              <button className={styles.controlBtn}>🔇</button>
              <div className={styles.volumeBar}>
                <div className={styles.volumeFill}></div>
              </div>
              <div className={styles.controlSep}></div>
              <button className={`${styles.controlBtn} ${styles.controlBtnRight}`}>⛶</button>
            </div>
          </div>

          {/* Sidebar */}
          <div className={styles.sidebar}>
            <div className={styles.sideSection}>
              <div className={styles.sideLabel}>LIVE CHANNELS</div>
              <div className={styles.channelList}>
                {liveStreams.map(stream => (
                  <div
                    key={stream.id}
                    className={`${styles.channelCard} ${active === stream.id ? styles.channelCardActive : ''} ${styles[`channelCard${stream.color.charAt(0).toUpperCase() + stream.color.slice(1)}`]}`}
                    onClick={() => setActive(stream.id)}
                  >
                    <div className={styles.channelTop}>
                      <div className={styles.channelLive}>
                        <span className={styles.liveDotSm}></span>
                        {stream.category}
                      </div>
                      <div className={styles.channelViewers}>👁 {stream.viewers}</div>
                    </div>
                    <div className={styles.channelTitle}>{stream.title}</div>
                    <div className={styles.channelSub}>{stream.subtitle} · {stream.location}</div>
                  </div>
                ))}
              </div>
            </div>

            <div className={styles.sideSection}>
              <div className={styles.sideLabel}>UP NEXT</div>
              <div className={styles.upcomingList}>
                {upcomingStreams.map((s, i) => (
                  <div key={i} className={styles.upcomingItem}>
                    <div className={styles.upcomingTime}>{s.time}</div>
                    <div className={styles.upcomingTitle}>{s.title}</div>
                    <div className={styles.upcomingCat}>{s.category}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}

'use client'
import { useState } from 'react'
import Nav from '@/components/Nav'
import styles from './merch.module.css'

const products = [
  { id: '1', name: 'NITROUS Team Hoodie', icon: '👕', price: 89, category: 'apparel', tag: 'BESTSELLER', desc: 'Heavyweight 400gsm cotton. Embroidered logo.', sizes: ['S','M','L','XL','XXL'], colors: ['#1a1a2e', '#2d1b1b', '#1a2d1b'] },
  { id: '2', name: 'NITROUS Race Cap', icon: '🧢', price: 42, category: 'apparel', tag: null, desc: 'Structured 6-panel. Laser-perforated.', sizes: ['ONE SIZE'], colors: ['#111827', '#7f1d1d'] },
  { id: '3', name: 'Racing Jacket — Pro Cut', icon: '🏎️', price: 189, category: 'apparel', tag: 'NEW', desc: 'Windproof shell. Reflective detailing.', sizes: ['S','M','L','XL'], colors: ['#020408', '#1e3a8a'] },
  { id: '4', name: 'Paddock Long Sleeve', icon: '👔', price: 64, category: 'apparel', tag: null, desc: 'Moisture-wicking technical fabric.', sizes: ['S','M','L','XL','XXL'], colors: ['#020408', '#0a1220'] },
  { id: '5', name: 'Pit Watch — Chrono', icon: '⌚', price: 249, category: 'accessories', tag: 'LIMITED', desc: 'Sapphire glass. 200m water resistant.', sizes: null, colors: ['#1c1c1c', '#111827'] },
  { id: '6', name: 'Gear Backpack 28L', icon: '🎒', price: 120, category: 'accessories', tag: null, desc: 'MOLLE webbing. Laptop sleeve. Waterproof.', sizes: null, colors: ['#111827', '#1a1a2e'] },
  { id: '7', name: 'Carbon Fiber Wallet', icon: '💳', price: 68, category: 'accessories', tag: null, desc: 'RFID-blocking. 8-card capacity.', sizes: null, colors: ['#1c1c1c'] },
  { id: '8', name: 'Podium Trophy — Limited', icon: '🏆', price: 380, category: 'collectibles', tag: 'LIMITED', desc: 'Hand-cast resin. Numbered 1–50.', sizes: null, colors: null },
  { id: '9', name: 'NITROUS Keychain', icon: '🔑', price: 28, category: 'collectibles', tag: null, desc: 'Die-cast zinc. Enamel fill.', sizes: null, colors: null },
  { id: '10', name: 'Enamel Pin Set (6pc)', icon: '📌', price: 38, category: 'collectibles', tag: 'NEW', desc: 'Hard enamel. Gold-plated backing.', sizes: null, colors: null },
  { id: '11', name: 'NITROUS Decal Pack', icon: '🎨', price: 16, category: 'collectibles', tag: null, desc: 'UV resistant vinyl. 12 stickers.', sizes: null, colors: null },
  { id: '12', name: 'Water Bottle 750ml', icon: '🧴', price: 44, category: 'accessories', tag: null, desc: 'Double-wall insulated. 24hr cold.', sizes: null, colors: ['#020408', '#0a1220', '#7f1d1d'] },
]

const cats = ['all', 'apparel', 'accessories', 'collectibles']

const tagColors: Record<string, string> = {
  'BESTSELLER': 'var(--cyan)',
  'NEW': '#4ade80',
  'LIMITED': 'var(--red)',
}

export default function MerchPage() {
  const [filter, setFilter] = useState('all')
  const [cart, setCart] = useState<string[]>([])
  const [added, setAdded] = useState<string | null>(null)

  const filtered = filter === 'all' ? products : products.filter(p => p.category === filter)

  function addToCart(id: string) {
    setCart(prev => [...prev, id])
    setAdded(id)
    setTimeout(() => setAdded(null), 1200)
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Header */}
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ MERCH</div>
            <h1 className={styles.pageTitle}>TEAM STORE</h1>
            <p className={styles.pageSubtitle}>Official gear, limited drops, and collector pieces</p>
          </div>
          <div className={styles.cartBadge}>
            <span className={styles.cartIcon}>🛒</span>
            <span className={styles.cartCount}>{cart.length}</span>
            <span className={styles.cartLabel}>CART</span>
          </div>
        </div>

        {/* Filter tabs */}
        <div className={styles.filterBar}>
          {cats.map(cat => (
            <button
              key={cat}
              className={`${styles.catBtn} ${filter === cat ? styles.catBtnActive : ''}`}
              onClick={() => setFilter(cat)}
            >
              {cat === 'all' ? 'ALL PRODUCTS' : cat.toUpperCase()}
            </button>
          ))}
          <div className={styles.filterRight}>
            <span className={styles.countTxt}>{filtered.length} items</span>
          </div>
        </div>

        {/* Products Grid */}
        <div className={styles.productsGrid}>
          {filtered.map(product => (
            <div key={product.id} className={styles.productCard}>
              {/* Tag */}
              {product.tag && (
                <div className={styles.productTag} style={{ color: tagColors[product.tag], borderColor: `${tagColors[product.tag]}55` }}>
                  {product.tag}
                </div>
              )}

              {/* Product visual */}
              <div className={styles.productVisual}>
                <div className={styles.productIcon}>{product.icon}</div>
                {/* Color swatches */}
                {product.colors && (
                  <div className={styles.swatches}>
                    {product.colors.map((c, i) => (
                      <div key={i} className={styles.swatch} style={{ background: c }}></div>
                    ))}
                  </div>
                )}
              </div>

              <div className={styles.productInfo}>
                <div className={styles.productCat}>{product.category}</div>
                <div className={styles.productName}>{product.name}</div>
                <div className={styles.productDesc}>{product.desc}</div>

                {/* Sizes */}
                {product.sizes && (
                  <div className={styles.sizes}>
                    {product.sizes.map(s => (
                      <span key={s} className={styles.sizeChip}>{s}</span>
                    ))}
                  </div>
                )}
              </div>

              <div className={styles.productBottom}>
                <div className={styles.productPrice}>${product.price}</div>
                <button
                  className={`${styles.addBtn} ${added === product.id ? styles.addBtnSuccess : ''}`}
                  onClick={() => addToCart(product.id)}
                >
                  {added === product.id ? '✓ ADDED' : '+ ADD TO CART'}
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Bottom banner */}
        <div className={styles.promoBanner}>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>MEMBERS GET 15% OFF</span>
            <span className={styles.promoCta}>Sign in or create an account →</span>
          </div>
          <div className={styles.promoDivider}></div>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>FREE SHIPPING OVER $100</span>
            <span className={styles.promoCta}>Worldwide delivery available</span>
          </div>
          <div className={styles.promoDivider}></div>
          <div className={styles.promoText}>
            <span className={styles.promoLabel}>LIMITED DROPS</span>
            <span className={styles.promoCta}>Subscribe for early access</span>
          </div>
        </div>
      </main>
    </>
  )
}

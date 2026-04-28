import React from 'react'
import Hero from '@/components/Hero'

describe('Hero Component', () => {
  before(() => {
    if (!document.getElementById('__next_css__DO_NOT_USE__')) {
      const anchor = document.createElement('div')
      anchor.id = '__next_css__DO_NOT_USE__'
      document.head.appendChild(anchor)
    }
  })

  beforeEach(() => {
    cy.mount(<Hero />)
  })

  describe('Structure', () => {
    it('renders hero section', () => {
      cy.get('section').should('exist')
    })

    it('renders background image wrapper', () => {
      cy.get('[class*="heroImgWrap"]').should('exist')
    })

    it('renders circuit layer', () => {
      cy.get('[class*="circuitLayer"]').should('exist')
    })

    it('renders hero content container', () => {
      cy.get('[class*="heroContent"]').should('exist')
    })

    it('renders hero nav rail', () => {
      cy.get('[class*="heroNavRail"]').should('exist')
    })
  })

  describe('Background and Visual Elements', () => {
    it('displays the background image', () => {
      cy.get('img[alt="Nitrous wireframe car"]').should('exist')
    })

    it('renders circuit layer traces', () => {
      cy.get('[class*="trace"]').should('have.length', 5)
    })

    it('renders circuit layer nodes', () => {
      cy.get('[class*="node"]').should('have.length', 5)
    })

    it('renders energy swirl animations', () => {
      cy.get('[class*="energySwirl"]').should('have.length', 2)
    })

    it('renders HUD corner elements', () => {
      cy.get('[class*="corner"]').should('have.length', 2)
      cy.get('svg').should('have.length.greaterThan', 0)
    })
  })

  describe('HUD and Text Content', () => {
    it('displays HUD label with system status', () => {
      cy.get('[class*="heroHudLabel"]').should('exist')
      cy.contains('System Online').should('be.visible')
    })

    it('displays event qualify text in HUD', () => {
      cy.contains('Daytona 500').should('be.visible')
    })

    it('renders HUD line element', () => {
      cy.get('[class*="hudLine"]').should('exist')
    })

    it('renders HUD dot indicator', () => {
      cy.get('[class*="hudDot"]').should('exist')
    })
  })

  describe('Hero Title and Subtitle', () => {
    it('displays main title with NITROUS', () => {
      cy.get('h1').should('contain', 'NITROUS')
    })

    it('displays FUEL with glow styling', () => {
      cy.get('[class*="glow"]').should('contain', 'FUEL')
    })

    it('displays YOUR SPEED with outline styling', () => {
      cy.get('[class*="outline"]').should('contain', 'YOUR SPEED')
    })

    it('displays subtitle text', () => {
      cy.contains('Stream every race on the planet').should('be.visible')
      cy.contains('Book VIP passes').should('be.visible')
      cy.contains('Full throttle').should('be.visible')
    })
  })

  describe('Action Buttons', () => {
    it('renders Ignite Stream button', () => {
      cy.contains('Ignite Stream').should('be.visible')
      cy.contains('Ignite Stream').should('match', 'a')
    })

    it('renders Explore Events button', () => {
      cy.contains('Explore Events').should('be.visible')
      cy.contains('Explore Events').should('match', 'a')
    })

    it('buttons are not disabled', () => {
      cy.get('[class*="btnNitro"]').should('not.have.attr', 'disabled')
      cy.get('[class*="btnGhost"]').should('not.have.attr', 'disabled')
    })

    it('buttons have correct styling classes', () => {
      cy.get('[class*="btnNitro"]').should('exist')
      cy.get('[class*="btnGhost"]').should('exist')
    })
  })

  describe('Hero Navigation Cards', () => {
    it('renders hero nav rail with cards', () => {
      cy.get('[class*="hnrCard"]').should('have.length', 6)
    })

    it('renders ACCESS GARAGE card with correct href', () => {
      cy.get('a[href="/garage"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '🚗')
      })
    })

    it('renders ACCESS EVENT PASSES card with correct href', () => {
      cy.get('a[href="/passes"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '🎫')
      })
    })

    it('renders ACCESS LIVE STREAMS card with correct href', () => {
      cy.get('[class*="heroNavRail"] a[href="/live"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '📺')
      })
    })

    it('renders TEAMS card with correct href and icon', () => {
      cy.get('a[href="/teams"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '🏆')
      })
    })

    it('renders JOURNEYS card with correct href and icon', () => {
      cy.get('a[href="/journeys"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '🌍')
      })
    })

    it('renders MERCH card with correct href and icon', () => {
      cy.get('a[href="/merch"]').within(() => {
        cy.get('[class*="hnrIcon"]').should('contain', '👕')
      })
    })

    it('each card has progress bar', () => {
      cy.get('[class*="hnrBarWrap"]').should('have.length', 6)
      cy.get('[class*="hnrBar"]:not([class*="hnrBarWrap"])').should('have.length', 6)
    })

    it('cards have label text', () => {
      cy.get('[class*="hnrLabel"]').should('have.length', 6)
    })
  })

  describe('Styling and Classes', () => {
    it('hero section has hero class', () => {
      cy.get('section[class*="hero"]').should('exist')
    })

    it('hero actions container exists', () => {
      cy.get('[class*="heroActions"]').should('exist')
    })

    it('all cards have styled color classes', () => {
      cy.get('[class*="hnrCard"]').each(($card) => {
        expect($card.attr('class')).to.match(/hnrCard(Grey|Red|Cyan|Orange|Blue|Gold)/)
      })
    })
  })
})
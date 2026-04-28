import React from 'react'
import Nav from '@/components/Nav'
import { it } from 'node:test'

const MOCK_USER = {
  id: 'user-001',
  name: 'Alex Rider',
  email: 'alex@nitrous.io',
}

const MOCK_TOKEN = 'mock-jwt-token-abc123'

function seedAuth() {
  window.localStorage.setItem('nitrous_token', MOCK_TOKEN)
  window.localStorage.setItem('nitrous_user', JSON.stringify(MOCK_USER))
}

function clearAuth() {
  window.localStorage.removeItem('nitrous_token')
  window.localStorage.removeItem('nitrous_user')
}

describe('Nav Component', () => {
  before(() => {
    if (!document.getElementById('__next_css__DO_NOT_USE__')) {
      const anchor = document.createElement('div')
      anchor.id = '__next_css__DO_NOT_USE__'
      document.head.appendChild(anchor)
    }
  })

  describe('Logged out', () => {
    beforeEach(() => {
      clearAuth()
      cy.mount(<Nav />)
    })

    it('renders the NITROUS logo', () => {
      cy.contains('NITROUS').should('be.visible')
    })

    it('logo links to /', () => {
      cy.get('a[href="/"]').should('exist')
    })

    it('shows nav links: Live, Events, Teams, Journeys, Merch', () => {
      cy.contains('a', 'Live').should('be.visible')
      cy.contains('a', 'Events').should('be.visible')
      cy.contains('a', 'Teams').should('be.visible')
      cy.contains('a', 'Journeys').should('be.visible')
      cy.contains('a', 'Merch').should('be.visible')
    })

    it('nav links have correct hrefs', () => {
      cy.get('a[href="/live"]').should('exist')
      cy.get('a[href="/events"]').should('exist')
      cy.get('a[href="/teams"]').should('exist')
      cy.get('a[href="/journeys"]').should('exist')
      cy.get('a[href="/merch"]').should('exist')
    })

    it('shows 4 Events Live status badge', () => {
      cy.contains('4 Events Live').should('be.visible')
    })

    it('shows Sign In button when logged out', () => {
      cy.contains('Sign In').should('be.visible')
    })

    it('does not show user avatar when logged out', () => {
      cy.get('[aria-label="User menu"]').should('not.exist')
    })
  })

  describe('Logged in', () => {
    beforeEach(() => {
      seedAuth()
      cy.mount(<Nav />)
    })

    afterEach(() => {
      clearAuth()
    })

    it('shows user initials avatar instead of Sign In', () => {
      cy.contains('Sign In').should('not.exist')
      cy.contains('AR').should('be.visible')
    })

    it('renders user menu button with aria-label', () => {
      cy.get('[aria-label="User menu"]').should('exist')
    })

    it('opens dropdown on avatar click', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.contains('Alex Rider').should('be.visible')
      cy.contains('alex@nitrous.io').should('be.visible')
    })

    it('dropdown contains all menu items', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.contains('My Garage').should('be.visible')
      cy.contains('My Passes').should('be.visible')
      cy.contains('My Orders').should('be.visible')
      cy.contains('My Journeys').should('be.visible')
      cy.contains('Settings').should('be.visible')
      cy.contains('Sign Out').should('be.visible')
    })

    it('dropdown menu items have correct hrefs', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.get('a[href="/garage"]').should('exist')
      cy.get('a[href="/passes"]').should('exist')
      cy.get('a[href="/orders"]').should('exist')
      cy.get('a[href="/settings"]').should('exist')
    })

    it('closes dropdown on outside click', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.contains('Alex Rider').should('be.visible')
      cy.get('body').click(0, 0)
      cy.contains('Alex Rider').should('not.exist')
    })

    it('closes dropdown when a menu item is clicked', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.contains('My Garage').click()
      cy.contains('Alex Rider').should('not.exist')
    })

    it('sign out clears user and shows Sign In', () => {
      cy.get('[aria-label="User menu"]').click()
      cy.contains('Sign Out').click()
      cy.contains('Sign In').should('be.visible')
      cy.get('[aria-label="User menu"]').should('not.exist')
    })
  })

  describe('Structure', () => {
    beforeEach(() => {
      clearAuth()
      cy.mount(<Nav />)
    })

    it('renders a nav element', () => {
      cy.get('nav').should('exist')
    })

    it('renders nav center links container', () => {
      cy.get('nav a[href="/live"]').should('exist')
    })

    it('renders nav right section', () => {
      cy.contains('4 Events Live').should('exist')
    })
  })
})
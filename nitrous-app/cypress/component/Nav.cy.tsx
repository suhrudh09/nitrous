import React from 'react';
import Nav from '@/components/Nav'

describe('Nav Component', () => {
  before(() => {
    // FIX: Next.js looks for this specific element to inject CSS. 
    // If it's missing, the 'parentNode' error triggers.
    if (!document.getElementById('__next_css__DO_NOT_USE__')) {
      const anchor = document.createElement('div');
      anchor.id = '__next_css__DO_NOT_USE__';
      document.head.appendChild(anchor);
    }
  })

  beforeEach(() => {
    // Ensure we are mounting the component correctly for React 18
    cy.mount(<Nav />)
  })

  it('renders the navigation element', () => {
    cy.get('nav').should('exist')
  })

  it('displays the NITROUS logo', () => {
    cy.contains('NITROUS').should('be.visible')
  })

  it('renders all 5 navigation links', () => {
    cy.get('[class*="navLink"]').should('have.length', 5)
  })

  it('displays the live event status', () => {
    cy.contains('4 Events Live').should('be.visible')
  })

  it('renders the Sign In button', () => {
    cy.get('button').contains('Sign In').should('be.visible')
  })
})
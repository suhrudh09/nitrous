describe('Hero Section Interactions', () => {
  beforeEach(() => {
    cy.visit('http://localhost:3000')
  })

  it('displays hero action buttons and verifies they are clickable', () => {
    const igniteButton = cy.contains('Ignite Stream')
    igniteButton.should('be.visible')
    igniteButton.should('not.be.disabled')
  })

  it('verifies all navigation cards are visible in hero section', () => {
    cy.contains('GARAGE').should('be.visible')
    cy.contains('EVENT PASSES').should('be.visible')
    cy.contains('LIVE STREAMS').should('be.visible')
    cy.contains('TEAMS').should('be.visible')
    cy.contains('JOURNEYS').should('be.visible')
    cy.contains('MERCH').should('be.visible')
  })

  it('can click the Ignite Stream button', () => {
    cy.contains('Ignite Stream').click()
    cy.contains('NITROUS').should('be.visible')
  })

  it('can click the Explore Events button', () => {
    cy.contains('Explore Events').click()
    cy.contains('NITROUS').should('be.visible')
  })

  it('can navigate through hero nav cards', () => {
    cy.contains('GARAGE').click()
    cy.url().should('include', '/garage')
  })

  it('can navigate to live streams from hero', () => {
    cy.contains('LIVE STREAMS').click()
    cy.url().should('include', '/live')
  })

  it('verifies hero section styling elements exist', () => {
    cy.get('h1').should('be.visible')
    cy.contains('Ignite Stream').should('be.visible')
    cy.contains('Explore Events').should('be.visible')
  })
})
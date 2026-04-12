describe('Home Page Navigation', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  it('loads the home page successfully', () => {
    cy.contains('NITROUS').should('be.visible')
  })

  it('displays navigation menu with all links', () => {
    cy.contains('Live').should('be.visible')
    cy.contains('Events').should('be.visible')
    cy.contains('Teams').should('be.visible')
    cy.contains('Journeys').should('be.visible')
    cy.contains('Merch').should('be.visible')
  })

  it('displays Sign In link', () => {
    cy.contains('a', 'Sign In').should('be.visible')
  })

  it('displays hero title and subtitle', () => {
    cy.contains('NITROUS').should('be.visible')
    cy.contains('FUEL').should('be.visible')
    cy.contains('SPEED').should('be.visible')
    cy.contains('Stream every race on the planet').should('be.visible')
  })

  it('displays action buttons', () => {
    cy.contains('Ignite Stream').should('be.visible')
    cy.contains('Explore Events').should('be.visible')
  })

  it('navigates to Events page when clicking Events link', () => {
    cy.contains('Events').click()
    cy.url().should('include', '/events')
  })

  it('navigates to Teams page when clicking Teams link', () => {
    cy.contains('Teams').click()
    cy.url().should('include', '/teams')
  })

  it('navigates to Journeys page when clicking Journeys link', () => {
    cy.contains('Journeys').click()
    cy.url().should('include', '/journeys')
  })

  it('navigates to Merch page when clicking Merch link', () => {
    cy.contains('Merch').click()
    cy.url().should('include', '/merch')
  })

  it('displays live events status', () => {
    cy.contains('4 Events Live').should('be.visible')
  })
})

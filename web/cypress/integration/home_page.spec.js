describe('Home page test', () => {
  it('Visits home page and verifies the user is not logged in.', () => {
    cy.visit('/')
    cy.contains('Login')
    cy.contains('Signup')
  })

  it('Visits home page with logged in user and displays exchange link', () => {
    cy.loginByAuth0Api(
      Cypress.env('auth0_username'),
      Cypress.env('auth0_password')
    )
    cy.contains('Exchange')
  })
})
  
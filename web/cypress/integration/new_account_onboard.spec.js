describe('New Account Initial State', () => {
  before(() => {
    cy.intercept('GET', '**/account').as('getAccounts')
    cy.intercept('GET', '**/account/*').as('getAccount')
    cy.loginByAuth0Api(
      Cypress.env('auth0_username'),
      Cypress.env('auth0_password')
    )
    
    cy.wait('@getAccounts').its('response.statusCode').should('be.oneOf', [200])
    cy.wait('@getAccount').its('response.statusCode').should('be.oneOf', [200])
  })

  it('Has initial account data', () => {
    cy.get('header').contains('Exchange').click()
    cy.get('h5').contains('Assets').parent('div').within(() => {
      cy.get('ul').within(() => {

        cy.contains('Bitcoin')
        cy.contains('0.00000000 BTC')
        cy.contains('Ethereum')
        cy.contains('0.000000000000000000 ETH')
      })
    })
  })

  it('Has Market Order Form Active', () => {
    cy.get('h5').contains('Create New Order').parent('div').within(() => {
      cy.contains('Buy').closest('button').should('have.attr', 'aria-checked', 'true')
      cy.contains('Market Order')
      cy.get('input[type="radio"][value="ETH"]').should('be.checked')
    })
  })
})

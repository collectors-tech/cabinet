import { mount } from 'cypress/react'
import { RecordActionContractDemo } from '@/components/data-table/record-action-contract-demo'

describe('RecordActionContractDemo component', () => {
  it('documents the shared menu, edit dialog, and destructive confirmation standard', () => {
    mount(<RecordActionContractDemo />)

    cy.contains('Record action contract demo').should('be.visible')
    cy.contains('td', 'Charizard Base Set').parent('tr').within(() => {
      cy.get('button[aria-label="Open actions for Charizard Base Set"]').click()
    })

    cy.get('[role="menuitem"]').then(($items) => {
      expect($items.map((_, item) => item.textContent?.trim()).get()).to.eql([
        'Open details',
        'Edit',
        'DuplicateRequires write permission',
        'Archive',
        'Delete',
        'Permanent delete',
      ])
    })
    cy.contains('[role="menuitem"]', 'Duplicate')
      .should('have.attr', 'aria-disabled', 'true')
      .and('contain.text', 'Requires write permission')

    cy.contains('[role="menuitem"]', 'Edit').click()
    cy.get('[role="dialog"]')
      .should('contain.text', 'Edit Charizard Base Set')
      .and('contain.text', 'Server and field errors remain actionable.')
    cy.contains('button', 'Cancel').click()

    cy.contains('td', 'Charizard Base Set').parent('tr').within(() => {
      cy.get('button[aria-label="Open actions for Charizard Base Set"]').click()
    })
    cy.contains('[role="menuitem"]', 'Permanent delete').click()
    cy.get('[role="alertdialog"]')
      .should('contain.text', 'Permanently delete Charizard Base Set?')
      .and('contain.text', 'This permanently deletes the record.')
      .and('contain.text', 'This cannot be undone.')
  })
})

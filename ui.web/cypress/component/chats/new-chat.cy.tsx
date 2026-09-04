import { mount } from 'cypress/react'
import { NewChat } from '@/features/chats/components/new-chat'

const users = [
  {
    id: 'collector-ops',
    fullName: 'Collector Ops',
    username: 'collector.ops',
    title: 'Collection operations',
    profile: '',
  },
  {
    id: 'inventory-review',
    fullName: 'Inventory Review',
    username: 'inventory.review',
    title: 'Inventory review',
    profile: '',
  },
]

describe('chats NewChat component', () => {
  it('selects and removes recipients while guarding the submit action', () => {
    const onOpenChange = cy.stub().as('onOpenChange')

    mount(<NewChat users={users} open={true} onOpenChange={onOpenChange} />)

    cy.contains('h2', 'New message').should('be.visible')
    cy.contains('button', 'Chat').should('be.disabled')

    cy.contains('[data-slot="command-item"]', 'Collector Ops').click()
    cy.contains('[data-slot="badge"]', 'Collector Ops').should('be.visible')
    cy.contains('button', 'Chat').should('not.be.disabled')

    cy.contains('[data-slot="badge"]', 'Collector Ops')
      .find('button')
      .focus()
      .type('{enter}')
    cy.contains('[data-slot="badge"]', 'Collector Ops').should('not.exist')
    cy.contains('button', 'Chat').should('be.disabled')
  })

  it('resets selected recipients when the dialog closes', () => {
    const onOpenChange = cy.stub().as('onOpenChange')

    mount(<NewChat users={users} open={true} onOpenChange={onOpenChange} />)

    cy.contains('[data-slot="command-item"]', 'Inventory Review').click()
    cy.contains('[data-slot="badge"]', 'Inventory Review').should('be.visible')

    cy.get('[data-slot="dialog-close"]').click()
    cy.get('@onOpenChange').should('have.been.calledWith', false)
  })
})

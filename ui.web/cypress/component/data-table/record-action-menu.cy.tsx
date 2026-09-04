import { mount } from 'cypress/react'
import {
  RecordActionMenu,
  type RecordActionDefinition,
} from '@/components/data-table/record-action-menu'

const actionLog: string[] = []

function actions(): RecordActionDefinition[] {
  return [
    {
      id: 'delete',
      label: 'Delete',
      kind: 'delete',
      onSelect: () => actionLog.push('delete'),
    },
    {
      id: 'view',
      label: 'Open details',
      kind: 'view',
      onSelect: () => actionLog.push('view'),
    },
    {
      id: 'duplicate',
      label: 'Duplicate',
      kind: 'duplicate',
      disabledReason: 'Requires write permission',
      onSelect: () => actionLog.push('duplicate'),
    },
    {
      id: 'restore',
      label: 'Restore',
      kind: 'restore',
      available: false,
      onSelect: () => actionLog.push('restore'),
    },
    {
      id: 'edit',
      label: 'Edit',
      kind: 'edit',
      onSelect: () => actionLog.push('edit'),
    },
  ]
}

describe('RecordActionMenu component', () => {
  beforeEach(() => {
    actionLog.length = 0
  })

  it('orders capability-driven actions and exposes truthful disabled reasons', () => {
    mount(
      <RecordActionMenu
        recordLabel='Charizard Base Set'
        actions={actions()}
      />
    )

    cy.get('button[aria-label="Open actions for Charizard Base Set"]').click()

    cy.get('[role="menuitem"]').then(($items) => {
      expect($items.map((_, item) => item.textContent?.trim()).get()).to.eql([
        'Open details',
        'Edit',
        'DuplicateRequires write permission',
        'Delete',
      ])
    })
    cy.contains('[role="menuitem"]', 'Restore').should('not.exist')
    cy.contains('[role="menuitem"]', 'Duplicate')
      .should('have.attr', 'aria-disabled', 'true')
      .and('contain.text', 'Requires write permission')
  })

  it('uses an accessible icon trigger and isolates row pointer events', () => {
    const rowClick = cy.stub().as('rowClick')
    const rowDoubleClick = cy.stub().as('rowDoubleClick')

    mount(
      <div onClick={rowClick} onDoubleClick={rowDoubleClick}>
        <RecordActionMenu
          recordLabel='Blastoise Base Set'
          actions={actions()}
        />
      </div>
    )

    cy.get('button[aria-label="Open actions for Blastoise Base Set"]')
      .should('have.attr', 'title', 'Open actions for Blastoise Base Set')
      .click()
      .dblclick()

    cy.get('@rowClick').should('not.have.been.called')
    cy.get('@rowDoubleClick').should('not.have.been.called')

    cy.contains('[role="menuitem"]', 'Edit').click()
    cy.wrap(actionLog).should('deep.equal', ['edit'])
    cy.get('@rowClick').should('not.have.been.called')
  })
})

import { mount } from 'cypress/react'
import { DataTableBulkActions } from '@/components/data-table/bulk-actions'
import { type Table } from '@tanstack/react-table'

type Row = {
  id: string
}

function makeTable(selectedCount: number) {
  return {
    getFilteredSelectedRowModel: () => ({
      rows: Array.from({ length: selectedCount }, (_, index) => ({
        id: `row-${index + 1}`,
      })),
    }),
    resetRowSelection: cy.stub().as('resetRowSelection'),
  } as unknown as Table<Row>
}

describe('DataTableBulkActions component', () => {
  it('stays hidden until rows are selected', () => {
    mount(
      <DataTableBulkActions table={makeTable(0)} entityName='item'>
        <button type='button'>Archive</button>
      </DataTableBulkActions>
    )

    cy.get('[role="toolbar"]').should('not.exist')
  })

  it('announces selected rows and clears selection from the toolbar', () => {
    mount(
      <DataTableBulkActions table={makeTable(2)} entityName='item'>
        <button type='button'>Archive</button>
      </DataTableBulkActions>
    )

    cy.get('[role="toolbar"]')
      .should('have.attr', 'aria-label', 'Bulk actions for 2 selected items')
      .and('be.visible')
    cy.contains('[data-slot="badge"]', '2').should('be.visible')
    cy.contains('items').should('be.visible')

    cy.get('button[aria-label="Clear selection"]').click()
    cy.get('@resetRowSelection').should('have.been.calledOnce')
  })

  it('supports keyboard traversal and Escape clear behavior', () => {
    mount(
      <DataTableBulkActions table={makeTable(1)} entityName='item'>
        <button type='button'>Archive</button>
        <button type='button'>Export</button>
      </DataTableBulkActions>
    )

    cy.get('button[aria-label="Clear selection"]').focus()
    cy.focused().type('{rightarrow}')
    cy.focused().should('have.text', 'Archive')
    cy.focused().type('{rightarrow}')
    cy.focused().should('have.text', 'Export')
    cy.focused().type('{esc}')
    cy.get('@resetRowSelection').should('have.been.calledOnce')
  })
})

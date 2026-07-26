import { useRef, useState } from 'react'
import { mount } from 'cypress/react'
import { Button } from '@/components/ui/button'
import { RecordDestructiveConfirmDialog } from '@/components/data-table/record-destructive-confirm-dialog'

function Harness({
  onConfirm,
}: {
  onConfirm: () => Promise<void>
}) {
  const triggerRef = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function confirm() {
    setSubmitting(true)
    await onConfirm()
    setSubmitting(false)
    setOpen(false)
  }

  return (
    <>
      <Button ref={triggerRef} onClick={() => setOpen(true)}>
        Delete record
      </Button>
      <RecordDestructiveConfirmDialog
        open={open}
        onOpenChange={setOpen}
        returnFocusRef={triggerRef}
        action='permanent-delete'
        recordLabel='Charizard Base Set'
        consequence='This removes the inventory record and cannot be undone.'
        isSubmitting={submitting}
        onConfirm={confirm}
        testId='record-destroy-dialog'
      />
    </>
  )
}

describe('RecordDestructiveConfirmDialog component', () => {
  it('names the record, distinguishes destructive consequence, prevents double submit, and returns focus', () => {
    const confirm = cy.stub().as('confirm').callsFake(
      () =>
        new Promise<void>((resolve) => {
          setTimeout(resolve, 150)
        })
    )

    mount(<Harness onConfirm={confirm} />)

    cy.contains('button', 'Delete record').as('trigger').click()
    cy.get('[role="alertdialog"]')
      .should('contain.text', 'Permanently delete Charizard Base Set?')
      .and('contain.text', 'This permanently deletes the record.')
      .and(
        'contain.text',
        'This removes the inventory record and cannot be undone.'
      )
    cy.contains('button', 'Permanently delete').dblclick()
    cy.contains('button', 'Deleting').should('be.disabled')
    cy.get('@confirm').should('have.been.calledOnce')

    cy.get('[role="alertdialog"]').should('not.exist')
    cy.get('@trigger').should('have.focus')
  })

  it('uses truthful copy for archive and soft delete actions', () => {
    mount(
      <RecordDestructiveConfirmDialog
        open
        onOpenChange={() => undefined}
        action='archive'
        recordLabel='Wishlist row'
        consequence='Archived rows can be restored from the table filters.'
        onConfirm={() => undefined}
      />
    )

    cy.get('[role="alertdialog"]')
      .should('contain.text', 'Archive Wishlist row?')
      .and('contain.text', 'This archives the record without permanent removal.')
      .and('contain.text', 'Archived rows can be restored')
    cy.contains('button', 'Archive').should('exist')
  })
})

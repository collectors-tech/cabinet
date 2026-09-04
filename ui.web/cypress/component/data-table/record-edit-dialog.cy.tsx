import { useRef, useState } from 'react'
import { mount } from 'cypress/react'
import { Pencil } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RecordEditDialog } from '@/components/data-table/record-edit-dialog'

function Harness({
  onSubmit,
}: {
  onSubmit: (name: string) => Promise<string | null>
}) {
  const triggerRef = useRef<HTMLButtonElement>(null)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [fieldError, setFieldError] = useState<string | undefined>()
  const [serverError, setServerError] = useState<string | undefined>()
  const [submitting, setSubmitting] = useState(false)

  async function submit() {
    setFieldError(undefined)
    setServerError(undefined)
    if (!name.trim()) {
      setFieldError('Name is required')
      return
    }
    setSubmitting(true)
    const error = await onSubmit(name)
    setSubmitting(false)
    if (error) {
      setServerError(error)
      return
    }
    setOpen(false)
    triggerRef.current?.focus()
  }

  return (
    <>
      <Button ref={triggerRef} onClick={() => setOpen(true)}>
        Edit record
      </Button>
      <RecordEditDialog
        open={open}
        onOpenChange={setOpen}
        returnFocusRef={triggerRef}
        title='Edit Charizard Base Set'
        description='Update the record without leaving the table.'
        icon={Pencil}
        isDirty={name.length > 0}
        isSubmitting={submitting}
        serverError={serverError}
        submitLabel='Save record'
        submittingLabel='Saving record'
        onSubmit={submit}
        testId='record-edit-dialog'
      >
        <div className='space-y-2'>
          <Label htmlFor='record-name'>Name</Label>
          <Input
            id='record-name'
            value={name}
            onChange={(event) => {
              setName(event.target.value)
              setFieldError(undefined)
            }}
            aria-invalid={Boolean(fieldError)}
            aria-describedby={fieldError ? 'record-name-error' : undefined}
            autoFocus
          />
          {fieldError ? (
            <p id='record-name-error' role='alert'>
              {fieldError}
            </p>
          ) : null}
        </div>
      </RecordEditDialog>
    </>
  )
}

describe('RecordEditDialog component', () => {
  it('shows validation/server errors, blocks dirty cancel, prevents double submit, and returns focus', () => {
    const submit = cy
      .stub()
      .as('submit')
      .callsFake(async (name: string) => {
        if (name === 'server') {
          return 'Name is already used'
        }
        return new Promise<string | null>((resolve) => {
          setTimeout(() => resolve(null), 150)
        })
      })

    mount(<Harness onSubmit={submit} />)

    cy.contains('button', 'Edit record').as('trigger').click()
    cy.get('[role="dialog"]')
      .should('contain.text', 'Edit Charizard Base Set')
      .and('contain.text', 'Update the record without leaving the table.')
    cy.get('[data-testid="record-edit-dialog-icon"]').should('exist')
    cy.focused().should('have.attr', 'id', 'record-name')

    cy.contains('button', 'Save record').click()
    cy.get('[role="alert"]').should('contain.text', 'Name is required')
    cy.get('@submit').should('not.have.been.called')

    cy.get('#record-name').clear().type('server')
    cy.contains('button', 'Save record').click()
    cy.get('[role="alert"]').should('contain.text', 'Name is already used')

    cy.get('#record-name').clear().type('updated')
    cy.contains('button', 'Cancel').click()
    cy.get('[data-testid="record-edit-dialog-dirty-warning"]')
      .should('be.visible')
      .and('contain.text', 'Unsaved changes')
    cy.get('[role="dialog"]').should('exist')

    cy.contains('button', 'Save record').dblclick()
    cy.contains('button', 'Saving record').should('be.disabled')
    cy.get('@submit').should('have.been.calledWith', 'updated')
    cy.get('@submit').should('have.callCount', 2)

    cy.get('[role="dialog"]').should('not.exist')
    cy.get('@trigger').should('have.focus')
  })
})

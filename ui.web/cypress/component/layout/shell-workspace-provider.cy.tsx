import { mount } from 'cypress/react'
import {
  ShellWorkspaceProvider,
} from '@/context/shell-workspace-provider'
import {
  type ShellWorkspace,
  useShellWorkspace,
} from '@/context/shell-workspace-context'

function WorkspaceProbe() {
  const {
    activeProfileId,
    activeWorkspace,
    setActiveWorkspace,
    toggleAssistantWorkspace,
  } = useShellWorkspace()

  const workspaces: ShellWorkspace[] = [
    'navigation',
    'search',
    'assistant',
    'inbox',
  ]

  return (
    <section>
      <p data-testid='active-profile'>{activeProfileId}</p>
      <p data-testid='active-workspace'>{activeWorkspace}</p>
      {workspaces.map((workspace) => (
        <button
          key={workspace}
          type='button'
          onClick={() => setActiveWorkspace(workspace)}
        >
          Open {workspace}
        </button>
      ))}
      <button type='button' onClick={toggleAssistantWorkspace}>
        Toggle assistant
      </button>
      <button
        type='button'
        onClick={() => {
          toggleAssistantWorkspace()
          toggleAssistantWorkspace()
        }}
      >
        Toggle assistant twice
      </button>
    </section>
  )
}

describe('ShellWorkspaceProvider', () => {
  beforeEach(() => {
    cy.clearLocalStorage()
  })

  it('loads the active profile and restores the saved workspace for that profile', () => {
    window.localStorage.setItem(
      'cabinet.shell.workspace.active.profile-alpha',
      'inbox'
    )

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-alpha' },
    }).as('activeProfile')

    mount(
      <ShellWorkspaceProvider>
        <WorkspaceProbe />
      </ShellWorkspaceProvider>
    )

    cy.wait('@activeProfile')
    cy.get('[data-testid="active-profile"]').should('have.text', 'profile-alpha')
    cy.get('[data-testid="active-workspace"]').should('have.text', 'inbox')
  })

  it('persists explicit workspace selections and supports assistant toggling', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-beta' },
    }).as('activeProfile')

    mount(
      <ShellWorkspaceProvider>
        <WorkspaceProbe />
      </ShellWorkspaceProvider>
    )

    cy.wait('@activeProfile')
    cy.contains('button', 'Open assistant').click()
    cy.get('[data-testid="active-workspace"]').should('have.text', 'assistant')
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.shell.workspace.active.profile-beta')
      .should('eq', 'assistant')

    cy.contains('button', 'Toggle assistant').click()
    cy.get('[data-testid="active-workspace"]').should('have.text', 'navigation')

    cy.contains('button', 'Open search').click()
    cy.get('[data-testid="active-workspace"]').should('have.text', 'search')
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.shell.workspace.active.profile-beta')
      .should('eq', 'search')
  })

  it('applies each assistant toggle against the latest workspace state', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-gamma' },
    }).as('activeProfile')

    mount(
      <ShellWorkspaceProvider>
        <WorkspaceProbe />
      </ShellWorkspaceProvider>
    )

    cy.wait('@activeProfile')
    cy.contains('button', 'Toggle assistant twice').click()
    cy.get('[data-testid="active-workspace"]').should('have.text', 'navigation')
    cy.window()
      .its('localStorage')
      .invoke('getItem', 'cabinet.shell.workspace.active.profile-gamma')
      .should('eq', 'navigation')
  })
})

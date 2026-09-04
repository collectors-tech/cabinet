import { mount } from 'cypress/react'
import { ContentSection } from '@/features/settings/components/content-section'

describe('settings ContentSection', () => {
  it('renders the section heading, description, and slotted controls', () => {
    mount(
      <ContentSection
        title='Display'
        desc='Choose how Cabinet presents dense workspace data.'
      >
        <button type='button'>Save display preferences</button>
      </ContentSection>
    )

    cy.contains('h3', 'Display').should('be.visible')
    cy.contains('Choose how Cabinet presents dense workspace data.').should(
      'be.visible'
    )
    cy.contains('button', 'Save display preferences').should('be.visible')
  })
})

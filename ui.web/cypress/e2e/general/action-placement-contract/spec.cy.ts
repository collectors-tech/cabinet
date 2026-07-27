import {
  actionPlacementRegionOrder,
  buildActionPlacementChecklist,
  getActionPlacementRegion,
  isPageActionRegion,
  shellUtilityActionIds,
} from '../../../../src/lib/action-placement'

describe('action placement contract', () => {
  it('UI-ACTION-PLACEMENT-001 defines canonical action regions and ordering', () => {
    expect(actionPlacementRegionOrder).to.deep.eq([
      'page-header',
      'table-toolbar',
      'bulk-toolbar',
      'row-menu',
      'dialog-footer',
      'shell-utility',
    ])

    expect(getActionPlacementRegion('page-header')).to.include({
      id: 'page-header',
      label: 'Page header',
      scope: 'whole-page',
    })
    expect(getActionPlacementRegion('row-menu')).to.include({
      id: 'row-menu',
      label: 'Row action menu',
      scope: 'single-record',
    })
    expect(getActionPlacementRegion('missing-region')).to.eq(undefined)
  })

  it('UI-ACTION-PLACEMENT-002 classifies page, toolbar, row, dialog, and shell actions', () => {
    const checklist = buildActionPlacementChecklist([
      {
        id: 'reports-refresh',
        label: 'Refresh',
        placement: 'page-header',
        kind: 'refresh',
      },
      {
        id: 'inventory-search',
        label: 'Search inventory',
        placement: 'table-toolbar',
        kind: 'filter',
      },
      {
        id: 'wishlist-bulk-delete',
        label: 'Delete selected',
        placement: 'bulk-toolbar',
        kind: 'bulk',
      },
      {
        id: 'media-archive',
        label: 'Archive asset',
        placement: 'row-menu',
        kind: 'record',
      },
      {
        id: 'confirm-delete',
        label: 'Delete',
        placement: 'dialog-footer',
        kind: 'dialog-confirm',
      },
      {
        id: 'theme-switch',
        label: 'Theme',
        placement: 'shell-utility',
        kind: 'shell',
      },
    ])

    expect(checklist.pageActions.map((action) => action.id)).to.deep.eq([
      'reports-refresh',
    ])
    expect(checklist.toolbarActions.map((action) => action.id)).to.deep.eq([
      'inventory-search',
    ])
    expect(checklist.bulkActions.map((action) => action.id)).to.deep.eq([
      'wishlist-bulk-delete',
    ])
    expect(checklist.rowActions.map((action) => action.id)).to.deep.eq([
      'media-archive',
    ])
    expect(checklist.dialogFooterActions.map((action) => action.id)).to.deep.eq([
      'confirm-delete',
    ])
    expect(checklist.shellUtilityActions.map((action) => action.id)).to.deep.eq([
      'theme-switch',
    ])
  })

  it('UI-ACTION-PLACEMENT-003 keeps shell utilities out of page action regions', () => {
    expect(shellUtilityActionIds).to.include.members([
      'language-switch',
      'theme-switch',
      'configuration',
      'profile-menu',
      'sidebar-toggle',
    ])

    shellUtilityActionIds.forEach((id) => {
      expect(
        isPageActionRegion({
          id,
          label: id,
          placement: 'shell-utility',
          kind: 'shell',
        }),
        id
      ).to.eq(false)
    })

    expect(
      isPageActionRegion({
        id: 'market-watch-new-query',
        label: 'New query',
        placement: 'page-header',
        kind: 'create',
      })
    ).to.eq(true)
  })
})

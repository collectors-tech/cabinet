import {
  actionPlacementRegionOrder,
  buildActionPlacementChecklist,
  buildPageHeaderActionLayout,
  buildShellActionBoundary,
  getActionPlacementAuditSummary,
  getRouteActionRegionContract,
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

  it('UI-ACTION-PLACEMENT-004 keeps primary header actions visible and moves secondary actions into accessible overflow', () => {
    const actions = [
      {
        id: 'market-watch-new-query',
        label: 'New query',
        placement: 'page-header',
        kind: 'create',
        priority: 'primary',
      },
      {
        id: 'market-watch-run',
        label: 'Run search',
        placement: 'page-header',
        kind: 'run',
        priority: 'secondary',
      },
      {
        id: 'market-watch-export',
        label: 'Export',
        placement: 'page-header',
        kind: 'export',
        priority: 'secondary',
      },
      {
        id: 'theme-switch',
        label: 'Theme',
        placement: 'shell-utility',
        kind: 'shell',
      },
    ] as const

    const narrowLayout = buildPageHeaderActionLayout(actions, {
      viewport: 'narrow',
      maxVisibleActions: 1,
    })

    expect(narrowLayout.visibleActions.map((action) => action.id)).to.deep.eq([
      'market-watch-new-query',
    ])
    expect(narrowLayout.overflowActions.map((action) => action.id)).to.deep.eq([
      'market-watch-run',
      'market-watch-export',
    ])
    expect(narrowLayout.shellUtilityActions.map((action) => action.id)).to.deep
      .eq(['theme-switch'])
    expect(narrowLayout.overflowMenu).to.deep.include({
      ariaLabel: 'More page actions',
      testId: 'page-header-action-overflow',
      keyboardReachable: true,
    })

    const wideLayout = buildPageHeaderActionLayout(actions, {
      viewport: 'wide',
      maxVisibleActions: 3,
    })

    expect(wideLayout.visibleActions.map((action) => action.id)).to.deep.eq([
      'market-watch-new-query',
      'market-watch-run',
      'market-watch-export',
    ])
    expect(wideLayout.overflowActions).to.deep.eq([])
    expect(wideLayout.overflowMenu).to.eq(undefined)
  })

  it('UI-ACTION-PLACEMENT-005 documents the shell utility and page action boundary', () => {
    const boundary = buildShellActionBoundary([
      {
        id: 'profile-menu',
        label: 'Profile',
        placement: 'shell-utility',
        kind: 'shell',
      },
      {
        id: 'reports-refresh',
        label: 'Refresh',
        placement: 'page-header',
        kind: 'refresh',
        priority: 'primary',
      },
      {
        id: 'inventory-search',
        label: 'Search inventory',
        placement: 'table-toolbar',
        kind: 'filter',
      },
    ])

    expect(boundary.shellUtilityActionIds).to.deep.eq(['profile-menu'])
    expect(boundary.pageActionIds).to.deep.eq(['reports-refresh'])
    expect(boundary.nonPageActionIds).to.deep.eq(['inventory-search'])
    expect(boundary.shellUtilityPlacement).to.deep.include({
      placement: 'shell-utility',
      includedInPageOverflow: false,
      ownsGlobalChrome: true,
    })
    expect(boundary.pageActionPlacement).to.deep.include({
      placement: 'page-header',
      includedInPageOverflow: true,
      ownsGlobalChrome: false,
    })
  })

  it('UI-ACTION-PLACEMENT-006 publishes route-level page action regions for packaged shell checks', () => {
    expect(getRouteActionRegionContract('/reports')).to.deep.include({
      route: '/reports',
      pageActionRegionTestId: 'reports-global-header-actions',
    })
    expect(
      getRouteActionRegionContract('/reports')?.wholePageActionIds
    ).to.deep.eq(['reports-refresh', 'reports-export'])

    expect(getRouteActionRegionContract('/scanner')).to.deep.include({
      route: '/scanner',
      pageActionRegionTestId: 'market-watch-global-header-actions',
    })
    expect(
      getRouteActionRegionContract('/scanner')?.wholePageActionIds
    ).to.deep.eq(['market-watch-create', 'market-watch-run'])

    expect(getRouteActionRegionContract('/help-center')).to.deep.include({
      route: '/help-center',
    })
    expect(getRouteActionRegionContract('/help-center')).not.to.have.property(
      'pageActionRegionTestId'
    )
    expect(
      getRouteActionRegionContract('/help-center')?.wholePageActionIds
    ).to.deep.eq([])
  })

  it('UI-ACTION-PLACEMENT-007 audits authenticated route action coverage without duplicate page actions', () => {
    const audit = getActionPlacementAuditSummary()

    expect(audit.routeCount).to.eq(25)
    expect(audit.routesMissingContract).to.deep.eq([])
    expect(audit.duplicateWholePageActionIds).to.deep.eq([])
    expect(audit.routesWithPageActions).to.deep.eq([
      '/inventory',
      '/collections',
      '/wishlist',
      '/media',
      '/purchases',
      '/reports',
      '/scanner',
      '/users',
    ])
    expect(audit.routesWithoutPageActions).to.include.members([
      '/dashboard',
      '/integrations',
      '/settings/display',
      '/help-center',
    ])
    expect(audit.pageActionRegionOrder).to.deep.eq([
      'page-header',
      'table-toolbar',
      'bulk-toolbar',
      'row-menu',
      'dialog-footer',
      'shell-utility',
    ])
    expect(audit.wholePageActionLabels).to.deep.include({
      id: 'reports-refresh',
      label: 'Refresh',
      route: '/reports',
    })
    expect(audit.wholePageActionLabels).to.deep.include({
      id: 'market-watch-create',
      label: 'Create Saved Watch',
      route: '/scanner',
    })
  })
})

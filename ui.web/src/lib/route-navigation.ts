import {
  authenticatedRouteMetadata,
  getRouteMetadata,
  type RouteMetadata,
} from './route-metadata'

export type SearchNavigationResult = {
  id: string
  title: string
  group: string
  path: string
  value: string
}

export type RouteHeaderTitleProps = Pick<
  RouteMetadata,
  'title' | 'description' | 'icon'
>

export type SidebarNavigationItem = {
  title: string
  url: string
  icon: RouteMetadata['icon']
  testIdKey: string
}

export type SidebarNavigationGroup = {
  title: 'General' | 'Other'
  testIdKey: 'general' | 'other'
  items: (
    | SidebarNavigationItem
    | {
        title: 'Settings'
        testIdKey: 'settings'
        icon: RouteMetadata['icon']
        items: SidebarNavigationItem[]
      }
  )[]
}

function navigationPathFor(metadata: RouteMetadata) {
  if (metadata.path === '/') {
    return metadata.aliases?.[0] ?? metadata.path
  }

  if (metadata.path === '/settings') {
    return metadata.aliases?.[0] ?? metadata.path
  }

  return metadata.path
}

function navResultKey(group: string, title: string, path: string) {
  return `${group}:${title}:${path}`.toLowerCase()
}

function toSearchNavigationResult(
  metadata: RouteMetadata
): SearchNavigationResult {
  const path = navigationPathFor(metadata)
  const group = metadata.navigationGroup
  const value = `${metadata.title} ${group} ${path}`

  return {
    id: navResultKey(group, metadata.title, path),
    title: metadata.title,
    group,
    path,
    value,
  }
}

function sidebarTestKeyFor(metadata: RouteMetadata) {
  const sidebarLink = metadata.testIds.sidebarLink
  if (!sidebarLink) {
    return metadata.title.trim().toLowerCase().replace(/\s+/g, '-')
  }

  return sidebarLink.replace(/^sidebar-nav-link-/, '')
}

function toSidebarNavigationItem(
  metadata: RouteMetadata
): SidebarNavigationItem {
  return {
    title: metadata.title,
    url: navigationPathFor(metadata),
    icon: metadata.icon,
    testIdKey: sidebarTestKeyFor(metadata),
  }
}

export function buildSearchNavigationResults() {
  return authenticatedRouteMetadata
    .filter((metadata) => metadata.navigationGroup !== 'System')
    .map(toSearchNavigationResult)
}

export function buildSidebarNavigationGroups(): SidebarNavigationGroup[] {
  const generalItems = authenticatedRouteMetadata
    .filter((metadata) => metadata.navigationGroup === 'General')
    .map(toSidebarNavigationItem)
  const settingsItems = authenticatedRouteMetadata
    .filter((metadata) => metadata.navigationGroup === 'Settings')
    .map(toSidebarNavigationItem)
  const otherItems = authenticatedRouteMetadata
    .filter((metadata) => metadata.navigationGroup === 'Other')
    .map(toSidebarNavigationItem)
  const settingsIcon =
    authenticatedRouteMetadata.find(
      (metadata) => metadata.path === '/settings/operations'
    )?.icon ?? settingsItems[0]?.icon

  return [
    {
      title: 'General',
      testIdKey: 'general',
      items: generalItems,
    },
    {
      title: 'Other',
      testIdKey: 'other',
      items: [
        ...(settingsIcon
          ? [
              {
                title: 'Settings',
                testIdKey: 'settings',
                icon: settingsIcon,
                items: settingsItems,
              } as const,
            ]
          : []),
        ...otherItems,
      ],
    },
  ]
}

export function getRouteHeaderTitleProps(pathname: string) {
  const metadata = getRouteMetadata(pathname)

  if (!metadata) {
    return undefined
  }

  return {
    title: metadata.title,
    description: metadata.description,
    icon: metadata.icon,
  } satisfies RouteHeaderTitleProps
}

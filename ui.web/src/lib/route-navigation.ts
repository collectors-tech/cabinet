import { authenticatedRouteMetadata, type RouteMetadata } from './route-metadata'

export type SearchNavigationResult = {
  id: string
  title: string
  group: string
  path: string
  value: string
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

export function buildSearchNavigationResults() {
  return authenticatedRouteMetadata
    .filter((metadata) => metadata.navigationGroup !== 'System')
    .map(toSearchNavigationResult)
}

## 1. Contract

- [x] 1.1 Define the canonical authenticated route metadata contract for
  #1940.
- [x] 1.2 Publish the complete authenticated route metadata matrix before
  implementation.

## 2. Shared registry

- [ ] 2.1 Add a typed route metadata registry with canonical path or pattern,
  title, description, icon, navigation group, document-title eligibility, and
  stable test IDs.
- [ ] 2.2 Add document-title resolution from the registry, including
  `/purchases` and the `/scanner` -> `Market Watch` correction.
- [ ] 2.3 Connect sidebar/search navigation and `HeaderTitle` consumers to the
  shared metadata where practical without changing navigation architecture.
- [ ] 2.4 Add Settings child route metadata with specific titles/icons while
  retaining Settings grouping.

## 3. Evidence

- [ ] 3.1 Add table-driven tests that cover every authenticated route in the
  registry and fail when a route lacks canonical metadata.
- [ ] 3.2 Add focused UI coverage for visible headers, icons, document titles,
  and responsive non-overlap on representative desktop and narrow widths.
- [ ] 3.3 Run focused route metadata tests, UI build, strict OpenSpec
  validation, and `git diff --check`.

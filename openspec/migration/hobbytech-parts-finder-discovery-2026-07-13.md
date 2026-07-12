# Hobbytech Parts Finder Discovery

Issue: #1499
Probe timestamp: 2026-07-13 04:13 Australia/Sydney / 2026-07-12 18:13 UTC
Probe log: `.work-agent/logs/issue-1499-hobbytech-parts-finder/20260713-0413/hobbytech-live-probe.txt`

## Findings

- `https://hobbytechtoys.com.au/robots.txt` returned HTTP 200 and identifies the site as Shopify-backed.
- Robots rules disallow `/cart`, `/checkout`, `/checkouts/`, `/orders`, `/account`, `/search`, and multiple filtered query patterns including `/*?pf_*`.
- `https://hobbytechtoys.com.au/pages/parts-finder` returned HTTP 200 with title metadata `Parts Finder`.
- `https://hobbytechtoys.com.au/pages/part-finder` returned 404.
- `https://hobbytechtoys.com.au/search?q=parts+finder` returned HTTP 200, but `/search` is robots-disallowed and should not be used as a default discovery path.

## Cabinet Boundary

Hobbytech remains a no-auth catalogue/source-matching provider. Cabinet may expose the Parts Finder as a read-only, preview-only workflow for public parts catalogue review and manual product URL handoff. It must not automate customer login, cart, checkout, payment, purchase, or account/order flows.

## Follow-Up

The next implementation slice can add parser fixtures only after a smaller probe identifies stable public Parts Finder page structure for brand/model/parts relationships. If the public page does not expose stable parseable structure, Cabinet should keep the workflow as manual URL capture plus Boost/mybcapps search fallback.

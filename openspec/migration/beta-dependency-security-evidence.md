# Cabinet beta dependency security evidence

Issue: #2051

Candidate parent: #1864 / #1868

## Red baseline

On clean `develop` lineage at `b89971d80633fa4f699ffce6a2fd4d77f672fa71`, `npm audit --omit=dev --json` reproduced 10 vulnerable production package families: 1 critical and 6 high, plus 2 moderate and 1 low. The critical/high paths included `seroval`, `axios`, `vite`, `postcss`, `form-data`, `nanoid`, and `picomatch`.

The first `node --test scripts/dependency-security.test.mjs` run failed because the governed checker did not exist. After the checker contract existed, `npm run security:dependencies` failed closed on the same 1 critical and 6 high findings before lockfile remediation.

## Patched graph

The lockfile was regenerated within the declared compatible dependency ranges, with patched direct minimums recorded for `axios`, `i18next-http-backend`, and `vite`. The resulting production graph includes:

- `axios` 1.19.0;
- `vite` 7.3.6;
- `postcss` 8.5.26;
- `seroval` 1.6.2;
- `form-data` 4.0.6;
- `i18next-http-backend` 3.0.6;
- `nanoid` 3.3.18 and 5.1.16; and
- `picomatch` 4.0.5 on the affected Vite/tinyglobby paths.

After a fresh `npm ci`, the repository gate and direct npm audit report zero critical or high production vulnerabilities. No vulnerability exception or manual advisory dismissal is used.

The final full and production audits retain one low-severity `esbuild` advisory affecting the Windows development server. Cabinet beta artifacts package static Vite output and serve it through the embedded Go runtime, so that development-server path is not part of the shipped runtime. The gate reports lower-severity findings but, as required by #2051, fails on critical or high findings; there is no critical/high exception.

GitHub Dependabot alerts are generated from the default-branch dependency graph and therefore cannot reconcile against this unmerged worktree. After the change reaches the candidate branch, release review must verify that the critical/high alerts close against the new lockfile rather than dismissing them without evidence.

## Clean-install drift classification

The clean Windows install removes the stale Clerk/cookie tree. Current npm versions report five packages as extraneous even though each is an `optional: true` lock entry owned by the `@tailwindcss/oxide-wasm32-wasi` lock-owned optional WebAssembly bundle:

- `@emnapi/core`;
- `@emnapi/runtime`;
- `@emnapi/wasi-threads`;
- `@napi-rs/wasm-runtime`; and
- `@tybys/wasm-util`.

The gate permits only this structural class: an extraneous report must resolve to a package path present in the exact lockfile with `optional: true`. Any undeclared, missing, invalid, or non-optional extraneous package remains a hard failure. This distinguishes npm's bundled optional-package reporting from stale product dependency drift.

## Enforcement

`npm run security:dependencies` validates the installed top-level tree and parses live `npm audit --omit=dev --json` output. It fails when audit metadata is absent or when critical/high counts are non-zero. Contract tests bind the gate to:

- Develop Quality Gate;
- Main Gate;
- Beta Release Candidate Gate; and
- Release Portable Packages.

The gate does not publish a release and does not promote `develop` to `main`.

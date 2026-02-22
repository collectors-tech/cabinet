# API Docs

Source of truth:
- `docs/api/openapi.yaml`

Human-readable docs:
- `docs/api/redoc.html` (loads `openapi.yaml` via Redoc CDN)

## Local Preview
Serve the `docs/api` folder (example PowerShell):

```powershell
cd docs/api
python -m http.server 8090
```

Open:
- `http://127.0.0.1:8090/redoc.html`

## Validate Spec

```powershell
./scripts/validate-openapi.ps1
```

## Generate Static HTML

```powershell
./scripts/generate-api-docs.ps1
```

This generates:
- `docs/api/index.html`

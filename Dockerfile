# syntax=docker/dockerfile:1.7

FROM node:22-bookworm AS ui-build
WORKDIR /src

COPY ui.web/package.json ui.web/package-lock.json ./ui.web/
WORKDIR /src/ui.web
RUN npm ci

COPY docs/help-center/ /src/docs/help-center/
COPY ui.web/ ./
RUN npm run build

FROM golang:1.24-bookworm AS app-build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=ui-build /src/internal/ui/static ./internal/ui/static

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/cabinet ./cmd/cabinet

FROM debian:bookworm-slim AS runtime

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --create-home --home-dir /home/cabinet --shell /usr/sbin/nologin cabinet \
  && mkdir -p /data \
  && chown cabinet:cabinet /data

WORKDIR /app
COPY --from=app-build /out/cabinet /app/cabinet
COPY docs/api/openapi.yaml /app/docs/api/openapi.yaml

ENV CABINET_OPEN_BROWSER=0
EXPOSE 17880
VOLUME ["/data"]
HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=12 CMD curl -fsS http://127.0.0.1:17880/healthz || exit 1

USER cabinet
ENTRYPOINT ["/app/cabinet"]
CMD ["--no-open-browser", "--listen", "0.0.0.0:17880", "--data-dir", "/data", "--profile", "e2e-cypress", "--instance-name", "cypress-container", "--allow-parallel"]

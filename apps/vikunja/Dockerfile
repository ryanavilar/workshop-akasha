# Workshop image — built by the self-hosted runner and pushed to Zot
FROM node:22-alpine AS frontend

WORKDIR /build

ENV PNPM_CACHE_FOLDER=.cache/pnpm/
ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV CYPRESS_INSTALL_BINARY=0

COPY frontend/pnpm-lock.yaml frontend/package.json frontend/.npmrc ./
RUN npm install -g corepack && corepack enable && \
    pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

FROM golang:1.24-alpine AS backend

RUN apk add --no-cache git make
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY --from=frontend /build/dist ./frontend/dist

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o vikunja ./

FROM alpine:3.21

RUN apk add --no-cache tzdata ca-certificates
RUN mkdir -p /tmp && chmod 1777 /tmp

WORKDIR /app/vikunja

COPY --from=backend /build/vikunja ./vikunja

USER 1000

ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/

EXPOSE 3456

ENTRYPOINT ["/app/vikunja/vikunja"]

# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:26.1.0-alpine AS skia-deps
ARG TARGETARCH
RUN apk add --no-cache wget
RUN if [[ "${TARGETARCH}" = 'arm64' ]]; then wget https://github.com/Brooooooklyn/canvas/releases/download/v1.0.0/skia.linux-arm64-musl.node && sha256sum skia.linux-arm64-musl.node | cut -d' ' -f1 | grep c21775d41575caf7153f0b17e6be3b631add4a86f0aff5ad9a812a8bd008820d >/dev/null; elif [[ "${TARGETARCH}" = 'amd64' ]]; then wget https://github.com/Brooooooklyn/canvas/releases/download/v1.0.0/skia.linux-x64-musl.node && sha256sum skia.linux-x64-musl.node | cut -d' ' -f1 | grep 51d5dd5095107a9c6bd63f9fe70ec8420a5470dda1dc38a7ac15dcab283b3808 >/dev/null; else false; fi

FROM --platform=$BUILDPLATFORM node:26.1.0-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY chartmaker/tsconfig.json chartmaker/package.json chartmaker/
RUN --mount=type=cache,target=/root/.local/share/pnpm/store npm i -g corepack@0.34.7 && corepack enable && corepack prepare --activate && pnpm i --filter chartmaker --filter maple-culvert-tracker --frozen-lockfile
COPY chartmaker/src chartmaker/src/
RUN pnpm run --filter chartmaker build


FROM node:26.1.0-alpine AS runner
RUN apk add --no-cache font-roboto
WORKDIR /chartmaker/

COPY --from=skia-deps /skia.linux-*-musl.node ./
COPY --from=builder /app/chartmaker/dist/chartmaker.js ./chartmaker.js
USER node

EXPOSE 3000

ENV PORT="3000"
CMD ["./chartmaker.js"]
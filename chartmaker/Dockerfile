# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM oven/bun:1-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY package.json bun.lock bunfig.toml tsconfig.json ./
COPY chartmaker/tsconfig.json chartmaker/package.json chartmaker/
COPY chartmaker/src chartmaker/src/
COPY culvert-web/package.json culvert-web/
RUN bun install --filter maple-culvert-tracker --filter chartmaker --frozen-lockfile
RUN bun --filter chartmaker build --target=bun-${TARGETOS}-${TARGETARCH/amd/x}-musl

FROM oven/bun:1-alpine AS runner
RUN apk add --no-cache font-roboto
WORKDIR /chartmaker/

COPY --from=builder /app/chartmaker/dist/chartmaker ./chartmaker
USER bun

EXPOSE 3000

ENV PORT="3000"

CMD ["./chartmaker"]
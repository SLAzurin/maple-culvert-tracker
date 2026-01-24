FROM oven/bun:1-alpine AS base

FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker/tsconfig.json chartmaker/package.json chartmaker/
COPY chartmaker/src chartmaker/src/
COPY culvert-web/package.json culvert-web/
RUN bun install && \
    cd chartmaker && \
    bun run build && \
    rm -rf /app/{package.json,pnpm-lock.yaml,pnpm-workspace.yaml,tsconfig.json} /app/chartmaker/{tsconfig.json,package.json,node_modules,src}

FROM oven/bun:1-alpine AS runner
RUN apk add --no-cache font-roboto
WORKDIR /chartmaker/

RUN addgroup --system --gid 1001 ovenbun
RUN adduser --system --uid 1001 bunuser

COPY --from=builder /app/chartmaker/dist/chartmaker ./chartmaker
USER bunuser

EXPOSE 3000

ENV PORT="3000"

CMD ["./chartmaker"]
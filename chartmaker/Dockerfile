FROM node:20-slim AS base
ARG BUILDARCH
# TODO: Either include this code or fork the project, replace with @napi-rs/canvas to make it work on arm64
RUN echo "$BUILDARCH" | grep -v "amd64" && apt-get update && \
    apt-get install -y build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev && \
    rm -rf /var/lib/apt/lists/* || true

FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker/tsconfig.json chartmaker/package.json chartmaker/
COPY chartmaker/src chartmaker/src/
RUN corepack enable && corepack prepare --activate && pnpm i && cd chartmaker && pnpm run build && rm -rf /app/node_modules /app/chartmaker/node_modules $(pnpm store path)


FROM base AS prod-deps
WORKDIR /app/chartmaker
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
COPY chartmaker/package.json /app/chartmaker/
RUN cd /app && corepack enable && corepack prepare --activate && pnpm i -P && rm -rf /app/{package.json,pnpm-lock.yaml,pnpm-workspace.yaml} /app/chartmaker/package.json


FROM base as runner
WORKDIR /chartmaker/

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/chartmaker/dist ./dist/
COPY --from=prod-deps /app/chartmaker/node_modules ./node_modules/
COPY --from=prod-deps /app/node_modules /node_modules/

RUN apt-get update && apt-get install -y -q libfontconfig1 && rm -rf /var/lib/apt/lists/*

USER nodeuser

EXPOSE 3000

ENV PORT 3000

CMD ["node", "dist/server.js"]
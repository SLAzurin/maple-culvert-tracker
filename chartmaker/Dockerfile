FROM node:18-slim AS base
RUN npm i -g pnpm
ARG BUILDARCH
# TODO: Either include this code or fork the project, replace with @napi-rs/canvas to make it work on arm64
RUN echo "$BUILDARCH" | grep -v "amd64" && apt-get update && \
    apt-get install -y build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev && \
    rm -rf /var/lib/apt/lists/* || true


FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker chartmaker/
RUN pnpm i -w chartmaker && cd chartmaker && pnpm run build && rm -rf /app/node_modules /app/chartmaker/node_modules $(pnpm store path)


FROM base AS runner
WORKDIR /app/chartmaker
ENV NODE_ENV production

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
COPY chartmaker/package.json /app/chartmaker/
RUN cd /app && pnpm i -w chartmaker

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/chartmaker/dist ./dist

RUN apt-get update && apt-get install -y -q libfontconfig1 && rm -rf /var/lib/apt/lists/*

USER nodeuser

EXPOSE 3000

ENV PORT 3000

CMD ["node", "dist/server.js"]
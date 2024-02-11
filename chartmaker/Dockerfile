FROM node:20-slim AS base
ARG BUILDARCH
# TODO: Either include this code or fork the project, replace with @napi-rs/canvas to make it work on arm64
RUN echo "$BUILDARCH" | grep -v "amd64" && apt-get update && \
    apt-get install -y build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev && \
    rm -rf /var/lib/apt/lists/* || true


FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker/package.json chartmaker/sea-*.json chartmaker/tsconfig.json chartmaker/
RUN corepack enable && pnpm i --ignore-scripts
COPY chartmaker/src chartmaker/src
RUN cd chartmaker && \
    pnpm run build

FROM base AS runner
WORKDIR /app/chartmaker
ENV NODE_ENV production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

RUN apt-get update && apt-get install -y -q libfontconfig1 && rm -rf /var/lib/apt/lists/*

RUN npm i -g chart.js@^3.9.1 sharp@^0.33.2 canvas@^2.11.2

USER nodeuser

EXPOSE 3000

ENV PORT 3000

WORKDIR /home/nodeuser

COPY --from=builder --chown=1001:1001 /app/chartmaker/dist/server /usr/local/bin/
ENV NODE_PATH /usr/local/lib/node_modules
CMD ["/usr/local/bin/server"]
FROM node:22-slim AS base

FROM base AS builder
ENV NODE_ENV dev
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY culvert-web/package.json culvert-web/
RUN corepack enable && corepack prepare --activate && pnpm i
COPY culvert-web/index.html culvert-web/tsconfig.json culvert-web/tsconfig.node.json culvert-web/vite.config.ts culvert-web/
COPY culvert-web/src culvert-web/src/
RUN corepack enable && corepack prepare --activate && cd culvert-web && pnpm run build


FROM nginx:stable-alpine-slim AS runner
RUN rm -rf /usr/share/nginx/html && mkdir -p /usr/share/nginx/html
COPY --from=builder /app/culvert-web/build /usr/share/nginx/html
COPY nginx/static.conf.d/default.conf /etc/nginx/conf.d/default.conf

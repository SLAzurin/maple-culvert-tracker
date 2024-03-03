FROM node:20-slim AS base

FROM base AS builder
ENV NODE_ENV dev
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY culvert-web/package.json culvert-web/
RUN corepack enable && pnpm install
COPY culvert-web culvert-web/
RUN corepack enable && cd culvert-web && pnpm run build


FROM base AS runner
WORKDIR /app/culvert-web
ENV NODE_ENV production

RUN npm i -g serve

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/culvert-web/build ./build

USER nodeuser

EXPOSE 8080

ENV PORT 8080

CMD ["serve", "-s", "build", "-l", "8080"]
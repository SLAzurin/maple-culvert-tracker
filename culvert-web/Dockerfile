FROM node:20-slim AS base
RUN npm i -g pnpm

FROM base AS builder
ENV NODE_ENV dev
WORKDIR /app
COPY pnpm-lock.yaml pnpm-workspace.yaml ./
COPY culvert-web/package.json culvert-web/
RUN cd culvert-web && pnpm install
COPY culvert-web culvert-web/
RUN cd culvert-web && pnpm run build


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
FROM node:18-slim AS base
RUN npm i -g pnpm


FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker chartmaker/
RUN pnpm i -w chartmaker
RUN cd chartmaker && pnpm run build


FROM base AS runner
WORKDIR /app/chartmaker
ENV NODE_ENV production

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
COPY chartmaker/package.json /app/chartmaker/
RUN cd /app && pnpm i -w chartmaker

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/chartmaker/dist ./dist

USER nodeuser

EXPOSE 3000

ENV PORT 3000

CMD ["node", "dist/server.js"]
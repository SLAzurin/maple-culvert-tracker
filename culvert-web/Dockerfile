FROM node:18-slim AS base
# TODO: EDIT THIS ITS NOT FIXED
RUN npm i -g pnpm

FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY culvert-web/package.json culvert-web/
RUN pnpm i -w culvert-web
COPY culvert-web culvert-web/
RUN cd culvert-web && pnpm run build


FROM base AS runner
WORKDIR /app/culvert-web
ENV NODE_ENV production

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
COPY culvert-web/package.json /app/culvert-web/
RUN cd /app && pnpm i -w culvert-web

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/culvert-web/dist ./dist

USER nodeuser

EXPOSE 3000

ENV PORT 3000

CMD ["node", "dist/server.js"]
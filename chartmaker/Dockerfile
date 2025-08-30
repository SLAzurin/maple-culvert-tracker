FROM node:24-alpine AS base
RUN apk add --no-cache 'libcrypto3>=3.3.2-r1'

FROM base AS skia-deps
ARG TARGETARCH
RUN apk add --no-cache wget
RUN if [[ "${TARGETARCH}" = 'arm64' ]]; then wget https://github.com/Brooooooklyn/canvas/releases/download/v0.1.78/skia.linux-arm64-musl.node && sha256sum skia.linux-arm64-musl.node | cut -d' ' -f1 | grep f22687e529e34c2b641f8e1fc8d61d756d4cf0ba1c5e0a3d67acb9bafee77c3e >/dev/null; elif [[ "${TARGETARCH}" = 'amd64' ]]; then wget https://github.com/Brooooooklyn/canvas/releases/download/v0.1.78/skia.linux-x64-musl.node && sha256sum skia.linux-x64-musl.node | cut -d' ' -f1 | grep 3eb5527e72e4cbed7fecb70b220321a3ae83924c8e91bd9306c2d71a4dfa68da >/dev/null; else false; fi

FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.json ./
COPY chartmaker/tsconfig.json chartmaker/package.json chartmaker/chartmakerBuilder.cjs chartmaker/
COPY chartmaker/src chartmaker/src/
RUN corepack enable && \
    corepack prepare --activate && \
    pnpm i && \
    cd chartmaker && \
    pnpm run build && \
    rm -rf /app/{package.json,pnpm-lock.yaml,pnpm-workspace.yaml,tsconfig.json} /app/chartmaker/{tsconfig.json,package.json,chartmakerBuilder.cjs,node_modules,src} $(pnpm store path)

FROM base AS prod-deps
WORKDIR /app/chartmaker
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
COPY chartmaker/package.json /app/chartmaker/
RUN cd /app && \
    corepack enable && \
    corepack prepare --activate && \
    pnpm i -P && \
    cp -rL ./chartmaker/node_modules ./chartmaker/node_modules_prod && \
    rm -rf /app/{package.json,pnpm-lock.yaml,pnpm-workspace.yaml} /app/chartmaker/{package.json,node_modules} $(pnpm store path)


FROM base AS runner
RUN apk add --no-cache font-roboto
WORKDIR /chartmaker/

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nodeuser

COPY --from=builder /app/chartmaker/dist ./dist/
COPY --from=skia-deps /skia.linux-*-musl.node ./node_modules/@napi-rs/canvas/
COPY --from=prod-deps /app/chartmaker/node_modules_prod ./node_modules/
USER nodeuser

EXPOSE 3000

ENV PORT="3000"

CMD ["node", "dist/server.js"]
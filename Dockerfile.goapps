# syntax=docker/dockerfile:1
FROM golang:1.25.0-alpine AS build
RUN apk --no-cache add tzdata

WORKDIR /src/

COPY go.mod go.sum /src/

RUN go mod download

COPY internal /src/internal
COPY .gen /src/.gen

ARG APPNAME=main
COPY cmd/${APPNAME} /src/cmd/${APPNAME}
RUN cd /src/cmd/${APPNAME} && CGO_ENABLED=0 GOOS=linux go build -o /src/main .

# culvert-web specific
FROM node:24-slim AS culvert-web-builder
ENV NODE_ENV dev
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY culvert-web/package.json culvert-web/
RUN corepack enable && corepack prepare --activate && pnpm i
COPY culvert-web/index.html culvert-web/tsconfig.json culvert-web/tsconfig.node.json culvert-web/vite.config.ts culvert-web/
COPY culvert-web/src culvert-web/src/
RUN corepack enable && corepack prepare --activate && cd culvert-web && pnpm run build

FROM scratch AS final

COPY --from=build /src/main /

WORKDIR /
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Vancouver
ENV GIN_MODE=release
USER nobody

ENTRYPOINT [ "/main" ]

FROM final AS culvert-web-final
COPY --from=culvert-web-builder /app/culvert-web/build /www/

ENTRYPOINT [ "/main" ]

FROM build AS cron-specific-build
# main arg already built auto_backup
COPY cmd/culvert_score_update_reminder /src/cmd/culvert_score_update_reminder
RUN cd /src/cmd/culvert_score_update_reminder && CGO_ENABLED=0 GOOS=linux go build -o /src/culvert_score_update_reminder .

FROM postgres:16-alpine AS cron
COPY --from=build /src/main /usr/local/bin/auto_backup
COPY --from=cron-specific-build /src/culvert_score_update_reminder /usr/local/bin/
COPY ./cron.d/root /etc/cron/

FROM final AS final-with-server-assets
COPY ./server_assets /server_assets/
COPY ./db_migrations /db_migrations/

ENTRYPOINT [ "/main" ]
#!/bin/bash
set -e
which node

JET_BIN="$(which jet 2>/dev/null || true)"
if [[ "$JET_BIN" = '' ]]; then
    JET_BIN="$HOME/go/bin/jet"
fi
eval "$(cat .env)"
ENCODED_PW="$(node -e "console.log(encodeURI(\"$POSTGRES_PASSWORD\"))")"
"$JET_BIN" -dsn="postgres://$POSTGRES_USER:$ENCODED_PW@$POSTGRES_HOST:$CLIENT_POSTGRES_PORT/$POSTGRES_DB?sslmode=disable" -path=./.gen
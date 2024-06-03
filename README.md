# maple-culvert-tracker

Track culvert score over a x amount of time progress

This software is best used in conjunction with my https://github.com/SLAzurin/gpq-image-ocr-gui

# Pre-requisites
Install Docker Engine (Linux) or Docker Desktop (Linux/Windows/Mac) for development.  
Install Golang 1.22.x or newer versions when developing.  
Install Nodejs 20.x or newer even number LTS versions. (Using [nvm.sh](https://github.com/nvm-sh/nvm) is recommended)  
Enable `pnpm` with this command:  
    - Command: `corepack enable && corepack prepare --activate`


# developer notes
1. Setup the discord bot and their permissions and make it join your test server.
2. Setup the `.env` file according to `.env.template`.
3. Run the `chartmaker`, `db`, `redis` containers.
    - Command: `docker compose -f base.yml -f dev.yml up -d chartmaker db redis`
4. Run the `update_commands` Go app once only.
    - Command: `go run cmd/update_commands/*.go`
5. Run the `main` Go app (discord bot) process and leave it running in the background.
    - Command: `go run cmd/main/*.go`
6. Install Nodejs dependencies with `pnpm`:
    - Command: `pnpm i`
7. Run the Website control panel and leave it in the background.
    - Command: `cd culvert-web ; pnpm run dev`

# production deployment
1. Setup the discord bot and their permissions and make it join your server.
2. Setup the `.env` file according to `.env.template`.
3. Build and run the update_commands entrypoint once
    - Command: `go build -o update_commands ./cmd/update_commands/*.go `
    - Copy it to the production server, next to the docker-compose.yml file
    - Run: `./update_commands`
4. Use docker compose, and run the following command:
    - Command: `docker compose up -d`


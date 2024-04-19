# maple-culvert-tracker

Track culvert score over a x year(s) progress


Missing features:
- search discord username/id (done)
- edit previous week's culvert scores (done)
- name changes (done)

Optional additional features
- list member's characters

# Pre-requisites
Install Docker Engine (Linux) or Docker Desktop (Linux/Windows/Mac)  
Install Golang 1.22.x or newer versions.  
Install Nodejs 20.x or newer even number versions. (Using nvmsh is recommended)  
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
1. Use docker engine, and run the following command:
    - Command: `docker compose up -d`
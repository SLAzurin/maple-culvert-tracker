# Maple Culvert Tracker

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/G2G0WUCP2)

This software helps track Maplestory culvert scores over time!

This software is best used in conjunction with my https://github.com/SLAzurin/gpq-image-ocr-gui

Join the Discord server for update notifications! Here: Server under construction.

A lot of work still needs to be done to be considered a well established Open Source Software.  
You should consider this codebase as Public Source instead.  

Anyone is welcome to contribute by opening new issues and/or open PRs.  
Please open issues to ask questions and open PRs for any code changes.  

## This Culvert Tracker is used by these lovely guilds:  
- Straw (Kronos)
- Bloom (Kronos)
- Beary (Kronos)
- Thrones (Kronos)

# Pre-requisites

Install Docker Engine (Linux) or Docker Desktop (Linux/Windows/Mac) for development.  
Install Golang 1.25.x or newer versions when developing.  
Install Nodejs 24.x or newer even number LTS versions. (Using [nvm.sh](https://github.com/nvm-sh/nvm) is recommended)  
Enable `pnpm` with this command:

- Command: `corepack enable && corepack prepare --activate`

## Deployment note

Although discouraged, hosting this at home is possible.  
Hosting this on the Cloud is recommended to keep stable uptime.

I suggest the following providers for competitive pricing:

- [Hetzner](https://www.hetzner.com/cloud/)
- [OVH Cloud](https://www.ovhcloud.com/en/vps/)

Requirements are relatively low spec. I'd suggest a minimum of 1cpu, 1gb of ram and 40gb of storage.

# developer setup notes

1. Setup the discord bot and their permissions and make it join your test server.
2. Setup the `.env` file according to `.env.template`.
3. Run the `chartmaker`, `db16`, `valkey` containers.
   - Command: `docker compose -f base.yml -f dev.yml up -d chartmaker db16 valkey`
   - Connect inside the db16 container: `docker compose exec db16 sh`
   - Run the sql files: `psql -U $POSTGRES_USER -d $POSTGRES_DB </root/sqlfiles/createdb.sql`
4. Run the `main` Go app (discord bot) process and leave it running in the background.
   - Command: `go run cmd/main/*.go`
5. Install Nodejs dependencies with `pnpm`:
   - Command: `pnpm i`
6. Run the Website control panel and leave it in the background.
   - Command: `cd culvert-web ; pnpm run dev`

# Developer codebase notes
Internally used discord id `1` means unlinked and not in guild anymore  
Internally used discord id `2` means unlinked but still in guild  

I never delete characters in case they ever return to the guild. This has happened before and their culvert scores will be kept.  

This culvert bot project started on April 11th 2023! :D

# production deployment

1. Setup the discord bot and their permissions and make it join your server.
2. Setup the `.env` file according to `.env.template`.
3. Use docker compose, and run the following command:
   - Command: `docker compose up -d`
   - Connect inside the db16 container: `docker compose exec db16 sh`
   - Run the sql files: `psql -U $POSTGRES_USER -d $POSTGRES_DB </root/sqlfiles/createdb.sql`

# backing up the postgres db

1. Connect to the db container with a shell and run a pg_dump
   - Run: `docker compose exec db16 sh -c "pg_dump -U \$POSTGRES_USER -d \$POSTGRES_DB >/root/sqlfiles/dump.sql"`
   - The `dump.sql` is the database dump file (inside the `./sqlfiles/` path)
   - Backup that file "somewhere".

# restoring a db backup

1. Copy the dump inside the container then connect into it and run the sql file.
   - Copy the `dump.sql` inside the `./sqlfiles/` path.
   - Run: `docker compose exec db16 sh`
   - Run: `psql -U $POSTGRES_USER -d postgres`
   - Drop and re-create $POSTGRES_DB: `drop database mapleculverttrackerdb; create database mapleculverttrackerdb;` then `exit` the db connection
   - Run the sql backup: `psql -U $POSTGRES_USER -d $POSTGRES_DB </root/sqlfiles/dump.sql`
   - You are done restoring the backup.



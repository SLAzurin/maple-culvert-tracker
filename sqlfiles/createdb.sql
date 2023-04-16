CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS discord_servers (
  id SERIAL PRIMARY KEY,
  guild_name varchar(255) not null,
  discord_server_native_id varchar(255) not null unique
);

-- many characters allowed per discord user
CREATE TABLE IF NOT EXISTS guild_characters (
  id SERIAL PRIMARY KEY,
  maple_character_name varchar(255) not null unique,
  discord_server_id int not null,
  discord_user_id varchar(255) not null,
  FOREIGN KEY (discord_server_id) references discord_servers(id) on delete cascade
);
CREATE INDEX "guild_characters_discord_user_id_key" on guild_characters(discord_user_id);

CREATE TABLE IF NOT EXISTS character_culvert_scores (
    culvert_date DATE not null,
    maple_character_name varchar(255) not null,
    PRIMARY KEY (maple_character_name, culvert_date)
);

-- auth: One time login with command on discord
CREATE TABLE IF NOT EXISTS otp_codes (
  id UUID  PRIMARY KEY DEFAULT uuid_generate_v4(),
  discord_server_id varchar(255) not null,
  discord_user_id varchar(255) not null,
  expires timestamptz DEFAULT (now() + interval '2 hours')
);
CREATE INDEX "otp_codes_discord_user_id_key" on otp_codes(discord_user_id);
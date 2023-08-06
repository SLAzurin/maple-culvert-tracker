-- many characters allowed per discord user
CREATE TABLE IF NOT EXISTS characters (
  id BIGSERIAL PRIMARY KEY,
  maple_character_name varchar(255) not null unique,
  discord_user_id varchar(255) not null
);
CREATE TABLE IF NOT EXISTS character_culvert_scores (
  id BIGSERIAL PRIMARY KEY,
  culvert_date DATE not null,
  character_id bigint not null,
  score int not null,
  FOREIGN KEY (character_id) references characters(id) on delete cascade
);
ALTER TABLE character_culvert_scores
  ADD CONSTRAINT character_culvert_scores_culvert_date_character_id UNIQUE (culvert_date, character_id);

-- Must have at least 1 character in the db for the frontend to work properly loles
INSERT INTO characters (maple_character_name, discord_user_id) VALUES ('iMonoxian', '82981249026621440');
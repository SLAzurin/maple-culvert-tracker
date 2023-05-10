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
CREATE INDEX character_culvert_scores_culvert_date ON character_culvert_scores (culvert_date);

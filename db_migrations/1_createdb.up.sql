CREATE TABLE IF NOT EXISTS characters (
  id BIGSERIAL PRIMARY KEY,
  maple_character_name VARCHAR(255) NOT NULL UNIQUE,
  discord_user_id VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS character_culvert_scores (
  id BIGSERIAL PRIMARY KEY,
  culvert_date DATE NOT NULL,
  character_id BIGINT NOT NULL,
  score INT NOT NULL,
  FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);
ALTER TABLE character_culvert_scores DROP CONSTRAINT IF EXISTS character_culvert_scores_culvert_date_character_id;
ALTER TABLE character_culvert_scores
  ADD CONSTRAINT character_culvert_scores_culvert_date_character_id UNIQUE (culvert_date, character_id);

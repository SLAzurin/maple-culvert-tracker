BEGIN;

CREATE TABLE secondary_characters (
  character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
  maple_character_name VARCHAR(255) NOT NULL CHECK (btrim(maple_character_name) <> '')
);
CREATE UNIQUE INDEX secondary_characters_name_key ON secondary_characters (lower(maple_character_name));
CREATE INDEX secondary_characters_character_id_idx ON secondary_characters (character_id);

COMMIT;

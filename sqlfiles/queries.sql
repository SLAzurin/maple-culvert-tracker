-- Get culvert score for a <discord_user> from a <discord_server>
SELECT character_culvert_scores.culvert_date, character_culvert_scores.score
FROM discord_servers
INNER JOIN guild_characters ON guild_characters.discord_server_id = discord_servers.id
INNER JOIN character_culvert_scores ON character_culvert_scores.maple_character_name = guild_characters.maple_character_name
WHERE discord_servers.discord_server_native_id = '' AND guild_characters.discord_user_id = ''
ORDER BY character_culvert_scores.culvert_date
LIMIT 52;


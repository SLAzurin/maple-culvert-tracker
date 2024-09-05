package apiredis

const (
	CONF_DISCORD_ADMIN_CHANNEL_ID        redisInternalKey = "CONF_DISCORD_ADMIN_CHANNEL_ID"
	CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID redisInternalKey = "CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID"
	CONF_DISCORD_GUILD_ROLE_IDS          redisInternalKey = "CONF_DISCORD_GUILD_ROLE_IDS"

	OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX    redisInternalKey = "OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX"
	OPTIONAL_CONF_MAPLE_REGION               redisInternalKey = "OPTIONAL_CONF_MAPLE_REGION" // default "na"
	OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL redisInternalKey = "OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL"

	// Internal keys below this line
	DATA_REDIS_VERSION redisInternalKey = "DATA_REDIS_VERSION"
	DATA_DB_VERSION    redisInternalKey = "DATA_DB_VERSION"

	DATA_FIXES_SUN_TO_WED redisInternalKey = "DATA_FIXES_SUN_TO_WED"
)

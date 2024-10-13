package apiredis

var (
	CONF_DISCORD_ADMIN_CHANNEL_ID                         = redisInternalKey{"CONF_DISCORD_ADMIN_CHANNEL_ID", editableTypeDiscordChannel}
	CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID redisInternalKey = redisInternalKey{"CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID", editableTypeDiscordChannel}
	CONF_DISCORD_GUILD_ROLE_IDS          redisInternalKey = redisInternalKey{"CONF_DISCORD_GUILD_ROLE_IDS", editableTypeDiscordRole}

	OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX    redisInternalKey = redisInternalKey{"OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX", editableTypeString}
	OPTIONAL_CONF_MAPLE_REGION               redisInternalKey = redisInternalKey{"OPTIONAL_CONF_MAPLE_REGION", editableTypeSelection} // default "na"
	OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL redisInternalKey = redisInternalKey{"OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL", editableTypeString}

	// Internal keys below this line
	DATA_REDIS_VERSION redisInternalKey = redisInternalKey{"DATA_REDIS_VERSION", editableTypeNone}
	DATA_DB_VERSION    redisInternalKey = redisInternalKey{"DATA_DB_VERSION", editableTypeNone}

	DATA_FIXES_SUN_TO_WED redisInternalKey = redisInternalKey{"DATA_FIXES_SUN_TO_WED", editableTypeNone}

	DATA_DISCORD_MEMBERS redisInternalKey = redisInternalKey{"DATA_DISCORD_MEMBERS", editableTypeNone}
)

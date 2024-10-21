package apiredis

var (
	CONF_DISCORD_ADMIN_CHANNEL_ID                         = redisInternalKey{"CONF_DISCORD_ADMIN_CHANNEL_ID", EditableTypeDiscordChannel, false}
	CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID redisInternalKey = redisInternalKey{"CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID", EditableTypeDiscordChannel, false}
	CONF_DISCORD_GUILD_ROLE_IDS          redisInternalKey = redisInternalKey{"CONF_DISCORD_GUILD_ROLE_IDS", EditableTypeDiscordRole, true}

	OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX    redisInternalKey = redisInternalKey{"OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX", EditableTypeString, false}
	OPTIONAL_CONF_MAPLE_REGION               redisInternalKey = redisInternalKey{"OPTIONAL_CONF_MAPLE_REGION", EditableTypeSelection, false} // default "na"
	OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL redisInternalKey = redisInternalKey{"OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL", EditableTypeString, false}

	// Internal keys below this line
	DATA_REDIS_VERSION redisInternalKey = redisInternalKey{"DATA_REDIS_VERSION", EditableTypeNone, false}
	DATA_DB_VERSION    redisInternalKey = redisInternalKey{"DATA_DB_VERSION", EditableTypeNone, false}

	DATA_FIXES_SUN_TO_WED redisInternalKey = redisInternalKey{"DATA_FIXES_SUN_TO_WED", EditableTypeNone, false}

	DATA_DISCORD_MEMBERS redisInternalKey = redisInternalKey{"DATA_DISCORD_MEMBERS", EditableTypeNone, false}

	DATA_DISCORD_NEW_FEATURES_ANNOUNCEMENT_VERSION redisInternalKey = redisInternalKey{"DATA_DISCORD_NEW_FEATURES_ANNOUNCEMENT_VERSION", EditableTypeNone, false}
)

var KeysMap = map[string]redisInternalKey{}

var EditableSelectionsMap = map[redisInternalKey][]string{
	OPTIONAL_CONF_MAPLE_REGION: {"", "na", "eu"},
}

func init() {
	KeysMap[CONF_DISCORD_ADMIN_CHANNEL_ID.Name] = CONF_DISCORD_ADMIN_CHANNEL_ID
	KeysMap[CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.Name] = CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID
	KeysMap[CONF_DISCORD_GUILD_ROLE_IDS.Name] = CONF_DISCORD_GUILD_ROLE_IDS
	KeysMap[OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.Name] = OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX
	KeysMap[OPTIONAL_CONF_MAPLE_REGION.Name] = OPTIONAL_CONF_MAPLE_REGION
	KeysMap[OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.Name] = OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL
	KeysMap[DATA_REDIS_VERSION.Name] = DATA_REDIS_VERSION
	KeysMap[DATA_DB_VERSION.Name] = DATA_DB_VERSION
	KeysMap[DATA_FIXES_SUN_TO_WED.Name] = DATA_FIXES_SUN_TO_WED
	KeysMap[DATA_DISCORD_MEMBERS.Name] = DATA_DISCORD_MEMBERS
}

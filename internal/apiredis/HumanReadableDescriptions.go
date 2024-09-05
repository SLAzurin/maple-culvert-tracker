package apiredis

type HumanReadableDescriptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var confReadableNameAndDescriptions = map[redisInternalKey]HumanReadableDescriptions{
	CONF_DISCORD_ADMIN_CHANNEL_ID:            {"Discord Admin Channel ID", "The ID of the channel where the bot will send admin notifications"},
	CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID:     {"Discord Members Main Channel ID", "The ID of the channel where the bot will interact with guild members"},
	CONF_DISCORD_GUILD_ROLE_IDS:              {"Discord Guild Role IDs", "The IDs of guild members. Comma separated"},
	OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX:    {"Discord Reminder Suffix", "Optional fluff suffix on the daily reminder message"},
	OPTIONAL_CONF_MAPLE_REGION:               {"Maple Region", "The region of the server. Must be 'na' or 'eu'"},
	OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL: {"Culvert Duel Thumbnail URL", "The URL of the thumbnail of the duel image"},
}

func GetHumanReadableDescriptions(k redisInternalKey) *HumanReadableDescriptions {
	if v, ok := confReadableNameAndDescriptions[k]; ok {
		return &v
	}
	return nil
}

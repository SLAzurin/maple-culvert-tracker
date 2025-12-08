package apiredis

type HumanReadableDescriptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var confReadableNameAndDescriptions = map[redisInternalKey]HumanReadableDescriptions{
	CONF_DISCORD_ADMIN_CHANNEL_ID:                {"Discord Admin Channel ID", "REQUIRED The ID of the channel where the bot will send admin notifications"},
	CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID:         {"Discord Members Main Channel ID", "REQUIRED The ID of the channel where the bot will interact with guild members"},
	CONF_DISCORD_GUILD_ROLE_IDS:                  {"Discord Guild Role IDs", "REQUIRED The IDs of guild members. Comma separated"},
	OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX:        {"Optional Discord Reminder Suffix", "Optional fluff suffix on the daily reminder message"},
	OPTIONAL_CONF_MAPLE_REGION:                   {"Optional Maple Region", "Optional region of the server. Must be 'na' or 'eu', empty defaults to 'na'"},
	OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL:     {"Optional Culvert Duel Thumbnail URL", "The URL of the thumbnail of the duel image"},
	OPTIONAL_CONF_SUBMIT_SCORES_SHOW_SANDBAGGERS: {"Optional Show sandbaggers after submitting scores", "Toggle yes/no if the bot should show sandbaggers after submitting weekly scores"},
	OPTIONAL_CONF_SUBMIT_SCORES_SHOW_RATS:        {"Optional Show rats after submitting scores", "Toggle 6/7 if the bot should show rats after submitting weekly scores"},
	OPTIONAL_CONF_SANDBAG_THRESHOLD:              {"Optional sandbagger threshold", "Threshold which defines a sandbagger, against their previous best culvert score. (Default or empty: 0.7)"},
}

func GetHumanReadableDescriptions(k redisInternalKey) *HumanReadableDescriptions {
	if v, ok := confReadableNameAndDescriptions[k]; ok {
		return &v
	}
	return nil
}

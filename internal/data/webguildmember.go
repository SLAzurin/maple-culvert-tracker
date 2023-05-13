package data

type WebGuildMember struct {
	DiscordUsername string `json:"discord_username"`
	DiscordUserID   string `json:"discord_user_id"`
	DiscordNickname string `json:"discord_nickname,omitempty"`
}

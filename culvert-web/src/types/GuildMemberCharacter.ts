export interface GuildMemberCharacter {
	[key: string]: {
		// key: character name
		discord_user_id: string;
		previousWeek: number;
		currentWeek?: number;
	};
}

import FetchResult from "../types/FetchResult";

const linkDiscordMaple = async (
	auth: string,
	discordUserID: string,
	characterName: string,
	link: boolean,
	bypass_name_check: boolean = false,
): Promise<FetchResult> => {
	const res = await fetch("/api/maple/link", {
		method: "POST",
		headers: {
			Authorization: `Bearer ${auth}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			discord_user_id: discordUserID,
			character_name: characterName,
			link,
			bypass_name_check,
		}),
	});

	if (res.status !== 200) {
		return res
			.text()
			.then((text) => Promise.resolve({ status: res.status, payload: text }));
	}

	return { status: res.status, payload: res.json() };
};

export default linkDiscordMaple;

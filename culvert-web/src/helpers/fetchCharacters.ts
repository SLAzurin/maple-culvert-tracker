const fetchCharacters = async (
	auth: string,
): Promise<
	| { character_name: string; character_id: number; discord_user_id: string }[]
	| number
> => {
	try {
		const res = await fetch("/api/maple/characters/fetch", {
			headers: {
				Authorization: `Bearer ${auth}`,
			},
		});
		if (res.status !== 200) {
			return Promise.resolve(res.status);
		}
		return await res.json();
	} catch (e) {
		return Promise.resolve(-1);
	}
};

export default fetchCharacters;

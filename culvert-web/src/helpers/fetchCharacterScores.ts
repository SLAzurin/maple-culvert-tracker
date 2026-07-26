const fetchCharacterScores = async (
	week?: string,
): Promise<
	| {
			weeks: string[];
			data: { character_id: number; culvert_date: string; score: number }[];
	  }
	| number
> => {
	try {
		const res = await fetch(
			`/api/maple/characters/culvert?week=${week || ""}`,
		);
		if (res.status !== 200) {
			return Promise.resolve(res.status);
		}
		return await res.json();
	} catch (e) {
		return Promise.resolve(-1);
	}
};

export default fetchCharacterScores;

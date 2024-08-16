import FetchResult from "../types/FetchResult";

const updateCulvertScores = async (
	auth: string,
	data: {
		payload: { character_id: number; score: number }[];
		isNew: boolean;
		week: string;
	},
): Promise<FetchResult> => {
	const res = await fetch("/api/maple/characters/culvert", {
		method: "POST",
		headers: {
			Authorization: `Bearer ${auth}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify(data),
	});

	if (res.status !== 200) {
		return res
			.text()
			.then((text) => Promise.resolve({ status: res.status, payload: text }));
	}

	return { status: res.status, payload: res.json() };
};

export default updateCulvertScores;

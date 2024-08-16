import FetchResult from "../types/FetchResult";

const renameCharacter = async (
	auth: string,
	data: {
		character_id: number;
		new_name: string;
		bypass_name_check: boolean;
	},
): Promise<FetchResult> => {
	const res = await fetch("/api/maple/characters/rename", {
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

export default renameCharacter;

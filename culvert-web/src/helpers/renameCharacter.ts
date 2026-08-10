import { type FetchResult } from "../types/FetchResult";

const renameCharacter = async (data: {
	character_id: number;
	new_name: string;
	bypass_name_check: boolean;
}): Promise<FetchResult> => {
	const res = await fetch("/api/maple/characters/rename", {
		method: "POST",
		headers: {
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

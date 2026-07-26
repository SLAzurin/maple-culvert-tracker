import { type MCTClaims } from "../types/MCTClaims";

const loginWithToken = async (
	token: string,
): Promise<MCTClaims | number> => {
	try {
		const res = await fetch("/api/auth/login", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ token }),
		});
		if (res.status !== 200) {
			return res.status;
		}
		return await res.json();
	} catch {
		return -1;
	}
};

export default loginWithToken;

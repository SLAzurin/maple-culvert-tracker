import { type MCTClaims } from "../types/MCTClaims";

const fetchMe = async (): Promise<MCTClaims | number> => {
	try {
		const res = await fetch("/api/auth/me");
		if (res.status !== 200) {
			return res.status;
		}
		return await res.json();
	} catch {
		return -1;
	}
};

export default fetchMe;

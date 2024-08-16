import MCTClaims from "../types/MCTClaims";

const validateJWT = (auth: string): { valid: boolean; claims?: MCTClaims } => {
	if (!auth.includes(".")) return { valid: false };
	const parts: string[] = auth.split(".");
	if (parts.length !== 3) return { valid: false };
	const claims: MCTClaims = JSON.parse(window.atob(auth.split(".")[1]));
	if (Number(claims.exp) * 1000 <= new Date().getTime()) {
		return { valid: false };
	}
	return { valid: true, claims };
};

export default validateJWT;

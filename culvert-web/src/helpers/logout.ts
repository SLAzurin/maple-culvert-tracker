const logout = async (): Promise<void> => {
	try {
		await fetch("/api/auth/logout", { method: "POST" });
	} catch {
		// ignore network errors; local state is still cleared by caller
	}
};

export default logout;

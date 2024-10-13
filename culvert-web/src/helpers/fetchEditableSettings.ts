const fetchEditableSettings = async (auth: string): Promise<any> => {
	try {
		const res = await fetch("/api/editable-settings", {
			headers: {
				Authorization: `Bearer ${auth}`,
			},
		});
		if (res.status !== 200) {
			return Promise.reject(res.status);
		}
		return await res.json();
	} catch (e) {
		return Promise.resolve(-1);
	}
};

export default fetchEditableSettings;

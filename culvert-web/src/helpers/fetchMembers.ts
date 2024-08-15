import GuildMember from "../types/GuildMember";

const fetchMembers = async (auth: string): Promise<GuildMember[] | number> => {
	try {
		const res = await fetch("/api/discord/members/fetch", {
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

export default fetchMembers;

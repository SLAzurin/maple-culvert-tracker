export const getHumanValue = (editableValues: any, key: string) => {
	switch (editableValues[key].editable_type) {
		case "discord_channel":
			if (editableValues[key].multiple) {
				// const cs = editableValues[key].value.split(",");
				// const cn: string[] = [];
				// cs.forEach((channelID: string) => {
				// 	const channelName = (
				// 		editableValues[key].available_channels as any[]
				// 	).find(
				// 		(cdata: any) => cdata.id === channelID && cdata.type === 0,
				// 	)?.name;
				// 	cn.push(channelName || channelID);
				// });
				// return cn.join(", ");
				//
				//
				//
				//
				// This shouldn't be possible yet
			} else
				return (
					editableValues[key].available_channels.find(
						(channelData: any) =>
							channelData.id === editableValues[key].value &&
							channelData.type === 0,
					)?.name || editableValues[key].value
				);
		case "discord_role":
			if (editableValues[key].multiple) {
				const rs = editableValues[key].value.split(",");
				const rn: string[] = [];
				rs.forEach((r: string) => {
					rn.push(
						editableValues[key].available_roles.find(
							(roleData: any) => roleData.id === r,
						)?.name || r,
					);
				});
				return rn.join(", ");
			} else
				return (
					editableValues[key].available_roles.find(
						(r: any) => r.id === editableValues[key].value,
					)?.name || editableValues[key].value
				);
		case "string":
			return editableValues[key].value;
		case "selection":
			return (
				editableValues[key].available_selections.find(
					(s: string) => s === editableValues[key].value,
				) ?? "BAD VALUE"
			);
	}
};

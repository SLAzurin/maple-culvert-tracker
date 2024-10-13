export const getHumanValue = (editableValues: any, key: string) => {
	switch (editableValues[key].editable_type) {
		case "discord_channel":
			if (editableValues[key].multiple) {
				const cs = editableValues[key].value.split(",");
				const cn: string[] = [];
				cs.forEach((c: string) => {
					cn.push(
						editableValues[key].available_channels.filter(
							(c: any) => c.id === c && c.type === 0,
						).name || c,
					);
				});
				return cn.join(", ");
			} else
				return (
					editableValues[key].available_channels.filter(
						(c: any) => c.id === editableValues[key].value && c.type === 0,
					).name || editableValues[key].value
				);
		case "discord_role":
			if (editableValues[key].multiple) {
				const rs = editableValues[key].value.split(",");
				const rn: string[] = [];
				rs.forEach((r: string) => {
					rn.push(
						editableValues[key].available_roles.filter((r: any) => r.id === r)
							.name || r,
					);
				});
				return rn.join(", ");
			} else
				return (
					editableValues[key].available_roles.filter(
						(r: any) => r.id === editableValues[key].value,
					).name || editableValues[key].value
				);
		case "string":
			return editableValues[key].value;
		case "selection":
			return editableValues[key].available_selections.filter(
				(s: string) => s === editableValues[key].value,
			);
	}
};

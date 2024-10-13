import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectToken } from "../features/login/loginSlice";
import fetchEditableSettings from "../helpers/fetchEditableSettings";
import { getHumanValue } from "./EditSettingsHelpers";

const EditSettings = () => {
	const navigate = useNavigate();
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [editableValues, setEditableValues] = useState({} as any);
	const [newValuesMap, setNewValuesMap] = useState({} as any);

	useEffect(() => {
		try {
			fetchEditableSettings(token).then((res) => {
				setEditableValues(res);
			});
		} catch {
			alert("Failed to fetch editable settings");
		}
	}, []);

	return (
		<div>
			<button className="btn btn-secondary" onClick={() => navigate(-1)}>
				Return to homepage
			</button>
			<br />
			<br />

			<form>
				{Object.keys(editableValues).map((key) => (
					<div key={"all-inputs-" + key}>
						<br />
						<h4>{editableValues[key]?.human_readable_description?.name}</h4>
						<p>
							{editableValues[key]?.human_readable_description?.description}
						</p>
						<p>
							<span style={{ color: "blue" }}>Current value: </span>
							{getHumanValue(editableValues, key)}
						</p>
						{(() => {
							switch (editableValues[key].editable_type) {
								case "string":
									return (
										<div>
											<span style={{ color: "red" }}>New value: </span>
											<input
												type="text"
												name={key}
												onChange={(e) => {
													setNewValuesMap({
														...newValuesMap,
														[key]: e.target.value,
													});
												}}
												value={newValuesMap[key] || editableValues[key].value}
											/>
										</div>
									);
								case "selection":
									return (
										<div>
											<select
												onChange={(e) => {
													setNewValuesMap({
														...newValuesMap,
														[key]: e.target.value,
													});
												}}
												defaultValue={editableValues[key].value}
											>
												{editableValues[key].available_selections.map(
													(s: string, i: number) => (
														<option
															key={i + "-available_selections_" + key}
															value={s}
														>
															{s}
														</option>
													),
												)}
											</select>
										</div>
									);
								case "discord_channel":
									if (editableValues[key].multiple) {
										// Not implemented cuz it is not yet possible
									} else {
										return (
											<div>
												<p>Use one of the following channels:</p>
												<select
													defaultValue={editableValues[key].value}
													onChange={(e) => {
														setNewValuesMap({
															...newValuesMap,
															[key]: e.target.value,
														});
													}}
												>
													{editableValues[key].available_channels
														.filter((c: any) => c.type === 0)
														.map((s: any, i: number) => (
															<option
																key={i + "_available_channels_" + key}
																value={s.id}
															>
																{s.name}
															</option>
														))}
												</select>
												<p style={{ color: "red" }}>
													New value:{" "}
													{editableValues[key].available_channels.find(
														(c: any) => c.id === newValuesMap[key],
													)?.name ??
														newValuesMap[key] ??
														editableValues[key].available_channels.find(
															(c: any) => c.id === editableValues[key].value,
														)?.name}
												</p>
											</div>
										);
									}
								case "discord_role":
									if (editableValues[key].multiple) {
										return (
											<div>
												<select
													value={""}
													onChange={(e) => {
														let roles: string[] = [];
														if (newValuesMap[key]) {
															roles = newValuesMap[key].split(",");
														}
														const r = roles.find(
															(r: string) => r === e.target.value,
														);
														if (r) {
															return;
														}
														roles.push(e.target.value);
														setNewValuesMap({
															...newValuesMap,
															[key]: roles.join(","),
														});
													}}
												>
													<option value="">Select role(s)</option>
													{editableValues[key].available_roles
														.filter((r: any) => r.name !== "@everyone")
														.map((s: any, i: number) => (
															<option
																key={i + "_available_roles_" + key}
																value={s.id}
															>
																{s.name}
															</option>
														))}
												</select>
												<p style={{ color: "red" }}>New values: </p>
												{newValuesMap[key] &&
													newValuesMap[key]
														.split(",")
														.map((s: string, i: number) => (
															<button
																key={i + "_available_roles_" + key}
																onClick={(e) => {
																	e.preventDefault();
																	const roles = newValuesMap[key].split(",");
																	roles.splice(i, 1);
																	setNewValuesMap({
																		...newValuesMap,
																		[key]: roles.join(","),
																	});
																}}
																className="btn btn-success"
															>
																{
																	editableValues[key].available_roles.find(
																		(r: any) => r.id === s,
																	)?.name
																}{" "}
																❌
															</button>
														))}
											</div>
										);
									} else {
										// Not implemented cuz it is not yet possible
									}
							}
						})()}
					</div>
				))}
			</form>
		</div>
	);
};

export default EditSettings;

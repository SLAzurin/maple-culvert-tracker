import { Login } from "./features/login/Login";
import "./App.css";
import { useEffect, useState } from "react";
import fetchMembers from "./helpers/fetchMembers";
import {
	resetToken,
	selectClaims,
	selectToken,
} from "./features/login/loginSlice";
import { selectMembers, setMembers } from "./features/members/membersSlice";
import { store } from "./app/store";
import { useSelector } from "react-redux";
import Container from "react-bootstrap/Container";
import Nav from "react-bootstrap/Nav";
import Navbar from "react-bootstrap/Navbar";
import {
	addNewCharacterScore,
	applyCulvertChanges,
	resetCharacterScores,
	selectCharacterScores,
	selectCharacters,
	selectMembersCharacters,
	selectEditableWeeks,
	selectSelectedWeek,
	selectUpdateCulvertScoresResult,
	setCharacterScores,
	setCharacters,
	setSelectedWeek,
	updateScoreValue,
} from "./features/characters/charactersSlice";
import fetchCharacters from "./helpers/fetchCharacters";
import fetchCharacterScores from "./helpers/fetchCharacterScores";
import { selectMembersByID } from "./features/members/membersSlice";
import renameCharacter from "./helpers/renameCharacter";
import { useNavigate } from "react-router-dom";
import linkDiscordMaple from "./helpers/linkDiscordMaple";
import GuildMember from "./types/GuildMember";
interface ImportedData {
	[key: string]: number;
}

function App() {
	const navigate = useNavigate();
	const token = useSelector(selectToken);
	const claims = useSelector(selectClaims);
	const members = useSelector(selectMembers);
	const membersByID = useSelector(selectMembersByID);
	const characters = useSelector(selectCharacters);
	const characterScores = useSelector(selectCharacterScores);
	const updateCulvertScoresResult = useSelector(
		selectUpdateCulvertScoresResult,
	);
	const membersCharacters = useSelector(selectMembersCharacters);
	const editableWeeks = useSelector(selectEditableWeeks);
	const selectedWeek = useSelector(selectSelectedWeek);
	const [action, setAction] = useState("");
	const [toggleDiscordLink, setToggleDiscordLink] = useState(false);
	const [toggleRenameCharacter, setToggleRenameCharacter] = useState(false);
	const [disabledLink, setDisabledLink] = useState(false);
	const [statusMessage, setStatusMessage] = useState("");
	const [successful, setSuccessful] = useState(true);
	const [selectedWeekFE, setSelectedWeekFE] = useState("");
	const [selectedCharacterID, setSelectedCharacterID] = useState(0);
	const [searchCharacter, setSearchCharacter] = useState("");
	const [newCharacterName, setNewCharacterName] = useState("");
	const [importedData, setImportedData] = useState("");
	const [importedDataStatus, setImportedDataStatus] = useState("");

	useEffect(() => {
		if (selectedWeekFE !== "") {
			store.dispatch(setSelectedWeek(selectedWeekFE));
		}
	}, [selectedWeekFE]);

	useEffect(() => {
		if (token !== "" && action === "culvert_score" && selectedWeek !== null) {
			fetchCharacterScores(token, selectedWeek).then((res) => {
				if (typeof res === "number") {
					setSuccessful(false);
					setStatusMessage("Failed with error " + res);
					return;
				}
				store.dispatch(setCharacterScores(res));
			});
		}
	}, [selectedWeek, token, action]);

	useEffect(() => {
		if (updateCulvertScoresResult !== null) {
			updateCulvertScoresResult.then((res) => {
				setDisabledLink(false);
				setSuccessful(res.status === 200);
				setStatusMessage(res.statusMessage);
				store.dispatch(resetCharacterScores());
			});
		}
	}, [updateCulvertScoresResult]);

	useEffect(() => {
		if (
			(action === "culvert_score" ||
				action === "rename_character" ||
				action === "link_member") &&
			Object.values(characters).length === 0
		) {
			console.log("action get characters");
			fetchCharacters(token).then((res) => {
				if (typeof res === "number") {
					setSuccessful(false);
					setStatusMessage("Failed with error " + res);
					return;
				}
				if (res.length > 0) store.dispatch(setCharacters(res));
			});
		}
	}, [action, characters, token]);
	useEffect(() => {
		if (
			action === "culvert_score" &&
			Object.values(characters).length !== 0 &&
			characterScores === null
		) {
			console.log("action get character scores");
			fetchCharacterScores(
				token,
				selectedWeek !== null ? selectedWeek : "",
			).then((res) => {
				if (typeof res === "number") {
					setSuccessful(false);
					setStatusMessage("Failed with error " + res);
					return;
				}
				store.dispatch(setCharacterScores(res));
			});
		}
	}, [action, characters, token, characterScores, selectedWeek]);
	useEffect(() => {
		// claims expired
		if (
			claims.exp !== "0" &&
			Number(claims.exp) * 1000 < new Date().getTime()
		) {
			alert("Expired login token");
			store.dispatch(resetToken());
			return;
		}
		// if new token was entered
		if (token !== "") {
			(async () => {
				console.log("fetching members");
				const res = await fetchMembers(token);
				if (typeof res === "number") {
					console.log("failed to get members", res);
					if (res === 401) {
						// Using store's dispatch to go around react hook exhaustive deps
						store.dispatch(resetToken());
					}
					return;
				}
				if (res.length > 0) store.dispatch(setMembers(res));
				setAction("culvert_score");
			})();
		}
	}, [token, claims]);

	useEffect(() => {
		// Handle importedData onChange
		if (!importedData) {
			return;
		}
		setImportedDataStatus("");
		let importedDataObj: ImportedData;
		try {
			importedDataObj = JSON.parse(importedData);
		} catch (e) {
			setImportedDataStatus("Failed to import. Likely wrong syntax.");
			return;
		}
		let characterMap: { [key: string]: number } = {};
		for (const [id, char] of Object.entries(characters)) {
			characterMap[char] = Number(id);
		}
		let importedScores: { [key: number]: number } = {};
		let scoreErrors: { [key: string]: number } = {};
		for (const [charName, score] of Object.entries(importedDataObj)) {
			if (characterMap[charName]) {
				importedScores[characterMap[charName]] = score;
			} else {
				scoreErrors[charName] = score;
			}
		}
		for (const [id, score] of Object.entries(importedScores)) {
			store.dispatch(addNewCharacterScore(Number(id)));
			store.dispatch(updateScoreValue({ character_id: Number(id), score }));
		}
		if (Object.keys(scoreErrors).length === 0) {
			setImportedDataStatus("Successfully imported all characters");
		} else {
			setImportedDataStatus(
				"Imported partially, errors with these chars/scores\n" +
					JSON.stringify(scoreErrors, null, 2),
			);
		}
		setImportedData("");
	}, [importedData, characters]);

	const untrackCharacter = (member: GuildMember, charID: string) => {
		const res = linkDiscordMaple(
			token,
			member.discord_user_id,
			characters[Number(charID)],
			false,
			true,
		);
		res
			.then((res) => {
				if (res.status === 200) {
					setSuccessful(true);
					setStatusMessage;
					store.dispatch(setCharacters([]));
				} else {
					setStatusMessage(
						"Error unlinking discord server: " + res.status + " " + res.payload,
					);
					setSuccessful(false);
				}
			})
			.catch((err) => {
				console.error(err);
				setStatusMessage("Error unlinking discord client: " + err.toString());
				setSuccessful(false);
			});
	};

	return (
		<div className="App">
			<header className="App-header">
				<Login />
				{token !== "" && (
					<div className="m-5">
						<Navbar
							expand="lg"
							sticky="top"
							className="bg-body-tertiary"
							variant="light"
						>
							<Container
								style={{ justifyContent: "space-between", maxWidth: "95%" }}
							>
								<Navbar.Collapse id="basic-navbar-nav">
									<Nav className="me-auto">
										<button
											className="btn btn-primary"
											onClick={() => {
												navigate("/edit-settings");
											}}
										>
											Edit Global Settings
										</button>
									</Nav>
								</Navbar.Collapse>
							</Container>
						</Navbar>
					</div>
				)}
				{statusMessage !== "" && (
					<div className="m-5" style={{ color: successful ? "green" : "red" }}>
						{statusMessage}
					</div>
				)}
				{action === "culvert_score" && (
					<div>
						{editableWeeks !== null && (
							<div style={{ display: "flex", flexDirection: "column" }}>
								<textarea
									style={{ resize: "none" }}
									value={importedData}
									rows={3}
									placeholder="Select date first, then
Paste data here to quickly set values.
Don't forget to submit"
									onChange={(e) => {
										setImportedData(e.target.value);
									}}
								></textarea>
								{importedDataStatus !== "" && <p>{importedDataStatus}</p>}
								<select
									onChange={(e) => {
										setSelectedWeekFE(e.target.value);
									}}
								>
									{editableWeeks.map((d) => (
										<option key={`editable-weeks-${d}`} value={d}>
											{d}
										</option>
									))}
								</select>
							</div>
						)}
						<br />
						<Navbar
							expand="lg"
							sticky="top"
							className="bg-body-tertiary"
							variant="light"
						>
							<Container
								style={{ justifyContent: "space-between", maxWidth: "95%" }}
							>
								<Navbar.Collapse id="basic-navbar-nav">
									<Nav className="me-auto">
										<button
											className="btn btn-success"
											onClick={() => {
												navigate("/newchar");
											}}
										>
											+ track character
										</button>
										<button
											className="btn btn-link"
											onClick={() => {
												setToggleDiscordLink(!toggleDiscordLink);
											}}
										>
											Toggle edit discord link
										</button>
										<button
											className="btn btn-link"
											onClick={() => {
												setToggleRenameCharacter(!toggleRenameCharacter);
											}}
										>
											Toggle rename character
										</button>
										<button
											className="btn btn-link"
											onClick={() => {
												navigator.clipboard.writeText(
													JSON.stringify(Object.values(characters), null, 4),
												);
												alert("copied");
											}}
										>
											Copy maple character names to clipboard
										</button>
									</Nav>
								</Navbar.Collapse>
								<Navbar.Collapse
									id="basic-navbar-nav-submit"
									style={{ flexGrow: "0" }}
								>
									<Nav
										className="me-auto"
										style={{ marginRight: "0px !important" }}
									>
										<button
											disabled={disabledLink}
											className="btn btn-primary"
											onClick={() => {
												setImportedDataStatus("");
												setDisabledLink(true);
												console.log("apply changes for culvert scores");
												store.dispatch(applyCulvertChanges(token));
											}}
										>
											Submit
										</button>
									</Nav>
								</Navbar.Collapse>
							</Container>
						</Navbar>
						<table>
							<thead>
								<tr>
									<th>Discord user</th>
									<th>Character name</th>
									<th>Last week</th>
									<th>This week</th>
									<th>Addition actions</th>
								</tr>
							</thead>
							<tbody>
								{Object.entries(characters)
									.sort(([charID1], [charID2]) => {
										return characters[Number(charID1)] >=
											characters[Number(charID2)]
											? 1
											: -1;
									})
									.map(([charID], i) => {
										const scores = characterScores
											? characterScores[Number(charID)] || {}
											: {};
										return (
											<tr key={"scores-" + i}>
												<td>
													<span>
														{membersCharacters &&
															Object.entries(membersCharacters).map(
																([discordID, charIDs], i) => {
																	if (
																		charIDs.includes(Number(charID)) &&
																		membersByID[discordID]
																	) {
																		const member = members.find((member) => {
																			return (
																				member.discord_user_id === discordID
																			);
																		});
																		return toggleDiscordLink ||
																			discordID === "2" ? (
																			<button
																				key={"discord_name-button-" + i}
																				className="btn btn-warning"
																				onClick={() => {
																					navigate(`/linkdiscord?id=${charID}`);
																				}}
																			>
																				{member?.discord_nickname ||
																					member?.discord_global_name ||
																					member?.discord_username ||
																					membersByID[discordID]}
																			</button>
																		) : (
																			<span key={"discord_name-button-" + i}>
																				{member?.discord_nickname ||
																					member?.discord_global_name ||
																					member?.discord_username ||
																					membersByID[discordID]}
																			</span>
																		);
																	}
																	return null;
																},
															)}
													</span>
												</td>
												<td>
													{toggleRenameCharacter ? (
														(i + 1) % 17 === 0 && i !== 0 ? (
															<button
																className="btn btn-warning"
																style={{ textDecoration: "underline" }}
																onClick={() => {
																	navigate(`/rename?id=${charID}`);
																}}
															>
																{characters[Number(charID)] || charID}
															</button>
														) : (
															<button
																className="btn btn-warning"
																onClick={() => {
																	navigate(`/rename?id=${charID}`);
																}}
															>
																{characters[Number(charID)] || charID}
															</button>
														)
													) : (i + 1) % 17 === 0 && i !== 0 ? (
														<span style={{ textDecoration: "underline" }}>
															{characters[Number(charID)] || charID}
														</span>
													) : (
														<span>{characters[Number(charID)] || charID}</span>
													)}
												</td>
												<td>
													<input
														placeholder={scores.prev?.toString()}
														disabled={true}
													/>
												</td>
												<td>
													<input
														onChange={(e) => {
															const n = Number(e.target.value);
															if (!Number.isNaN(n)) {
																store.dispatch(
																	updateScoreValue({
																		score: n,
																		character_id: Number(charID),
																	}),
																);
															}
														}}
														value={scores.current || ""}
													/>
												</td>
												<td>
													{(scores.current || 0) <= 0 &&
														membersCharacters &&
														[
															Object.entries(membersCharacters).find(
																([, linkedCharacters]) => {
																	return linkedCharacters.includes(
																		Number(charID),
																	);
																},
															),
														].map((entry) => {
															if (
																entry === undefined ||
																(scores.prev || 0) !== 0
															)
																return null;
															const [discordID] = entry;
															return (
																<button
																	key={"untrack-character-" + charID}
																	className="btn btn-danger"
																	onClick={() => {
																		untrackCharacter(
																			{
																				discord_user_id: discordID,
																				discord_global_name: "",
																				discord_nickname: "",
																				discord_username: "",
																			},
																			charID,
																		);
																	}}
																>
																	Untrack {characters[Number(charID)]} from bot
																</button>
															);
														})}
												</td>
											</tr>
										);
									})}
							</tbody>
						</table>
					</div>
				)}
				{action === "rename_character" && ( // We no longer change the action variable's value
					<div>
						<div>
							{selectedCharacterID !== 0 && (
								<div>Selected: {characters[selectedCharacterID]}</div>
							)}
							<input
								type="text"
								placeholder="character name"
								value={searchCharacter}
								onChange={(e) => {
									setSearchCharacter(e.target.value);
								}}
							/>
							{searchCharacter !== "" && (
								<div>
									{Object.keys(characters)
										.filter((m) => {
											return (
												characters[Number(m)]
													.toLowerCase()
													.includes(searchCharacter.toLowerCase()) ||
												characters[Number(m)]
													.toLowerCase()
													.includes(searchCharacter.toLowerCase()) ||
												characters[Number(m)]
													.toLowerCase()
													.includes(searchCharacter.toLowerCase())
											);
										})
										.map((m) => (
											<button
												key={"rename_character-select-character-" + m}
												className="btn btn-success"
												onClick={() => {
													setSelectedCharacterID(Number(m));
												}}
											>
												{characters[Number(m)]}
											</button>
										))}
								</div>
							)}
						</div>
						<input
							onChange={(e) => {
								setNewCharacterName(e.target.value);
							}}
							value={newCharacterName}
							placeholder="new name"
						/>
						<br />
						<button
							className="btn btn-danger"
							disabled={disabledLink}
							onClick={() => {
								setDisabledLink(true);
								renameCharacter(token, {
									character_id: selectedCharacterID,
									new_name: newCharacterName,
									bypass_name_check: false, // literally dead old code.
								}).then((res) => {
									setDisabledLink(false);
									if (res.status !== 200) {
										setSuccessful(false);
										setStatusMessage(res.payload);
									} else {
										setSuccessful(true);
										setStatusMessage(
											"Successfully renamed to " + newCharacterName,
										);
										store.dispatch(setCharacters([]));
									}
								});
							}}
						>
							Rename
						</button>
					</div>
				)}
			</header>
		</div>
	);
}

export default App;

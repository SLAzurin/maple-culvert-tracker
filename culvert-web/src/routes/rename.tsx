import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import {
	resetInitialStateCharacters,
	selectCharacters,
} from "../features/characters/charactersSlice";
import { selectToken } from "../features/login/loginSlice";
import { store } from "../app/store";
import renameCharacter from "../helpers/renameCharacter";

const Rename = () => {
	const navigate = useNavigate();
	const characters = useSelector(selectCharacters);
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [newName, setNewName] = useState("");
	const [bypassNameCheck, setBypassNameCheck] = useState(false);
	const [disabled, setDisabled] = useState(false);

	const [charID, setCharID] = useState("0");
	useEffect(() => {
		const queryString = window.location.search;
		const query = new URLSearchParams(queryString);
		const id = query.get("id");
		if (!id) {
			navigate(-1);
			return;
		}
		if (Number.isNaN(Number(id))) {
			navigate(-1);
			return;
		}
		if (!characters[Number(id)]) {
			navigate(-1);
			return;
		}
		setNewName(characters[Number(id)]);
		setCharID(id);
		return;
	}, []);
	return (
		<div>
			<h1>Rename - {characters[Number(charID)]}</h1>
			{status !== "" && <h2>Status: {status}</h2>}
			<input
				value={newName}
				placeholder="New Name"
				onChange={(e) => {
					setNewName(e.target.value);
				}}
			></input>
			<button
				disabled={disabled}
				className="btn btn-primary"
				onClick={async () => {
					if (newName.length <= 2) {
						setStatus("Error: Character Name is too short");
						return;
					}
					setStatus("Renaming character...");
					setDisabled(true);
					const res = await renameCharacter(token, {
						character_id: Number(charID),
						new_name: newName,
						bypass_name_check: bypassNameCheck,
					});
					if (res.status !== 200) {
						setDisabled(false);
						setStatus(`Error: ${res.status} ${res.payload}`);
						return;
					}
					store.dispatch(resetInitialStateCharacters());
					navigate("/");
				}}
			>
				Submit
			</button>
			<form>
				<input
					id="bypass"
					type="checkbox"
					checked={bypassNameCheck}
					onChange={(e) => setBypassNameCheck(e.target.checked)}
					className="m-2"
				></input>
				<label htmlFor="bypass">
					Skip name verification with official rankings
				</label>
			</form>
		</div>
	);
};

export default Rename;

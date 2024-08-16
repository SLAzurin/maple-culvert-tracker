import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { resetInitialStateCharacters } from "../features/characters/charactersSlice";
import { selectToken } from "../features/login/loginSlice";
import { store } from "../app/store";
import linkDiscordMaple from "../helpers/linkDiscordMaple";

const NewChar = () => {
	const navigate = useNavigate();
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [characterName, setCharacterName] = useState("");
	const [bypassNameCheck, setBypassNameCheck] = useState(false);

	return (
		<div>
			<h1>Track new character</h1>
			{status !== "" && <h2>Status: {status}</h2>}
			<input
				value={characterName}
				placeholder="Character Name"
				onChange={(e) => {
					setCharacterName(e.target.value);
				}}
			></input>
			<button
				className="btn btn-primary"
				onClick={async () => {
					if (characterName.length <= 2) {
						setStatus("Error: Character Name is too short");
						return;
					}
					linkDiscordMaple(
						token,
						"2",
						characterName,
						true,
						bypassNameCheck,
					).then((res) => {
						if (res.status !== 200) {
							setStatus(`Error: ${res.status} ${res.payload}`);
							return;
						}
						store.dispatch(resetInitialStateCharacters());
						navigate("/");
					});
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

export default NewChar;

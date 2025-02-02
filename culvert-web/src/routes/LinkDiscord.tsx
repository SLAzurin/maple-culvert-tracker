import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import {
	resetInitialStateCharacters,
	selectCharacters,
} from "../features/characters/charactersSlice";
import { selectMembers, setMembers } from "../features/members/membersSlice";
import { selectToken } from "../features/login/loginSlice";
import linkDiscordMaple from "../helpers/linkDiscordMaple";
import { store } from "../app/store";
import GuildMember from "../types/GuildMember";

const LinkDiscord = () => {
	const navigate = useNavigate();
	const characters = useSelector(selectCharacters);
	const members = useSelector(selectMembers);
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [disabled, setDisabled] = useState(false);

	const [charID, setCharID] = useState("0");

	const link = (member: GuildMember) => {
		setDisabled(true);
		setStatus("Linking character...");
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
					const res2 = linkDiscordMaple(
						token,
						member.discord_user_id,
						characters[Number(charID)],
						true,
						true,
					);
					res2
						.then((res) => {
							if (res.status === 200) {
								store.dispatch(setMembers([]));
								store.dispatch(resetInitialStateCharacters());
								navigate("/");
							} else {
								setDisabled(false);
								setStatus(
									"Error linking discord server: " +
										res.status +
										" " +
										res.payload,
								);
							}
						})
						.catch((err) => {
							console.error(err);
							setDisabled(false);
							setStatus("Error linking discord client: " + err.toString());
						});
				} else {
					setDisabled(false);
					setStatus(
						"Error unlinking discord server: " + res.status + " " + res.payload,
					);
				}
			})
			.catch((err) => {
				console.error(err);
				setDisabled(false);
				setStatus("Error unlinking discord client: " + err.toString());
			});
	};

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
		setCharID(id);
		return;
	}, []);

	return (
		<div>
			<h1>Link Discord - {characters[Number(charID)]}</h1>
			{status !== "" && <h2>{status}</h2>}

			<div>
				<button
					disabled={disabled}
					className="btn btn-warning mt-3"
					onClick={() =>
						link({
							discord_user_id: "2",
							discord_global_name: "",
							discord_nickname: "",
							discord_username: "",
						})
					}
				>
					Unlink from discord, but they're still a guildmate
				</button>
			</div>
			<br />
			<div>
				{members
					.filter((v) => v.discord_user_id !== "2")
					.sort((a, b) => {
						const aName =
							a.discord_nickname ||
							a.discord_global_name ||
							a.discord_username ||
							a.discord_user_id;
						const bName =
							b.discord_nickname ||
							b.discord_global_name ||
							b.discord_username ||
							b.discord_user_id;
						return aName.toLowerCase().localeCompare(bName.toLowerCase());
					})
					.map((member, i) => (
						<button
							disabled={disabled}
							className="btn btn-link"
							key={i}
							onClick={() => link(member)}
						>
							{member.discord_nickname ||
								member.discord_global_name ||
								member.discord_username ||
								member.discord_user_id}
						</button>
					))}
			</div>
		</div>
	);
};

export default LinkDiscord;

import { useDispatch, useSelector } from "react-redux";
import { selectClaims, setToken, selectToken } from "./loginSlice";

export function Login() {
	const token = useSelector(selectToken);
	const claims = useSelector(selectClaims);
	const dispatch = useDispatch();

	return (
		<div
			style={{
				backgroundColor:
					claims && claims.dev_mode === 1 ? "#ffebef" : undefined,
			}}
		>
			{claims &&
				claims.exp !== "0" &&
				"Expires " + new Date(Number(claims.exp) * 1000).toString()}
			<div>
				Login token:{" "}
				<input
					type="text"
					onChange={(e) => {
						dispatch(setToken(e.target.value));
					}}
					value={token}
				/>
			</div>
			{claims?.discord_username && (
				<div>
					<br />
					<p>Welcome {claims.discord_username}!</p>
					{claims.dev_mode === 1 && <p>THIS IS IN DEV MODE.</p>}
				</div>
			)}
		</div>
	);
}

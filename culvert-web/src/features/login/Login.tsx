import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
	selectClaims,
	selectAuthenticated,
	setClaims,
	resetToken,
} from "./loginSlice";
import loginWithToken from "../../helpers/loginWithToken";
import logout from "../../helpers/logout";

export function Login() {
	const authenticated = useSelector(selectAuthenticated);
	const claims = useSelector(selectClaims);
	const dispatch = useDispatch();
	const [tokenInput, setTokenInput] = useState("");
	const [error, setError] = useState("");
	const [pending, setPending] = useState(false);

	const submitLogin = async (tokenOverride?: string) => {
		const token = (tokenOverride ?? tokenInput).trim();
		if (!token) {
			setError("Paste a login token from Discord /login");
			return;
		}
		setPending(true);
		setError("");
		const res = await loginWithToken(token);
		setPending(false);
		if (typeof res === "number") {
			setError(
				res === 401
					? "Invalid or expired login token"
					: "Login failed (status " + res + ")",
			);
			return;
		}
		setTokenInput("");
		dispatch(setClaims(res));
	};

	const submitLogout = async () => {
		setPending(true);
		await logout();
		setPending(false);
		dispatch(resetToken());
	};

	return (
		<div
			style={{
				backgroundColor:
					claims && claims.dev_mode === 1 ? "#ffebef" : undefined,
			}}
		>
			{authenticated &&
				claims.exp !== "0" &&
				"Expires " + new Date(Number(claims.exp) * 1000).toString()}
			{!authenticated ? (
				<div>
					Login token:{" "}
					<input
						type="password"
						autoComplete="off"
						disabled={pending}
						onChange={(e) => {
							setTokenInput(e.target.value);
						}}
						onPaste={(e) => {
							const pasted = e.clipboardData.getData("text").trim();
							if (!pasted) return;
							e.preventDefault();
							setTokenInput(pasted);
							void submitLogin(pasted);
						}}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								void submitLogin();
							}
						}}
						value={tokenInput}
					/>
					<button
						className="btn btn-primary ms-2"
						disabled={pending}
						onClick={() => {
							void submitLogin();
						}}
					>
						Login
					</button>
					{error !== "" && (
						<p style={{ color: "red" }}>{error}</p>
					)}
				</div>
			) : (
				<div>
					<button
						className="btn btn-secondary"
						disabled={pending}
						onClick={() => {
							void submitLogout();
						}}
					>
						Logout
					</button>
				</div>
			)}
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

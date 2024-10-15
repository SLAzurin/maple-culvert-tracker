import { createSlice, PayloadAction } from "@reduxjs/toolkit";

import { RootState } from "../../app/store";
import MCTClaims from "../../types/MCTClaims";
import validateJWT from "../../helpers/validateJWT";

export interface LoginState {
	token: string;
	claims: MCTClaims;
}

const initialState: LoginState = {
	token: "",
	claims: {
		exp: "0",
		discord_username: "",
		discord_server_id: "",
		dev_mode: 0,
	},
};

export const loginSlice = createSlice({
	name: "login",
	initialState,
	reducers: {
		setToken: (state, action: PayloadAction<string>) => {
			const { valid, claims } = validateJWT(action.payload);
			if (!valid) return;
			state.token = action.payload;
			state.claims = claims as MCTClaims;
		},
		resetToken: (state) => {
			state.claims = initialState.claims;
			state.token = initialState.token;
		},
	},
});
export default loginSlice.reducer;

export const selectToken = (state: RootState) => state.login.token;
export const selectClaims = (state: RootState) => state.login.claims;

export const { setToken, resetToken } = loginSlice.actions;

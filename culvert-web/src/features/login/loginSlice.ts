import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import { type RootState } from "../../app/store";
import { type MCTClaims } from "../../types/MCTClaims";

export interface LoginState {
	authenticated: boolean;
	claims: MCTClaims;
}

const emptyClaims: MCTClaims = {
	exp: "0",
	discord_username: "",
	discord_server_id: "",
	discord_user_id: "",
	dev_mode: 0,
};

const initialState: LoginState = {
	authenticated: false,
	claims: emptyClaims,
};

export const loginSlice = createSlice({
	name: "login",
	initialState,
	reducers: {
		setClaims: (state, action: PayloadAction<MCTClaims>) => {
			state.authenticated = true;
			state.claims = action.payload;
		},
		resetToken: (state) => {
			state.authenticated = false;
			state.claims = emptyClaims;
		},
	},
});
export default loginSlice.reducer;

export const selectAuthenticated = (state: RootState) =>
	state.login.authenticated;
export const selectClaims = (state: RootState) => state.login.claims;

export const { setClaims, resetToken } = loginSlice.actions;

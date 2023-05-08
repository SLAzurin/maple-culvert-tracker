import { createSlice, PayloadAction } from "@reduxjs/toolkit"

import { RootState } from "../../app/store"

export interface LoginState {
  token: string
  status: "idle" | "loading" | "failed"
}

const initialState: LoginState = {
  token: "",
  status: "idle",
}

export const loginSlice = createSlice({
  name: "login",
  initialState,
  reducers: {
    setToken: (state, action: PayloadAction<string>) => {
      state.token = action.payload
    },
  },
})
export default loginSlice.reducer

export const selectToken = (state: RootState) => state.login.token

export const { setToken } = loginSlice.actions

import { createSlice, PayloadAction } from "@reduxjs/toolkit"

interface MembersState {}

const initialState: MembersState = {}

export const membersSlice = createSlice({
  name: "members",
  initialState,
  reducers: {},
})
export default membersSlice.reducer

// export const selectToken = (state: RootState) => state.login.token
// export const selectClaims = (state: RootState) => state.login.claims

// export const {  } = membersSlice.actions

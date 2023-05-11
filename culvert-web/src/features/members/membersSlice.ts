import { PayloadAction, createSlice } from "@reduxjs/toolkit"
import GuildMember from "../../types/GuildMember"
import { RootState } from "../../app/store"

interface MembersState {
  members: GuildMember[]
  membersByID: {
    [key: string]: string
  }
}

const initialState: MembersState = {
  members: [],
  membersByID: {},
}

export const membersSlice = createSlice({
  name: "members",
  initialState,
  reducers: {
    setMembers: (state, action: PayloadAction<GuildMember[]>) => {
      state.members = action.payload.sort((a, b) => {
        if (
          a.discord_username.toLowerCase() === b.discord_username.toLowerCase()
        )
          return 0
        return a.discord_username.toLowerCase() >
          b.discord_username.toLowerCase()
          ? 1
          : -1
      })
      const newMembersByID: {
        [key: string]: string
      } = {}
      for (const v of action.payload) {
        newMembersByID[v.discord_username] = v.discord_user_id
      }
      state.membersByID = newMembersByID
    },
  },
})
export default membersSlice.reducer

export const selectMembers = (state: RootState) => state.members.members
export const selectMembersByID = (state: RootState) => state.members.membersByID

export const { setMembers } = membersSlice.actions
// TODO: notes
/* 

sorting done in frontend 


scores: (MUST FRONTEND SORT, by gpq desc)
[]{
  charid
  score
}

transform into
id -> {
  last week,
  this week
}

frontend:
table:
charname, last week, this week (input edit old value or empty field. POST/PATCH)

post request:
filter by isNew
  2 payloads
make 2 post requests

*/

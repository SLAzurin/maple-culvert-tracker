import { PayloadAction, createSlice } from "@reduxjs/toolkit"
import { RootState } from "../../app/store"

interface CharactersState {
  characters: { [key: number]: string }
  characterScores: {
    [key: number]: {
      prev?: number
      current?: number
    }
  }
}

const initialState: CharactersState = {
  characters: {},
  characterScores: {},
}

export const membersSlice = createSlice({
  name: "members",
  initialState,
  reducers: {
    setCharacters: (
      state,
      action: PayloadAction<{ character_id: number; character_name: string }[]>,
    ) => {
      const newCharacters: { [key: number]: string } = {}
      for (let v of action.payload) {
        newCharacters[v.character_id] = v.character_name
      }
      state.characters = newCharacters
    },
    setCharacterScores: (
      state,
      action: PayloadAction<{
        current: string
        data: { character_id: number; culvert_date: string; score: number }[]
      }>,
    ) => {
      const newScores: {
        [key: number]: {
          prev?: number
          current?: number
        }
      } = {}
      for (let v of action.payload.data) {
        if (typeof newScores[v.character_id] === "undefined") {
          newScores[v.character_id] = {}
        }
        if (action.payload.current === v.culvert_date) {
          newScores[v.character_id].current = v.score
        } else {
          newScores[v.character_id].prev = v.score
        }
      }
      state.characterScores = newScores
    },
  },
})
export default membersSlice.reducer

export const selectCharacters = (state: RootState) =>
  state.characters.characters
export const selectCharacterScores = (state: RootState) =>
  state.characters.characterScores

export const { setCharacters, setCharacterScores } = membersSlice.actions

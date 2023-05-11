import { PayloadAction, createSlice } from "@reduxjs/toolkit"
import { RootState } from "../../app/store"
import updateCulvertScores from "../../helpers/updateCulvertScores"

interface CharactersState {
  characters: { [key: number]: string }
  characterScores: {
    [key: number]: {
      prev?: number
      current?: number
    }
  }
  characterScoresOriginal: {
    [key: number]: {
      prev?: number
      current?: number
    }
  }
  updateCulvertScoresResult: Promise<{
    status: number
    statusMessage: string
    date: Date
  }> | null
}

const initialState: CharactersState = {
  characters: {},
  characterScores: {},
  characterScoresOriginal: {},
  updateCulvertScoresResult: null,
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
    resetCharacterScores: (state) => {
      state.characterScoresOriginal = {}
      state.characterScores = {}
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
      state.characterScoresOriginal = newScores
    },
    updateScoreValue: (
      state,
      action: PayloadAction<{ character_id: number; score: number }>,
    ) => {
      const newScores = { ...state.characterScores }
      newScores[action.payload.character_id].current = action.payload.score
      state.characterScores = newScores
    },
    addNewCharacterScore: (state, action: PayloadAction<number>) => {
      const newScores = { ...state.characterScores }
      newScores[action.payload] = {}
      state.characterScores = newScores
    },
    applyCulvertChanges: (state, action: PayloadAction<string>) => {
      const edit: {
        payload: { character_id: number; score: number }[]
        isNew: boolean
      } = {
        payload: [],
        isNew: false,
      }
      const _new: {
        payload: { character_id: number; score: number }[]
        isNew: boolean
      } = {
        payload: [],
        isNew: true,
      }

      for (let [charID, { current }] of Object.entries(state.characterScores)) {
        if (!state.characterScoresOriginal[Number(charID)]) {
          _new.payload.push({
            character_id: Number(charID),
            score: current || 0,
          })
        } else if (
          state.characterScoresOriginal[Number(charID)].current !== current
        ) {
          edit.payload.push({
            character_id: Number(charID),
            score: current || 0,
          })
        }
      }

      state.updateCulvertScoresResult = (async () => {
        const mainRes: {
          status: number
          statusMessage: string
          date: Date
        } = { status: 200, date: new Date(), statusMessage: "" }
        if (_new.payload.length !== 0) {
          const res = await updateCulvertScores(action.payload, _new)
          if (res.status !== 200) {
            mainRes.status = res.status
            mainRes.statusMessage = res.payload as string
          }
        }
        if (edit.payload.length !== 0) {
          const res = await updateCulvertScores(action.payload, edit)
          if (res.status !== 200) {
            mainRes.status = res.status
            mainRes.statusMessage += " " + (res.payload as string)
          }
        }
        if (mainRes.status === 200) {
          mainRes.statusMessage = "Successfully updated all character scores"
        }
        return mainRes
      })()
    },
  },
})
export default membersSlice.reducer

export const selectCharacters = (state: RootState) =>
  state.characters.characters
export const selectCharacterScores = (state: RootState) =>
  state.characters.characterScores
export const selectUpdateCulvertScoresResult = (state: RootState) =>
  state.characters.updateCulvertScoresResult

export const {
  setCharacters,
  setCharacterScores,
  updateScoreValue,
  addNewCharacterScore,
  applyCulvertChanges,
  resetCharacterScores,
} = membersSlice.actions

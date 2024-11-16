import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { RootState } from "../../app/store";
import updateCulvertScores from "../../helpers/updateCulvertScores";

type character_id = number;
type score = number;
type week = string;
type character_name = string;

interface CharactersState {
	characters: { [key: character_id]: character_name };
	membersCharacters: { [key: character_name]: character_id[] };
	characterScores: {
		[key: character_id]: {
			prev?: number;
			current?: number;
		};
	} | null;
	characterScoresOriginal: {
		[key: character_id]: {
			prev?: number;
			current?: number;
		};
	} | null;
	updateCulvertScoresResult: Promise<{
		status: number;
		statusMessage: string;
		date: Date;
	}> | null;
	selectedWeek: week | null;
	editableWeeks: week[] | null;
	unsubmittedScores: { week: week | null; scores: Record<character_id, score> };
	scoresFetched: null | boolean;
}

const initialState: CharactersState = {
	characters: {},
	membersCharacters: {},
	characterScores: null,
	characterScoresOriginal: {},
	updateCulvertScoresResult: null,
	selectedWeek: null,
	editableWeeks: null,
	unsubmittedScores: {
		week: null,
		scores: {},
	},
	scoresFetched: null,
};

export const charactersSlice = createSlice({
	name: "characters",
	initialState,
	reducers: {
		setCharacters: (
			state,
			action: PayloadAction<
				{
					character_id: number;
					character_name: string;
					discord_user_id: string;
				}[]
			>,
		) => {
			const newCharacters: { [key: number]: string } = {};
			const newMembersCharacters: { [key: string]: number[] } = {};
			for (let v of action.payload) {
				newCharacters[v.character_id] = v.character_name;
				if (!newMembersCharacters[v.discord_user_id])
					newMembersCharacters[v.discord_user_id] = [];
				newMembersCharacters[v.discord_user_id].push(v.character_id);
			}
			state.characters = newCharacters;
			state.membersCharacters = newMembersCharacters;
		},
		resetCharacterScores: (state) => {
			// called after successful scores submit
			state.characterScoresOriginal = null;
			state.characterScores = null;
			const newUnsubmittedScores = {
				week: state.unsubmittedScores.week,
				scores: {},
			};
			state.unsubmittedScores = newUnsubmittedScores;
		},
		setCharacterScores: (
			state,
			action: PayloadAction<{
				weeks: string[];
				data: { character_id: number; culvert_date: string; score: number }[];
			}>,
		) => {
			const newScores: {
				[key: number]: {
					prev?: number;
					current?: number;
				};
			} = {};
			if (state.editableWeeks == null) {
				state.editableWeeks = action.payload.weeks;
			}
			if (state.editableWeeks.length > 0 && state.selectedWeek == null) {
				if (
					state.unsubmittedScores.week !== null &&
					state.editableWeeks.includes(state.unsubmittedScores.week) &&
					Object.keys(state.unsubmittedScores.scores).length > 0
				) {
					state.selectedWeek = state.unsubmittedScores.week;
				} else {
					state.selectedWeek = action.payload.weeks[0];
					state.unsubmittedScores = {
						week: action.payload.weeks[0],
						scores: {},
					};
				}
				state.scoresFetched = true;
			}

			for (let v of action.payload.data) {
				if (typeof state.characters[v.character_id] === "undefined") {
					continue;
				}
				if (typeof newScores[v.character_id] === "undefined") {
					newScores[v.character_id] = {};
				}
				if (state.selectedWeek === v.culvert_date) {
					newScores[v.character_id].current = v.score;
				} else {
					newScores[v.character_id].prev = v.score;
				}
			}
			state.characterScores = newScores;
			state.characterScoresOriginal = newScores;
		},
		updateScoreValue: (
			state,
			action: PayloadAction<{ character_id: number; score: number }>,
		) => {
			const newScores = { ...state.characterScores };
			if (typeof newScores[action.payload.character_id] === "undefined") {
				newScores[action.payload.character_id] = {};
			}
			newScores[action.payload.character_id].current = action.payload.score;
			state.characterScores = newScores;
			const newUnsubmittedScores = { ...state.unsubmittedScores };
			newUnsubmittedScores.scores[action.payload.character_id] =
				action.payload.score;
			state.unsubmittedScores = newUnsubmittedScores;
		},
		addNewCharacterScore: (state, action: PayloadAction<number>) => {
			const newScores = { ...state.characterScores };
			newScores[action.payload] = {};
			state.characterScores = newScores;
		},
		applyCulvertChanges: (state, action: PayloadAction<string>) => {
			if (
				state.characterScores === null ||
				state.characterScoresOriginal === null
			)
				return;
			const edit: {
				payload: { character_id: number; score: number }[];
				isNew: boolean;
				week: string;
			} = {
				payload: [],
				isNew: false,
				week: state.selectedWeek !== null ? state.selectedWeek : "",
			};
			const _new: {
				payload: { character_id: number; score: number }[];
				isNew: boolean;
				week: string;
			} = {
				payload: [],
				isNew: true,
				week: state.selectedWeek !== null ? state.selectedWeek : "",
			};

			for (let [charID, { current }] of Object.entries(state.characterScores)) {
				if (
					typeof state.characterScoresOriginal[Number(charID)] ===
						"undefined" ||
					typeof state.characterScoresOriginal[Number(charID)].current ===
						"undefined"
				) {
					_new.payload.push({
						character_id: Number(charID),
						score: current || 0,
					});
				} else if (
					state.characterScoresOriginal[Number(charID)].current !== current
				) {
					edit.payload.push({
						character_id: Number(charID),
						score: current || 0,
					});
				}
			}

			state.updateCulvertScoresResult = (async () => {
				const mainRes: {
					status: number;
					statusMessage: string;
					date: Date;
				} = { status: 200, date: new Date(), statusMessage: "" };
				for (let v of [_new, edit]) {
					if (v.payload.length !== 0) {
						const res = await updateCulvertScores(action.payload, v);
						if (res.status !== 200) {
							mainRes.status = res.status;
							mainRes.statusMessage += (res.payload as string) + " ";
						}
					}
				}
				if (mainRes.status === 200) {
					mainRes.statusMessage = "Successfully updated all character scores";
				}
				return mainRes;
			})();
		},
		setSelectedWeek: (state, action: PayloadAction<string>) => {
			if (state.selectedWeek !== action.payload) {
				state.unsubmittedScores = {
					week: action.payload,
					scores: {},
				};
			}
			state.selectedWeek = action.payload;
			state.scoresFetched = true;
		},
		resetInitialStateCharacters: (state) => {
			// Do all keys except for unsubmittedScores
			Object.keys(state).forEach((key) => {
				if (key !== "unsubmittedScores")
					(state as any)[key] = (initialState as any)[key];
			});
		},
		resetUnsubmittedScores: (state) => {
			const newUnsubmittedScores = {
				week: state.unsubmittedScores.week,
				scores: {},
			};
			state.scoresFetched = null;
			state.unsubmittedScores = newUnsubmittedScores;
		},
		resetScoresFetched: (state) => {
			state.scoresFetched = null;
		},
		setScoresFetched: (state) => {
			state.scoresFetched = true;
		},
	},
});
export default charactersSlice.reducer;

export const selectCharacters = (state: RootState) =>
	state.characters.characters;
export const selectCharacterScores = (state: RootState) =>
	state.characters.characterScores;
export const selectUpdateCulvertScoresResult = (state: RootState) =>
	state.characters.updateCulvertScoresResult;
export const selectEditableWeeks = (state: RootState) =>
	state.characters.editableWeeks;
export const selectSelectedWeek = (state: RootState) =>
	state.characters.selectedWeek;
export const selectMembersCharacters = (state: RootState) =>
	state.characters.membersCharacters;
export const selectScoresFetched = (state: RootState) =>
	state.characters.scoresFetched;
export const selectUnsubmittedScores = (state: RootState) =>
	state.characters.unsubmittedScores;

export const {
	setCharacters,
	setCharacterScores,
	updateScoreValue,
	addNewCharacterScore,
	applyCulvertChanges,
	resetCharacterScores,
	setSelectedWeek,
	resetInitialStateCharacters,
	resetUnsubmittedScores,
	resetScoresFetched,
	setScoresFetched,
} = charactersSlice.actions;

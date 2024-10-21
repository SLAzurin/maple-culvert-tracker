import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { RootState } from "../../app/store";
import updateCulvertScores from "../../helpers/updateCulvertScores";

interface CharactersState {
	characters: { [key: number]: string };
	membersCharacters: { [key: string]: number[] };
	fetchedScoresFromServer: boolean;
	characterScoresGroup: {
		characterScores: {
			[key: number]: {
				prev?: number;
				current?: number;
			};
		} | null;
		characterScoresUnsubmitted: {
			[key: number]: number;
		};
	};
	characterScoresOriginal: {
		[key: number]: {
			prev?: number;
			current?: number;
		};
	} | null;
	updateCulvertScoresResult: Promise<{
		status: number;
		statusMessage: string;
		date: Date;
	}> | null;
	selectedWeek: string | null;
	editableWeeks: string[] | null;
}

const initialState: CharactersState = {
	characters: {},
	membersCharacters: {},
	fetchedScoresFromServer: false,
	characterScoresOriginal: {},
	characterScoresGroup: {
		characterScores: null,
		characterScoresUnsubmitted: {},
	},
	updateCulvertScoresResult: null,
	selectedWeek: null,
	editableWeeks: null,
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
		resetUnsubmittedScores: (state) => {
			const newCharacterScoresGroup = { ...state.characterScoresGroup };
			newCharacterScoresGroup.characterScoresUnsubmitted = {};
			state.characterScoresGroup = newCharacterScoresGroup;
		},
		resetCharacterScores: (state) => {
			state.characterScoresOriginal = null;
			state.characterScoresGroup = {
				characterScores: null,
				characterScoresUnsubmitted: {},
			};
		},
		setCharacterScores: (
			state,
			action: PayloadAction<{
				weeks: string[];
				data: { character_id: number; culvert_date: string; score: number }[];
				wasFetchedFromServer?: boolean;
			}>,
		) => {
			if (action.payload.wasFetchedFromServer) {
				state.fetchedScoresFromServer = true;
			}
			const newScores: {
				[key: number]: {
					prev?: number;
					current?: number;
				};
			} = {};
			if (state.editableWeeks == null) {
				state.editableWeeks = action.payload.weeks;
			}
			if (state.selectedWeek == null) {
				state.selectedWeek = action.payload.weeks[0];
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
			const newCharacterScoresGroup = { ...state.characterScoresGroup };
			newCharacterScoresGroup.characterScores = newScores;
			state.characterScoresGroup = newCharacterScoresGroup;
			state.characterScoresOriginal = newScores;
		},
		updateScoreValue: (
			state,
			action: PayloadAction<{ character_id: number; score: number }>,
		) => {
			const newScoresGroup = { ...state.characterScoresGroup };
			if (
				typeof newScoresGroup.characterScores === "undefined" ||
				newScoresGroup.characterScores === null
			) {
				newScoresGroup.characterScores = {};
			}
			if (
				typeof newScoresGroup.characterScores[action.payload.character_id] ===
				"undefined"
			) {
				newScoresGroup.characterScores[action.payload.character_id] = {};
			}

			newScoresGroup.characterScoresUnsubmitted[action.payload.character_id] =
				action.payload.score;

			newScoresGroup.characterScores[action.payload.character_id].current =
				action.payload.score;
			state.characterScoresGroup = newScoresGroup;
		},
		addNewCharacterScore: (state, action: PayloadAction<number>) => {
			const newScoresGroup = { ...state.characterScoresGroup };
			if (
				typeof newScoresGroup.characterScores === "undefined" ||
				newScoresGroup.characterScores === null
			) {
				newScoresGroup.characterScores = {};
			}
			newScoresGroup.characterScores[action.payload] = {};
			state.characterScoresGroup = newScoresGroup;
		},
		applyCulvertChanges: (state, action: PayloadAction<string>) => {
			if (
				state.characterScoresGroup.characterScores === null ||
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

			for (let [charID, { current }] of Object.entries(
				state.characterScoresGroup.characterScores,
			)) {
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
			state.selectedWeek = action.payload;
		},
		resetInitialStateCharacters: (state) => {
			Object.keys(state).forEach((key) => {
				(state as any)[key] = (initialState as any)[key];
			});
		},
	},
});
export default charactersSlice.reducer;

export const selectCharacters = (state: RootState) =>
	state.characters.characters;
export const selectUpdateCulvertScoresResult = (state: RootState) =>
	state.characters.updateCulvertScoresResult;
export const selectEditableWeeks = (state: RootState) =>
	state.characters.editableWeeks;
export const selectSelectedWeek = (state: RootState) =>
	state.characters.selectedWeek;
export const selectMembersCharacters = (state: RootState) =>
	state.characters.membersCharacters;
export const selectCharacterScoresGroup = (state: RootState) =>
	state.characters.characterScoresGroup;
export const selectFetchedScoresFromServer = (state: RootState) =>
	state.characters.fetchedScoresFromServer;
export const {
	setCharacters,
	setCharacterScores,
	updateScoreValue,
	addNewCharacterScore,
	applyCulvertChanges,
	resetCharacterScores,
	resetUnsubmittedScores,
	setSelectedWeek,
	resetInitialStateCharacters,
} = charactersSlice.actions;

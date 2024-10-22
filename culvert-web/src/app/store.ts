import {
	configureStore,
	ThunkAction,
	Action,
	combineReducers,
} from "@reduxjs/toolkit";
import loginReducer from "../features/login/loginSlice";
import membersReducer from "../features/members/membersSlice";

import { persistStore, persistReducer } from "redux-persist";
import storage from "redux-persist/lib/storage";
import charactersSlice from "../features/characters/charactersSlice";

const persistLoginConfig = {
	key: "login",
	storage,
};

const persistCharactersConfig = {
	key: "characters",
	storage,
	blacklist: [
		"characters",
		"membersCharacters",
		"characterScores",
		"characterScoresOriginal",
		"updateCulvertScoresResult",
		"selectedWeek",
		"editableWeeks",
		// The only omitted fields are characterScores
	],
};

const rootReducer = combineReducers({
	login: persistReducer(persistLoginConfig, loginReducer),
	members: membersReducer,
	characters: persistReducer(persistCharactersConfig, charactersSlice),
});

export const store = configureStore({
	reducer: rootReducer,
	middleware: (getDefaultMiddleware) =>
		getDefaultMiddleware({ serializableCheck: false }),
});

export const persistor = persistStore(store);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
export type AppThunk<ReturnType = void> = ThunkAction<
	ReturnType,
	RootState,
	unknown,
	Action<string>
>;

import {
  configureStore,
  ThunkAction,
  Action,
  combineReducers,
} from "@reduxjs/toolkit"
import loginReducer from "../features/login/loginSlice"
import membersReducer from "../features/members/membersSlice"

import { persistStore, persistReducer } from "redux-persist"
import storage from "redux-persist/lib/storage"
import charactersSlice from "../features/characters/charactersSlice"

const persistConfig = {
  blacklist: ["members", "characters"],
  key: "root",
  storage,
}

const persistedReducer = persistReducer(
  persistConfig,
  combineReducers({
    login: loginReducer,
    members: membersReducer,
    characters: charactersSlice,
  }),
)

export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({ serializableCheck: false }),
})

export const persistor = persistStore(store)

export type AppDispatch = typeof store.dispatch
export type RootState = ReturnType<typeof store.getState>
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  RootState,
  unknown,
  Action<string>
>

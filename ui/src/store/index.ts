import { configureStore } from "@reduxjs/toolkit";
import { useDispatch, useSelector } from "react-redux";
import { accountsReducer } from "../features/accounts/accountsSlice";
import { authReducer } from "../features/auth";
import { mailboxReducer } from "../features/mailbox/mailboxSlice";
import { rulesReducer } from "../features/rules/rulesSlice";
import { annotationsApi } from "../api/annotations";
import { annotationUiReducer } from "./annotationUiSlice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    accounts: accountsReducer,
    mailbox: mailboxReducer,
    rules: rulesReducer,
    [annotationsApi.reducerPath]: annotationsApi.reducer,
    annotationUi: annotationUiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(annotationsApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();

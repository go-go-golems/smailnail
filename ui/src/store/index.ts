import { configureStore } from "@reduxjs/toolkit";
import { useDispatch, useSelector } from "react-redux";
import { accountsReducer } from "../features/accounts/accountsSlice";
import { mailboxReducer } from "../features/mailbox/mailboxSlice";
import { rulesReducer } from "../features/rules/rulesSlice";

export const store = configureStore({
  reducer: {
    accounts: accountsReducer,
    mailbox: mailboxReducer,
    rules: rulesReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();

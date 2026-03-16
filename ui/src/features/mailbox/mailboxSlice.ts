import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api } from "../../api/client";
import type { MailboxInfo, MessageView } from "../../api/types";

interface MailboxState {
  accountId: string | null;
  mailboxes: MailboxInfo[];
  mailboxLoadState: "idle" | "loading" | "loaded" | "error";
  selectedMailbox: string | null;
  messages: MessageView[];
  messageLoadState: "idle" | "loading" | "loaded" | "error";
  totalCount: number;
  offset: number;
  limit: number;
  selectedMessage: MessageView | null;
  messageDetailLoadState: "idle" | "loading" | "loaded" | "error";
  error: string | null;
}

const initialState: MailboxState = {
  accountId: null,
  mailboxes: [],
  mailboxLoadState: "idle",
  selectedMailbox: null,
  messages: [],
  messageLoadState: "idle",
  totalCount: 0,
  offset: 0,
  limit: 20,
  selectedMessage: null,
  messageDetailLoadState: "idle",
  error: null,
};

export const fetchMailboxes = createAsyncThunk(
  "mailbox/fetchMailboxes",
  async (accountId: string) => {
    const res = await api.listMailboxes(accountId);
    return { accountId, mailboxes: res.data };
  },
);

export const fetchMessages = createAsyncThunk(
  "mailbox/fetchMessages",
  async ({
    accountId,
    mailbox,
    offset,
    limit,
  }: {
    accountId: string;
    mailbox: string;
    offset?: number;
    limit?: number;
  }) => {
    const res = await api.listMessages(accountId, {
      mailbox,
      offset: offset ?? 0,
      limit: limit ?? 20,
      includeContent: true,
      contentType: "text/plain",
    });
    const totalCount =
      (res.meta?.["totalCount"] as number | undefined) ?? 0;
    return { messages: res.data, totalCount, mailbox };
  },
);

export const fetchMessageDetail = createAsyncThunk(
  "mailbox/fetchMessageDetail",
  async ({
    accountId,
    mailbox,
    uid,
  }: {
    accountId: string;
    mailbox: string;
    uid: number;
  }) => {
    const res = await api.getMessage(accountId, mailbox, uid);
    return res.data;
  },
);

const mailboxSlice = createSlice({
  name: "mailbox",
  initialState,
  reducers: {
    openAccount(state, action: { payload: string }) {
      if (state.accountId !== action.payload) {
        Object.assign(state, { ...initialState, accountId: action.payload });
      }
    },
    selectMailbox(state, action: { payload: string }) {
      state.selectedMailbox = action.payload;
      state.messages = [];
      state.offset = 0;
      state.totalCount = 0;
      state.selectedMessage = null;
      state.messageLoadState = "idle";
      state.messageDetailLoadState = "idle";
    },
    clearSelectedMessage(state) {
      state.selectedMessage = null;
      state.messageDetailLoadState = "idle";
    },
    closeExplorer() {
      return initialState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMailboxes.pending, (state) => {
        state.mailboxLoadState = "loading";
        state.error = null;
      })
      .addCase(fetchMailboxes.fulfilled, (state, action) => {
        state.mailboxLoadState = "loaded";
        state.mailboxes = action.payload.mailboxes;
        state.accountId = action.payload.accountId;
      })
      .addCase(fetchMailboxes.rejected, (state, action) => {
        state.mailboxLoadState = "error";
        state.error = action.error.message ?? "Failed to load mailboxes";
      })

      .addCase(fetchMessages.pending, (state) => {
        state.messageLoadState = "loading";
        state.error = null;
      })
      .addCase(fetchMessages.fulfilled, (state, action) => {
        state.messageLoadState = "loaded";
        state.messages = action.payload.messages;
        state.totalCount = action.payload.totalCount;
        state.selectedMailbox = action.payload.mailbox;
      })
      .addCase(fetchMessages.rejected, (state, action) => {
        state.messageLoadState = "error";
        state.error = action.error.message ?? "Failed to load messages";
      })

      .addCase(fetchMessageDetail.pending, (state) => {
        state.messageDetailLoadState = "loading";
      })
      .addCase(fetchMessageDetail.fulfilled, (state, action) => {
        state.messageDetailLoadState = "loaded";
        state.selectedMessage = action.payload;
      })
      .addCase(fetchMessageDetail.rejected, (state, action) => {
        state.messageDetailLoadState = "error";
        state.error = action.error.message ?? "Failed to load message";
      });
  },
});

export const { openAccount, selectMailbox, clearSelectedMessage, closeExplorer } =
  mailboxSlice.actions;
export const mailboxReducer = mailboxSlice.reducer;

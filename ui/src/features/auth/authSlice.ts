import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api, ApiRequestError } from "../../api/client";
import type { CurrentUser } from "../../api/types";

type AuthStatus = "idle" | "loading" | "authenticated" | "unauthenticated" | "error";

interface AuthError {
  code: string;
  message: string;
  status?: number;
}

interface AuthState {
  status: AuthStatus;
  user: CurrentUser | null;
  error: AuthError | null;
}

const initialState: AuthState = {
  status: "idle",
  user: null,
  error: null,
};

export const fetchCurrentUser = createAsyncThunk<
  CurrentUser,
  void,
  { rejectValue: AuthError }
>("auth/fetchCurrentUser", async (_, { rejectWithValue }) => {
  try {
    const response = await api.getCurrentUser();
    return response.data;
  } catch (error) {
    if (error instanceof ApiRequestError) {
      return rejectWithValue({
        code: error.code,
        message: error.message,
        status: error.status,
      });
    }
    throw error;
  }
});

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    clearAuthState(state) {
      state.status = "unauthenticated";
      state.user = null;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchCurrentUser.pending, (state) => {
        state.status = "loading";
        state.error = null;
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action) => {
        state.status = "authenticated";
        state.user = action.payload;
        state.error = null;
      })
      .addCase(fetchCurrentUser.rejected, (state, action) => {
        const payload = action.payload;
        if (payload?.status === 401 || payload?.code === "unauthenticated") {
          state.status = "unauthenticated";
          state.user = null;
          state.error = null;
          return;
        }

        state.status = "error";
        state.user = null;
        state.error = payload ?? {
          code: "auth-bootstrap-failed",
          message: action.error.message ?? "Failed to load the current user.",
        };
      });
  },
});

export const { clearAuthState } = authSlice.actions;
export const authReducer = authSlice.reducer;

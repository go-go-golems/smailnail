import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

interface ReviewQueueState {
  selected: string[];
  filterTag: string | null;
  expandedId: string | null;
}

interface QueryEditorState {
  sql: string;
  activeSourcePath: string | null;
}

interface AnnotationUiState {
  reviewQueue: ReviewQueueState;
  queryEditor: QueryEditorState;
}

const initialState: AnnotationUiState = {
  reviewQueue: {
    selected: [],
    filterTag: null,
    expandedId: null,
  },
  queryEditor: {
    sql: "",
    activeSourcePath: null,
  },
};

const annotationUiSlice = createSlice({
  name: "annotationUi",
  initialState,
  reducers: {
    // ── Review queue ─────────────────────────────
    toggleSelected(state, action: PayloadAction<string>) {
      const id = action.payload;
      const idx = state.reviewQueue.selected.indexOf(id);
      if (idx >= 0) {
        state.reviewQueue.selected.splice(idx, 1);
      } else {
        state.reviewQueue.selected.push(id);
      }
    },
    setSelected(state, action: PayloadAction<string[]>) {
      state.reviewQueue.selected = action.payload;
    },
    clearSelected(state) {
      state.reviewQueue.selected = [];
    },
    setFilterTag(state, action: PayloadAction<string | null>) {
      state.reviewQueue.filterTag = action.payload;
    },
    toggleExpandedId(state, action: PayloadAction<string>) {
      state.reviewQueue.expandedId =
        state.reviewQueue.expandedId === action.payload ? null : action.payload;
    },

    // ── Query editor ─────────────────────────────
    setSql(state, action: PayloadAction<string>) {
      state.queryEditor.sql = action.payload;
    },
    setActiveSourcePath(state, action: PayloadAction<string | null>) {
      state.queryEditor.activeSourcePath = action.payload;
    },
  },
});

export const {
  toggleSelected,
  setSelected,
  clearSelected,
  setFilterTag,
  toggleExpandedId,
  setSql,
  setActiveSourcePath,
} = annotationUiSlice.actions;

export const annotationUiReducer = annotationUiSlice.reducer;

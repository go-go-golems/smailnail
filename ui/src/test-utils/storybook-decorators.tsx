import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import { theme } from "../theme/theme";
import { annotationsApi } from "../api/annotations";
import { annotationUiReducer } from "../store/annotationUiSlice";
import type { Decorator } from "@storybook/react";

/** Wraps in ThemeProvider */
export const withTheme: Decorator = (Story) => (
  <ThemeProvider theme={theme}>
    <CssBaseline />
    <Story />
  </ThemeProvider>
);

/** Wraps in a MemoryRouter at a given path */
export function withRouter(initialPath = "/"): Decorator {
  return (Story) => (
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path="*" element={<Story />} />
      </Routes>
    </MemoryRouter>
  );
}

/** Wraps in Redux Provider with a fresh store (includes RTK Query middleware for MSW) */
export const withStore: Decorator = (Story) => {
  const store = configureStore({
    reducer: {
      [annotationsApi.reducerPath]: annotationsApi.reducer,
      annotationUi: annotationUiReducer,
    },
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware().concat(annotationsApi.middleware),
  });
  return (
    <Provider store={store}>
      <Story />
    </Provider>
  );
};

/** Combined: store + router + theme */
export function withAll(initialPath = "/"): Decorator {
  return (Story) => {
    const store = configureStore({
      reducer: {
        [annotationsApi.reducerPath]: annotationsApi.reducer,
        annotationUi: annotationUiReducer,
      },
      middleware: (getDefaultMiddleware) =>
        getDefaultMiddleware().concat(annotationsApi.middleware),
    });
    return (
      <Provider store={store}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <MemoryRouter initialEntries={[initialPath]}>
            <Routes>
              <Route path="*" element={<Story />} />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </Provider>
    );
  };
}

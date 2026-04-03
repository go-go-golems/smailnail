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

/**
 * Wraps in a MemoryRouter with proper route patterns for param parsing.
 * @param initialPath - The URL to render, e.g. "/annotations/runs/run-42"
 * @param routePattern - The route pattern with params, e.g. "/annotations/runs/:runId"
 *                        Defaults to a catch-all "*" if not specified.
 */
export function withRouter(
  initialPath = "/",
  routePattern?: string,
): Decorator {
  return (Story) => (
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path={routePattern ?? "*"} element={<Story />} />
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

/**
 * Combined: store + router + theme.
 * @param initialPath - The URL to render
 * @param routePattern - Route pattern for param parsing (e.g. "/annotations/senders/:email")
 */
export function withAll(
  initialPath = "/",
  routePattern?: string,
): Decorator {
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
              <Route
                path={routePattern ?? "*"}
                element={<Story />}
              />
            </Routes>
          </MemoryRouter>
        </ThemeProvider>
      </Provider>
    );
  };
}

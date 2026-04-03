import Box from "@mui/material/Box";
import { Outlet } from "react-router-dom";
import { AnnotationSidebar } from "./AnnotationSidebar";

/**
 * Layout shell for annotation pages — sidebar + content area.
 * Renders child routes via <Outlet />.
 */
export function AnnotationLayout() {
  return (
    <Box
      data-widget="annotation-layout"
      sx={{
        display: "flex",
        height: "100vh",
        bgcolor: "background.default",
        color: "text.primary",
      }}
    >
      <AnnotationSidebar />
      <Box
        data-part="content"
        sx={{
          flex: 1,
          overflow: "auto",
          minWidth: 0,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}

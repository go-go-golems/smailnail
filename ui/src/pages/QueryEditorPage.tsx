import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export function QueryEditorPage() {
  return (
    <Box data-widget="query-editor-page" sx={{ p: 3 }}>
      <Typography variant="h2">SQL Workbench</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
        SQL query editor (coming soon)
      </Typography>
    </Box>
  );
}

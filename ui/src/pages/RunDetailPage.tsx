import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { useParams } from "react-router-dom";

export function RunDetailPage() {
  const { runId } = useParams<{ runId: string }>();

  return (
    <Box data-widget="run-detail-page" sx={{ p: 3 }}>
      <Typography variant="h2">Run: {runId}</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
        Run detail view (coming soon)
      </Typography>
    </Box>
  );
}

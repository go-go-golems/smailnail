import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export function ReviewQueuePage() {
  return (
    <Box data-widget="review-queue-page" sx={{ p: 3 }}>
      <Typography variant="h2">Review Queue</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
        Annotations pending review (coming soon)
      </Typography>
    </Box>
  );
}

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export function DashboardPage() {
  return (
    <Box data-widget="dashboard-page" sx={{ p: 3 }}>
      <Typography variant="h2">Annotations Dashboard</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
        Overview of annotation activity (coming soon)
      </Typography>
    </Box>
  );
}

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { useParams } from "react-router-dom";

export function SenderDetailPage() {
  const { email } = useParams<{ email: string }>();

  return (
    <Box data-widget="sender-detail-page" sx={{ p: 3 }}>
      <Typography variant="h2">Sender: {email}</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
        Sender detail view (coming soon)
      </Typography>
    </Box>
  );
}

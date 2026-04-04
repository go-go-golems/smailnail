import Paper from "@mui/material/Paper";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Divider from "@mui/material/Divider";
import { StatBox, TagChip } from "../shared";
import type { SenderDetail } from "../../types/annotations";

export interface SenderProfileCardProps {
  sender: SenderDetail;
}

export function SenderProfileCard({ sender }: SenderProfileCardProps) {
  const firstSeen = sender.firstSeen
    ? new Date(sender.firstSeen).toLocaleDateString("en-CA")
    : "—";
  const lastSeen = sender.lastSeen
    ? new Date(sender.lastSeen).toLocaleDateString("en-CA")
    : "—";

  return (
    <Paper data-widget="sender-profile" sx={{ mb: 2 }}>
      <Stack
        direction="row"
        spacing={0}
        divider={<Divider orientation="vertical" flexItem />}
        sx={{ p: 0.5 }}
      >
        <StatBox
          value={sender.messageCount}
          label="Messages"
          color="info.main"
        />
        <StatBox value={sender.domain} label="Domain" />
        <StatBox value={firstSeen} label="First Seen" />
        <StatBox value={lastSeen} label="Last Seen" />
        <StatBox
          value={sender.hasUnsubscribe ? "Yes" : "No"}
          label="Unsubscribe"
          color={sender.hasUnsubscribe ? "success.main" : "text.secondary"}
        />
      </Stack>
      {sender.tags.length > 0 && (
        <Box sx={{ px: 2, pb: 1.5 }}>
          <Stack direction="row" spacing={0.5}>
            {sender.tags.map((tag) => (
              <TagChip key={tag} tag={tag} />
            ))}
          </Stack>
        </Box>
      )}
    </Paper>
  );
}

import Paper from "@mui/material/Paper";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import type { TargetGroup, GroupMember } from "../../types/annotations";
import { ReviewStateBadge, SourceBadge, TargetLink } from "../shared";

export interface GroupCardProps {
  group: TargetGroup;
  members?: GroupMember[];
  onNavigateTarget?: (targetType: string, targetId: string) => void;
}

export function GroupCard({
  group,
  members = [],
  onNavigateTarget,
}: GroupCardProps) {
  return (
    <Paper data-widget="group-card" sx={{ mb: 1.5 }}>
      {/* Header */}
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          gap: 1.5,
          p: 1.5,
          borderBottom: 1,
          borderColor: "divider",
        }}
      >
        <Typography variant="body1" sx={{ fontWeight: 600, flex: 1 }}>
          {group.name}
        </Typography>
        <Chip
          label={`${members.length} members`}
          size="small"
          variant="outlined"
          sx={{ fontFamily: "monospace" }}
        />
        <ReviewStateBadge state={group.reviewState} />
        <SourceBadge
          sourceKind={group.sourceKind}
          sourceLabel={group.sourceLabel}
        />
      </Box>

      {/* Body */}
      <Box sx={{ p: 1.5 }}>
        {group.description && (
          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ mb: 1.5 }}
          >
            {group.description}
          </Typography>
        )}

        {members.length > 0 && (
          <Stack spacing={0.5}>
            {members.map((m) => (
              <TargetLink
                key={`${m.targetType}-${m.targetId}`}
                targetType={m.targetType}
                targetId={m.targetId}
                onClick={
                  onNavigateTarget
                    ? () => onNavigateTarget(m.targetType, m.targetId)
                    : undefined
                }
              />
            ))}
          </Stack>
        )}
      </Box>
    </Paper>
  );
}

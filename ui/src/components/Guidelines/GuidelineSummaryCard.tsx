import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import EditIcon from "@mui/icons-material/Edit";
import ArchiveIcon from "@mui/icons-material/Archive";
import UnarchiveIcon from "@mui/icons-material/Unarchive";
import { GuidelineScopeBadge } from "../shared";
import { parts } from "./parts";
import type { ReviewGuideline } from "../../types/reviewGuideline";

export interface GuidelineSummaryCardProps {
  guideline: ReviewGuideline;
  linkedRunCount?: number;
  onEdit: () => void;
  onArchive?: () => void;
  onActivate?: () => void;
}

const statusColor: Record<string, "success" | "default" | "warning"> = {
  active: "success",
  draft: "warning",
  archived: "default",
};

const priorityLabel = (p: number): string => {
  if (p >= 80) return "Critical";
  if (p >= 50) return "High";
  if (p >= 20) return "Medium";
  return "Low";
};

export function GuidelineSummaryCard({
  guideline,
  linkedRunCount = 0,
  onEdit,
  onArchive,
  onActivate,
}: GuidelineSummaryCardProps) {
  return (
    <Paper
      data-part={parts.guidelineSummaryCard}
      variant="outlined"
      sx={{ p: 2, mb: 1.5 }}
    >
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
        <GuidelineScopeBadge scopeKind={guideline.scopeKind} />
        <Chip
          label={guideline.status}
          size="small"
          color={statusColor[guideline.status] ?? "default"}
          variant="outlined"
        />
        <Chip
          label={`Priority: ${priorityLabel(guideline.priority)}`}
          size="small"
          variant="outlined"
        />
        {linkedRunCount > 0 && (
          <Chip
            label={`${linkedRunCount} run${linkedRunCount !== 1 ? "s" : ""}`}
            size="small"
            color="info"
            variant="outlined"
          />
        )}
        <Box sx={{ flex: 1 }} />
        {guideline.status === "archived" ? (
          onActivate && (
            <Button
              size="small"
              startIcon={<UnarchiveIcon />}
              onClick={onActivate}
            >
              Activate
            </Button>
          )
        ) : (
          onArchive && (
            <Button
              size="small"
              color="inherit"
              startIcon={<ArchiveIcon />}
              onClick={onArchive}
            >
              Archive
            </Button>
          )
        )}
        <Button size="small" startIcon={<EditIcon />} onClick={onEdit}>
          Edit
        </Button>
      </Stack>

      <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.5 }}>
        {guideline.title}
      </Typography>
      <Typography
        variant="caption"
        color="text.secondary"
        sx={{ display: "block", mb: 1 }}
      >
        {guideline.slug}
      </Typography>

      <Box
        sx={{
          overflow: "hidden",
          display: "-webkit-box",
          WebkitLineClamp: 2,
          WebkitBoxOrient: "vertical",
          fontSize: "0.8125rem",
          color: "text.secondary",
          lineHeight: 1.6,
        }}
      >
        {guideline.bodyMarkdown}
      </Box>

      <Typography
        variant="caption"
        color="text.secondary"
        sx={{ display: "block", mt: 1 }}
      >
        Created {new Date(guideline.createdAt).toLocaleDateString()} by{" "}
        {guideline.createdBy} · Updated{" "}
        {new Date(guideline.updatedAt).toLocaleDateString()}
      </Typography>
    </Paper>
  );
}

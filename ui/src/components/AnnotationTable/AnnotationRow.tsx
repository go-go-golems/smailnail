import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import Checkbox from "@mui/material/Checkbox";
import IconButton from "@mui/material/IconButton";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CancelIcon from "@mui/icons-material/Cancel";
import SpeakerNotesIcon from "@mui/icons-material/SpeakerNotes";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import type { Annotation } from "../../types/annotations";
import { TagChip, TargetLink, SourceBadge, ReviewStateBadge } from "../shared";
import { parts } from "./parts";

export interface AnnotationRowProps {
  annotation: Annotation;
  isSelected: boolean;
  isExpanded: boolean;
  onToggleSelect: () => void;
  onToggleExpand: () => void;
  onApprove: () => void;
  onDismiss: () => void;
  onDismissExplain?: () => void;
  onNavigateTarget?: () => void;
}

export function AnnotationRow({
  annotation,
  isSelected,
  isExpanded,
  onToggleSelect,
  onToggleExpand,
  onApprove,
  onDismiss,
  onDismissExplain,
  onNavigateTarget,
}: AnnotationRowProps) {
  const date = new Date(annotation.createdAt);
  const dateStr = date.toLocaleDateString("en-CA");

  return (
    <TableRow
      data-part={parts.annotationRow}
      data-state={annotation.reviewState}
      hover
      selected={isSelected}
      sx={{
        cursor: "pointer",
        "&.Mui-selected": {
          bgcolor: "rgba(245, 166, 35, 0.04)",
        },
      }}
    >
      <TableCell padding="checkbox" onClick={(e) => e.stopPropagation()}>
        <Checkbox
          checked={isSelected}
          onChange={onToggleSelect}
          size="small"
          sx={{ p: 0.5 }}
        />
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <TargetLink
          targetType={annotation.targetType}
          targetId={annotation.targetId}
          onClick={onNavigateTarget}
        />
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <TagChip tag={annotation.tag} />
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <Typography
          variant="body2"
          sx={{
            maxWidth: 280,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {annotation.noteMarkdown}
        </Typography>
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <SourceBadge
          sourceKind={annotation.sourceKind}
          sourceLabel={annotation.sourceLabel}
        />
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <ReviewStateBadge state={annotation.reviewState} />
      </TableCell>

      <TableCell onClick={onToggleExpand}>
        <Typography
          variant="caption"
          sx={{ fontFamily: "monospace", color: "text.secondary" }}
        >
          {dateStr}
        </Typography>
      </TableCell>

      <TableCell onClick={(e) => e.stopPropagation()}>
        <Stack direction="row" spacing={0.5}>
          <IconButton
            size="small"
            color="success"
            onClick={onApprove}
            title="Approve"
            disabled={annotation.reviewState === "reviewed"}
          >
            <CheckCircleIcon fontSize="small" />
          </IconButton>
          <IconButton
            size="small"
            color="error"
            onClick={onDismiss}
            title="Dismiss"
            disabled={annotation.reviewState === "dismissed"}
          >
            <CancelIcon fontSize="small" />
          </IconButton>
          {onDismissExplain && (
            <IconButton
              size="small"
              color="warning"
              onClick={onDismissExplain}
              title="Dismiss & Explain"
            >
              <SpeakerNotesIcon fontSize="small" />
            </IconButton>
          )}
          <IconButton
            size="small"
            onClick={onToggleExpand}
            title={isExpanded ? "Collapse" : "Expand"}
          >
            <ExpandMoreIcon
              fontSize="small"
              sx={{
                transform: isExpanded ? "rotate(180deg)" : "rotate(0deg)",
                transition: "transform 0.15s",
              }}
            />
          </IconButton>
        </Stack>
      </TableCell>
    </TableRow>
  );
}

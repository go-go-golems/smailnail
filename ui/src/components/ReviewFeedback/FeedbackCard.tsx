import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Chip from "@mui/material/Chip";
import {
  FeedbackKindBadge,
  FeedbackStatusBadge,
  MarkdownRenderer,
} from "../shared";
import type { ReviewFeedback } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface FeedbackCardProps {
  feedback: ReviewFeedback;
  onAcknowledge?: () => void;
  onResolve?: () => void;
  compact?: boolean;
}

export function FeedbackCard({
  feedback,
  onAcknowledge,
  onResolve,
  compact = false,
}: FeedbackCardProps) {
  const created = new Date(feedback.createdAt).toLocaleString();

  return (
    <Paper
      data-part={parts.feedbackCard}
      sx={{ p: 1.5, mb: 1 }}
    >
      {/* Header row: kind badge, status badge, author, date */}
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
        <FeedbackKindBadge kind={feedback.feedbackKind} />
        <FeedbackStatusBadge status={feedback.status} />
        <Typography variant="caption" color="text.secondary">
          {feedback.createdBy}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {created}
        </Typography>
      </Stack>

      {/* Title */}
      <Typography variant="body2" sx={{ fontWeight: 600, mb: 0.5 }}>
        {feedback.title}
      </Typography>

      {/* Body */}
      {!compact && feedback.bodyMarkdown && (
        <Box sx={{ mb: 1 }}>
          <MarkdownRenderer content={feedback.bodyMarkdown} />
        </Box>
      )}

      {/* Target count */}
      {feedback.targets.length > 0 && (
        <Chip
          label={`${feedback.targets.length} target${feedback.targets.length > 1 ? "s" : ""}`}
          size="small"
          variant="outlined"
          sx={{ mr: 1, fontSize: "0.6875rem" }}
        />
      )}

      {/* Action buttons */}
      {feedback.status === "open" && (
        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
          {onAcknowledge && (
            <Button size="small" variant="outlined" onClick={onAcknowledge}>
              Acknowledge
            </Button>
          )}
          {onResolve && (
            <Button
              size="small"
              variant="outlined"
              color="success"
              onClick={onResolve}
            >
              Resolve
            </Button>
          )}
        </Stack>
      )}

      {feedback.status === "acknowledged" && onResolve && (
        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
          <Button
            size="small"
            variant="outlined"
            color="success"
            onClick={onResolve}
          >
            Resolve
          </Button>
        </Stack>
      )}
    </Paper>
  );
}

import { useState } from "react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Button from "@mui/material/Button";
import Collapse from "@mui/material/Collapse";
import SendIcon from "@mui/icons-material/Send";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import { GuidelinePicker } from "./GuidelinePicker";
import type { FeedbackKind } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface ReviewCommentInlineProps {
  open: boolean;
  annotationId: string;
  agentRunId: string;
  mailboxName?: string;
  onJustDismiss: () => void;
  onSubmit: (payload: {
    feedbackKind: FeedbackKind;
    title: string;
    bodyMarkdown: string;
    guidelineIds: string[];
  }) => void;
  onCancel: () => void;
}

const FEEDBACK_KINDS: { value: FeedbackKind; label: string }[] = [
  { value: "reject_request", label: "Reject Request" },
  { value: "comment", label: "Comment" },
  { value: "clarification", label: "Clarification" },
];

export function ReviewCommentInline({
  open,
  onJustDismiss,
  onSubmit,
  onCancel,
}: ReviewCommentInlineProps) {
  const [feedbackKind, setFeedbackKind] = useState<FeedbackKind>("reject_request");
  const [title, setTitle] = useState("");
  const [bodyMarkdown, setBodyMarkdown] = useState("");
  const [guidelineIds, setGuidelineIds] = useState<string[]>([]);
  const [guidelinesExpanded, setGuidelinesExpanded] = useState(false);

  const handleSubmit = () => {
    if (!title.trim()) return;
    onSubmit({ feedbackKind, title, bodyMarkdown, guidelineIds });
    setTitle("");
    setBodyMarkdown("");
    setGuidelineIds([]);
    setGuidelinesExpanded(false);
  };

  const toggleGuideline = (id: string) => {
    setGuidelineIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  return (
    <Collapse in={open}>
      <Box
        data-part={parts.feedbackForm}
        sx={{
          p: 1.5,
          mt: 1,
          border: 1,
          borderColor: "divider",
          borderRadius: 1,
          bgcolor: "background.paper",
        }}
      >
        <Stack spacing={1.5}>
          <TextField
            select
            label="Kind"
            value={feedbackKind}
            onChange={(e) => setFeedbackKind(e.target.value as FeedbackKind)}
            size="small"
            fullWidth
          >
            {FEEDBACK_KINDS.map((k) => (
              <MenuItem key={k.value} value={k.value}>
                {k.label}
              </MenuItem>
            ))}
          </TextField>

          <TextField
            label="Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            size="small"
            fullWidth
            required
          />

          <TextField
            label="Note"
            value={bodyMarkdown}
            onChange={(e) => setBodyMarkdown(e.target.value)}
            size="small"
            fullWidth
            multiline
            minRows={2}
            maxRows={6}
          />

          <Box>
            <Button
              size="small"
              variant="text"
              startIcon={
                guidelinesExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />
              }
              onClick={() => setGuidelinesExpanded(!guidelinesExpanded)}
            >
              Attach Guideline
              {guidelineIds.length > 0 && ` (${guidelineIds.length})`}
            </Button>
            <Collapse in={guidelinesExpanded}>
              <Box sx={{ mt: 0.5 }}>
                <GuidelinePicker
                  selectedIds={guidelineIds}
                  onToggle={toggleGuideline}
                />
              </Box>
            </Collapse>
          </Box>

          <Stack direction="row" spacing={1} justifyContent="flex-end">
            <Button size="small" variant="outlined" onClick={onCancel}>
              Cancel
            </Button>
            <Button size="small" variant="text" onClick={onJustDismiss}>
              Just Dismiss
            </Button>
            <Button
              size="small"
              variant="contained"
              color="error"
              startIcon={<SendIcon />}
              disabled={!title.trim()}
              onClick={handleSubmit}
            >
              Dismiss & Explain
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Collapse>
  );
}

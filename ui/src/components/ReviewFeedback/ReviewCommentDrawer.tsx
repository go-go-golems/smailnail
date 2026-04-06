import { useEffect, useState } from "react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import Collapse from "@mui/material/Collapse";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import CancelIcon from "@mui/icons-material/Cancel";
import SendIcon from "@mui/icons-material/Send";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import { MailboxBadge } from "../shared";
import { GuidelinePicker } from "./GuidelinePicker";
import type { FeedbackKind } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface ReviewCommentDrawerProps {
  open: boolean;
  mode: "single" | "batch" | "run";
  targetCount: number;
  agentRunId?: string;
  mailboxName?: string;
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
  { value: "guideline_request", label: "Guideline Request" },
  { value: "clarification", label: "Clarification" },
];

export function ReviewCommentDrawer({
  open,
  mode,
  targetCount,
  mailboxName,
  onSubmit,
  onCancel,
}: ReviewCommentDrawerProps) {
  const [feedbackKind, setFeedbackKind] = useState<FeedbackKind>(
    mode === "run" ? "comment" : "reject_request",
  );
  const [title, setTitle] = useState("");
  const [bodyMarkdown, setBodyMarkdown] = useState("");
  const [guidelineIds, setGuidelineIds] = useState<string[]>([]);
  const [guidelinesExpanded, setGuidelinesExpanded] = useState(false);

  const resetForm = () => {
    setFeedbackKind(mode === "run" ? "comment" : "reject_request");
    setTitle("");
    setBodyMarkdown("");
    setGuidelineIds([]);
    setGuidelinesExpanded(false);
  };

  useEffect(() => {
    if (open) {
      resetForm();
    }
  }, [open, mode]);

  const handleSubmit = () => {
    if (!title.trim()) return;
    onSubmit({ feedbackKind, title, bodyMarkdown, guidelineIds });
    resetForm();
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  const submitLabel =
    mode === "batch"
      ? `Reject ${targetCount} Items`
      : mode === "single"
        ? "Dismiss & Explain"
        : "Submit Feedback";

  const headerLabel =
    mode === "batch"
      ? `Reject & Explain (${targetCount} items)`
      : mode === "single"
        ? "Dismiss with Feedback"
        : "Add Run Feedback";

  const toggleGuideline = (id: string) => {
    setGuidelineIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  return (
    <Dialog
      open={open}
      onClose={handleCancel}
      fullWidth
      maxWidth="sm"
      scroll="paper"
    >
      <DialogTitle>{headerLabel}</DialogTitle>
      <DialogContent dividers>
        <Box data-part={parts.commentDrawer}>
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
              autoFocus
            />

            <TextField
              label="Explanation"
              value={bodyMarkdown}
              onChange={(e) => setBodyMarkdown(e.target.value)}
              size="small"
              fullWidth
              multiline
              minRows={3}
              maxRows={8}
              placeholder="Describe what was wrong or what should change..."
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
                Attach Guidelines
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

            {mailboxName && (
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <Typography variant="caption" color="text.secondary">
                  Mailbox:
                </Typography>
                <MailboxBadge mailboxName={mailboxName} variant="inline" />
              </Box>
            )}

            <Divider />
          </Stack>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button
          size="small"
          variant="outlined"
          startIcon={<CancelIcon />}
          onClick={handleCancel}
        >
          Cancel
        </Button>
        <Button
          size="small"
          variant="contained"
          color="error"
          startIcon={<SendIcon />}
          disabled={!title.trim()}
          onClick={handleSubmit}
        >
          {submitLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

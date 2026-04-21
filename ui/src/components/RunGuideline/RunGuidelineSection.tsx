import { useState } from "react";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import AddIcon from "@mui/icons-material/Add";
import PostAddIcon from "@mui/icons-material/PostAdd";
import {
  useLinkGuidelineToRunMutation,
  useUnlinkGuidelineFromRunMutation,
} from "../../api/annotations";
import { GuidelineCard } from "./GuidelineCard";
import { GuidelineLinkPicker } from "../ReviewFeedback";
import type { ReviewGuideline } from "../../types/reviewGuideline";
import { parts } from "./parts";

export interface RunGuidelineSectionProps {
  runId: string;
  guidelines: ReviewGuideline[];
  onCreateAndLink: () => void;
}

export function RunGuidelineSection({
  runId,
  guidelines,
  onCreateAndLink,
}: RunGuidelineSectionProps) {
  const [pickerOpen, setPickerOpen] = useState(false);
  const [linkError, setLinkError] = useState<string | null>(null);
  const [linkGuidelineToRun] = useLinkGuidelineToRunMutation();
  const [unlinkGuidelineFromRun] = useUnlinkGuidelineFromRunMutation();

  const handleLink = async (guidelineIds: string[]) => {
    setLinkError(null);
    try {
      await Promise.all(
        guidelineIds.map((id) =>
          linkGuidelineToRun({ runId, guidelineId: id }).unwrap(),
        ),
      );
      setPickerOpen(false);
    } catch {
      setLinkError("Failed to link one or more guidelines to this run. Please try again.");
      throw new Error("guideline link failed");
    }
  };

  const handleUnlink = async (guidelineId: string) => {
    setLinkError(null);
    try {
      await unlinkGuidelineFromRun({ runId, guidelineId }).unwrap();
    } catch {
      setLinkError("Failed to unlink the guideline from this run. Please try again.");
    }
  };

  return (
    <Box data-part={parts.runGuidelineSection}>
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Linked Guidelines ({guidelines.length})
      </Typography>

      {linkError && (
        <Alert severity="error" sx={{ mb: 1.5 }}>
          {linkError}
        </Alert>
      )}

      {guidelines.length === 0 && (
        <Box
          sx={{
            textAlign: "center",
            py: 2,
            color: "text.secondary",
            border: 1,
            borderColor: "divider",
            borderRadius: 1,
            mb: 1,
          }}
        >
          <Typography variant="body2">
            No guidelines linked to this run.
          </Typography>
        </Box>
      )}

      {guidelines.map((g) => (
        <GuidelineCard
          key={g.id}
          guideline={g}
          onUnlink={() => void handleUnlink(g.id)}
        />
      ))}

      <Box sx={{ mt: 1, display: "flex", gap: 1 }}>
        <Button
          size="small"
          variant="outlined"
          startIcon={<AddIcon />}
          onClick={() => {
            setLinkError(null);
            setPickerOpen(true);
          }}
        >
          Link Existing Guideline
        </Button>
        <Button
          size="small"
          variant="text"
          startIcon={<PostAddIcon />}
          onClick={onCreateAndLink}
        >
          Create New Guideline for This Run
        </Button>
      </Box>

      <GuidelineLinkPicker
        open={pickerOpen}
        runId={runId}
        alreadyLinkedIds={guidelines.map((g) => g.id)}
        onLink={handleLink}
        onClose={() => setPickerOpen(false)}
      />
    </Box>
  );
}

import { useState } from "react";
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
  const [linkGuidelineToRun] = useLinkGuidelineToRunMutation();
  const [unlinkGuidelineFromRun] = useUnlinkGuidelineFromRunMutation();

  const handleLink = (guidelineIds: string[]) => {
    for (const id of guidelineIds) {
      void linkGuidelineToRun({ runId, guidelineId: id });
    }
    setPickerOpen(false);
  };

  const handleUnlink = (guidelineId: string) => {
    void unlinkGuidelineFromRun({ runId, guidelineId });
  };

  return (
    <Box data-part={parts.runGuidelineSection}>
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Linked Guidelines ({guidelines.length})
      </Typography>

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
          onUnlink={() => handleUnlink(g.id)}
        />
      ))}

      <Box sx={{ mt: 1, display: "flex", gap: 1 }}>
        <Button
          size="small"
          variant="outlined"
          startIcon={<AddIcon />}
          onClick={() => setPickerOpen(true)}
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

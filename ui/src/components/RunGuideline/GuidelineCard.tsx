import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import LinkIcon from "@mui/icons-material/Link";
import { GuidelineScopeBadge, MarkdownRenderer } from "../shared";
import type { ReviewGuideline } from "../../types/reviewGuideline";
import { parts } from "./parts";

export interface GuidelineCardProps {
  guideline: ReviewGuideline;
  onUnlink?: () => void;
  compact?: boolean;
}

export function GuidelineCard({
  guideline,
  onUnlink,
  compact = false,
}: GuidelineCardProps) {
  return (
    <Paper
      data-part={parts.guidelineCard}
      variant="outlined"
      sx={{ p: 1.5, mb: 1 }}
    >
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
        <LinkIcon fontSize="small" color="action" />
        <Typography variant="body2" sx={{ fontWeight: 600, flex: 1 }} noWrap>
          {guideline.slug}
        </Typography>
        <GuidelineScopeBadge scopeKind={guideline.scopeKind} />
        <Typography
          variant="caption"
          sx={{
            fontFamily: "monospace",
            color:
              guideline.status === "active"
                ? "success.main"
                : guideline.status === "draft"
                  ? "warning.main"
                  : "text.disabled",
          }}
        >
          {guideline.status}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          pri: {guideline.priority}
        </Typography>
      </Stack>

      {!compact && (
        <>
          <Typography variant="body2" sx={{ mb: 0.5 }} noWrap>
            {guideline.title}
          </Typography>
          <Box
            sx={{
              maxHeight: 60,
              overflow: "hidden",
              mb: 0.5,
              "& p": { m: 0, fontSize: "0.75rem" },
            }}
          >
            <MarkdownRenderer content={guideline.bodyMarkdown} />
          </Box>
        </>
      )}

      {compact && (
        <Typography variant="caption" color="text.secondary" noWrap>
          {guideline.title}
        </Typography>
      )}

      {onUnlink && (
        <Box sx={{ mt: 1, textAlign: "right" }}>
          <Button size="small" variant="text" color="error" onClick={onUnlink}>
            Unlink
          </Button>
        </Box>
      )}
    </Paper>
  );
}

import Box from "@mui/material/Box";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import Collapse from "@mui/material/Collapse";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import type { Annotation } from "../../types/annotations";
import type { ReviewFeedback } from "../../types/reviewFeedback";
import type { ReviewGuideline } from "../../types/reviewGuideline";
import { FeedbackCard } from "../ReviewFeedback";
import { GuidelineCard } from "../RunGuideline/GuidelineCard";
import { MarkdownRenderer, TagChip, SourceBadge, ReviewStateBadge } from "../shared";
import { parts } from "./parts";

export interface AnnotationDetailProps {
  annotation: Annotation;
  isExpanded: boolean;
  /** Other annotations on the same target */
  relatedAnnotations?: Annotation[];
  /** Review feedback attached to this annotation */
  feedback?: ReviewFeedback[];
  /** Guidelines relevant to this annotation, typically via its run */
  guidelines?: ReviewGuideline[];
  onNavigateTarget?: () => void;
  columnCount: number;
}

export function AnnotationDetail({
  annotation,
  isExpanded,
  relatedAnnotations = [],
  feedback = [],
  guidelines = [],
  onNavigateTarget,
  columnCount,
}: AnnotationDetailProps) {
  const created = new Date(annotation.createdAt).toLocaleString();
  const updated = new Date(annotation.updatedAt).toLocaleString();

  return (
    <TableRow data-part={parts.annotationDetail}>
      <TableCell
        colSpan={columnCount}
        sx={{ p: 0, borderBottom: isExpanded ? undefined : "none" }}
      >
        <Collapse in={isExpanded}>
          <Box sx={{ p: 2, bgcolor: "background.default" }}>
            {/* Header info */}
            <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 1.5 }}>
              <TagChip tag={annotation.tag} />
              <SourceBadge
                sourceKind={annotation.sourceKind}
                sourceLabel={annotation.sourceLabel}
              />
              <ReviewStateBadge state={annotation.reviewState} />
              <Typography variant="caption" color="text.secondary">
                Created: {created}
              </Typography>
              {annotation.createdAt !== annotation.updatedAt && (
                <Typography variant="caption" color="text.secondary">
                  Updated: {updated}
                </Typography>
              )}
            </Stack>

            {/* Full note */}
            <Box
              sx={{
                p: 1.5,
                bgcolor: "background.paper",
                borderRadius: 1,
                border: 1,
                borderColor: "divider",
                mb: 1.5,
              }}
            >
              <MarkdownRenderer content={annotation.noteMarkdown} />
            </Box>

            {/* Navigate to target */}
            {onNavigateTarget && (
              <Button
                size="small"
                variant="outlined"
                startIcon={<OpenInNewIcon />}
                onClick={onNavigateTarget}
                sx={{ mb: 1.5 }}
              >
                View {annotation.targetType}: {annotation.targetId}
              </Button>
            )}

            {/* Feedback attached to this annotation */}
            {feedback.length > 0 && (
              <>
                <Divider sx={{ my: 1.5 }} />
                <Typography variant="overline" sx={{ display: "block", mb: 1 }}>
                  Review feedback ({feedback.length})
                </Typography>
                <Box sx={{ mb: relatedAnnotations.length > 0 ? 1.5 : 0 }}>
                  {feedback.map((item) => (
                    <FeedbackCard key={item.id} feedback={item} />
                  ))}
                </Box>
              </>
            )}

            {/* Linked guidelines relevant to this annotation */}
            {guidelines.length > 0 && (
              <>
                <Divider sx={{ my: 1.5 }} />
                <Typography variant="overline" sx={{ display: "block", mb: 1 }}>
                  Linked guidelines for this run ({guidelines.length})
                </Typography>
                <Box sx={{ mb: relatedAnnotations.length > 0 ? 1.5 : 0 }}>
                  {guidelines.map((guideline) => (
                    <GuidelineCard key={guideline.id} guideline={guideline} compact />
                  ))}
                </Box>
              </>
            )}

            {/* Related annotations on same target */}
            {relatedAnnotations.length > 0 && (
              <>
                <Divider sx={{ my: 1.5 }} />
                <Typography variant="overline" sx={{ display: "block", mb: 1 }}>
                  Other annotations on this target ({relatedAnnotations.length})
                </Typography>
                <Stack spacing={0.5}>
                  {relatedAnnotations.map((rel) => (
                    <Stack
                      key={rel.id}
                      direction="row"
                      spacing={1}
                      alignItems="center"
                      sx={{
                        p: 0.75,
                        borderRadius: 0.5,
                        bgcolor: "action.hover",
                      }}
                    >
                      <TagChip tag={rel.tag} />
                      <Typography
                        variant="body2"
                        sx={{
                          flex: 1,
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {rel.noteMarkdown}
                      </Typography>
                      <ReviewStateBadge state={rel.reviewState} />
                    </Stack>
                  ))}
                </Stack>
              </>
            )}
          </Box>
        </Collapse>
      </TableCell>
    </TableRow>
  );
}

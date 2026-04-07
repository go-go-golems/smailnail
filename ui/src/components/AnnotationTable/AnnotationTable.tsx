import { Fragment, memo, useMemo } from "react";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Checkbox from "@mui/material/Checkbox";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import type { Annotation } from "../../types/annotations";
import type { ReviewFeedback } from "../../types/reviewFeedback";
import { AnnotationRow } from "./AnnotationRow";
import { AnnotationDetail } from "./AnnotationDetail";
import { parts } from "./parts";

export interface AnnotationTableProps {
  annotations: Annotation[];
  /** IDs of selected annotations */
  selected: string[];
  /** ID of the expanded annotation, or null */
  expandedId: string | null;
  /** Toggle selection of an annotation */
  onToggleSelect: (id: string) => void;
  /** Toggle all selections */
  onToggleAll: () => void;
  /** Expand/collapse an annotation detail */
  onToggleExpand: (id: string) => void;
  /** Approve a single annotation */
  onApprove: (id: string) => void;
  /** Dismiss a single annotation */
  onDismiss: (id: string) => void;
  /** Dismiss a single annotation with a note */
  onDismissExplain?: (id: string) => void;
  /** Navigate to an annotation's target */
  onNavigateTarget?: (targetType: string, targetId: string) => void;
  /** Find related annotations for the expanded row */
  getRelated?: (annotation: Annotation) => Annotation[];
  /** Find feedback for the expanded row */
  getFeedback?: (annotation: Annotation) => ReviewFeedback[];
}

interface AnnotationTableItemProps {
  annotation: Annotation;
  isSelected: boolean;
  isExpanded: boolean;
  relatedAnnotations: Annotation[];
  feedback: ReviewFeedback[];
  onToggleSelect: (id: string) => void;
  onToggleExpand: (id: string) => void;
  onApprove: (id: string) => void;
  onDismiss: (id: string) => void;
  onDismissExplain?: (id: string) => void;
  onNavigateTarget?: (targetType: string, targetId: string) => void;
  columnCount: number;
}

const COLUMN_COUNT = 8;
const EMPTY_RELATED: Annotation[] = [];
const EMPTY_FEEDBACK: ReviewFeedback[] = [];

const AnnotationTableItem = memo(
  function AnnotationTableItem({
    annotation,
    isSelected,
    isExpanded,
    relatedAnnotations,
    feedback,
    onToggleSelect,
    onToggleExpand,
    onApprove,
    onDismiss,
    onDismissExplain,
    onNavigateTarget,
    columnCount,
  }: AnnotationTableItemProps) {
    return (
      <Fragment>
        <AnnotationRow
          annotation={annotation}
          isSelected={isSelected}
          isExpanded={isExpanded}
          onToggleSelect={() => onToggleSelect(annotation.id)}
          onToggleExpand={() => onToggleExpand(annotation.id)}
          onApprove={() => onApprove(annotation.id)}
          onDismiss={() => onDismiss(annotation.id)}
          onDismissExplain={
            onDismissExplain ? () => onDismissExplain(annotation.id) : undefined
          }
          onNavigateTarget={
            onNavigateTarget
              ? () => onNavigateTarget(annotation.targetType, annotation.targetId)
              : undefined
          }
        />
        {isExpanded ? (
          <AnnotationDetail
            annotation={annotation}
            isExpanded={true}
            relatedAnnotations={relatedAnnotations}
            feedback={feedback}
            onNavigateTarget={
              onNavigateTarget
                ? () => onNavigateTarget(annotation.targetType, annotation.targetId)
                : undefined
            }
            columnCount={columnCount}
          />
        ) : null}
      </Fragment>
    );
  },
  (prev, next) =>
    prev.annotation === next.annotation &&
    prev.isSelected === next.isSelected &&
    prev.isExpanded === next.isExpanded &&
    prev.relatedAnnotations === next.relatedAnnotations &&
    prev.feedback === next.feedback &&
    prev.onToggleSelect === next.onToggleSelect &&
    prev.onToggleExpand === next.onToggleExpand &&
    prev.onApprove === next.onApprove &&
    prev.onDismiss === next.onDismiss &&
    prev.onDismissExplain === next.onDismissExplain &&
    prev.onNavigateTarget === next.onNavigateTarget &&
    prev.columnCount === next.columnCount,
);

export function AnnotationTable({
  annotations,
  selected,
  expandedId,
  onToggleSelect,
  onToggleAll,
  onToggleExpand,
  onApprove,
  onDismiss,
  onDismissExplain,
  onNavigateTarget,
  getRelated,
  getFeedback,
}: AnnotationTableProps) {
  const selectedSet = useMemo(() => new Set(selected), [selected]);
  const allSelected =
    annotations.length > 0 && selected.length === annotations.length;
  const someSelected = selected.length > 0 && !allSelected;

  const expandedAnnotation = useMemo(
    () => annotations.find((annotation) => annotation.id === expandedId) ?? null,
    [annotations, expandedId],
  );

  const expandedRelated = useMemo(() => {
    if (!expandedAnnotation || !getRelated) {
      return EMPTY_RELATED;
    }
    return getRelated(expandedAnnotation);
  }, [expandedAnnotation, getRelated]);

  const expandedFeedback = useMemo(() => {
    if (!expandedAnnotation || !getFeedback) {
      return EMPTY_FEEDBACK;
    }
    return getFeedback(expandedAnnotation);
  }, [expandedAnnotation, getFeedback]);

  if (annotations.length === 0) {
    return (
      <Box
        data-widget={parts.annotationTable}
        data-state="empty"
        sx={{ textAlign: "center", py: 6, color: "text.secondary" }}
      >
        <Typography variant="h4" sx={{ mb: 1 }}>
          No annotations
        </Typography>
        <Typography variant="body2">
          Nothing to show with the current filters.
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer
      component={Paper}
      data-widget={parts.annotationTable}
      sx={{ border: "none" }}
    >
      <Table size="small" stickyHeader>
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                checked={allSelected}
                indeterminate={someSelected}
                onChange={onToggleAll}
                size="small"
                sx={{ p: 0.5 }}
              />
            </TableCell>
            <TableCell>Target</TableCell>
            <TableCell>Tag</TableCell>
            <TableCell>Note</TableCell>
            <TableCell>Source</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Actions</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {annotations.map((annotation) => {
            const isExpanded = expandedId === annotation.id;

            return (
              <AnnotationTableItem
                key={annotation.id}
                annotation={annotation}
                isSelected={selectedSet.has(annotation.id)}
                isExpanded={isExpanded}
                relatedAnnotations={isExpanded ? expandedRelated : EMPTY_RELATED}
                feedback={isExpanded ? expandedFeedback : EMPTY_FEEDBACK}
                onToggleSelect={onToggleSelect}
                onToggleExpand={onToggleExpand}
                onApprove={onApprove}
                onDismiss={onDismiss}
                onDismissExplain={onDismissExplain}
                onNavigateTarget={onNavigateTarget}
                columnCount={COLUMN_COUNT}
              />
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

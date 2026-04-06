import { Fragment } from "react";
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
  /** Navigate to an annotation's target */
  onNavigateTarget?: (targetType: string, targetId: string) => void;
  /** Find related annotations for the expanded row */
  getRelated?: (annotation: Annotation) => Annotation[];
}

const COLUMN_COUNT = 8;

export function AnnotationTable({
  annotations,
  selected,
  expandedId,
  onToggleSelect,
  onToggleAll,
  onToggleExpand,
  onApprove,
  onDismiss,
  onNavigateTarget,
  getRelated,
}: AnnotationTableProps) {
  const allSelected =
    annotations.length > 0 && selected.length === annotations.length;
  const someSelected = selected.length > 0 && !allSelected;

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
          {annotations.map((ann) => {
            const isSelected = selected.includes(ann.id);
            const isExpanded = expandedId === ann.id;

            return (
              <Fragment key={ann.id}>
                <AnnotationRow
                  annotation={ann}
                  isSelected={isSelected}
                  isExpanded={isExpanded}
                  onToggleSelect={() => onToggleSelect(ann.id)}
                  onToggleExpand={() => onToggleExpand(ann.id)}
                  onApprove={() => onApprove(ann.id)}
                  onDismiss={() => onDismiss(ann.id)}
                  onNavigateTarget={
                    onNavigateTarget
                      ? () =>
                          onNavigateTarget(ann.targetType, ann.targetId)
                      : undefined
                  }
                />
                <AnnotationDetail
                  annotation={ann}
                  isExpanded={isExpanded}
                  relatedAnnotations={
                    isExpanded ? (getRelated?.(ann) ?? []) : []
                  }
                  onNavigateTarget={
                    onNavigateTarget
                      ? () =>
                          onNavigateTarget(ann.targetType, ann.targetId)
                      : undefined
                  }
                  columnCount={COLUMN_COUNT}
                />
              </Fragment>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

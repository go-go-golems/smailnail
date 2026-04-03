import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CancelIcon from "@mui/icons-material/Cancel";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import { parts } from "./parts";

export interface BatchActionBarProps {
  /** Total number of items in the list */
  totalCount: number;
  /** Number of currently selected items */
  selectedCount: number;
  /** Whether all items are selected */
  allSelected: boolean;
  /** Toggle select-all */
  onToggleAll: () => void;
  /** Approve selected items */
  onApprove: () => void;
  /** Dismiss selected items */
  onDismiss: () => void;
  /** Reset selected items to to_review */
  onReset?: () => void;
}

export function BatchActionBar({
  totalCount,
  selectedCount,
  allSelected,
  onToggleAll,
  onApprove,
  onDismiss,
  onReset,
}: BatchActionBarProps) {
  const hasSelection = selectedCount > 0;

  return (
    <Box
      data-part={parts.batchBar}
      sx={{
        display: "flex",
        alignItems: "center",
        gap: 1.5,
        py: 1,
        px: 1.5,
        borderBottom: 1,
        borderColor: "divider",
        bgcolor: hasSelection ? "rgba(245, 166, 35, 0.04)" : "transparent",
        transition: "background-color 0.15s",
      }}
    >
      <Checkbox
        checked={allSelected}
        indeterminate={hasSelection && !allSelected}
        onChange={onToggleAll}
        size="small"
        sx={{ p: 0.5 }}
      />

      <Typography variant="body2" sx={{ color: "text.secondary", minWidth: 100 }}>
        {hasSelection
          ? `${selectedCount} of ${totalCount} selected`
          : `${totalCount} items`}
      </Typography>

      <Stack direction="row" spacing={1} sx={{ ml: "auto" }}>
        <Button
          size="small"
          variant="outlined"
          color="success"
          startIcon={<CheckCircleIcon />}
          disabled={!hasSelection}
          onClick={onApprove}
        >
          Approve
        </Button>
        <Button
          size="small"
          variant="outlined"
          color="error"
          startIcon={<CancelIcon />}
          disabled={!hasSelection}
          onClick={onDismiss}
        >
          Dismiss
        </Button>
        {onReset && (
          <Button
            size="small"
            variant="outlined"
            color="inherit"
            startIcon={<RestartAltIcon />}
            disabled={!hasSelection}
            onClick={onReset}
          >
            Reset
          </Button>
        )}
      </Stack>
    </Box>
  );
}

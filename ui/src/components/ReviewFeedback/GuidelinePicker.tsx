import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Checkbox from "@mui/material/Checkbox";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import {
  useListGuidelinesQuery,
} from "../../api/annotations";
import { parts } from "./parts";

export interface GuidelinePickerProps {
  /** Currently selected guideline IDs */
  selectedIds: string[];
  /** Toggle a guideline in/out of selection */
  onToggle: (id: string) => void;
}

export function GuidelinePicker({ selectedIds, onToggle }: GuidelinePickerProps) {
  const { data: guidelines = [], isLoading } = useListGuidelinesQuery({
    status: "active",
  });

  if (isLoading) {
    return (
      <Box sx={{ py: 1, textAlign: "center" }}>
        <CircularProgress size={20} />
      </Box>
    );
  }

  if (guidelines.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
        No active guidelines available.
      </Typography>
    );
  }

  return (
    <Box data-part={parts.guidelinePicker}>
      {guidelines.map((g) => {
        const isSelected = selectedIds.includes(g.id);
        return (
          <Stack
            key={g.id}
            direction="row"
            spacing={1}
            alignItems="center"
            sx={{
              py: 0.5,
              px: 1,
              borderRadius: 0.5,
              bgcolor: isSelected ? "action.selected" : "transparent",
              cursor: "pointer",
              "&:hover": { bgcolor: "action.hover" },
            }}
            onClick={() => onToggle(g.id)}
          >
            <Checkbox checked={isSelected} size="small" sx={{ p: 0.25 }} />
            <Typography variant="body2" sx={{ flex: 1 }}>
              {g.slug}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              pri {g.priority}
            </Typography>
          </Stack>
        );
      })}
    </Box>
  );
}

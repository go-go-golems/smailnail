import { useState, useMemo } from "react";
import Box from "@mui/material/Box";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Checkbox from "@mui/material/Checkbox";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import { useListGuidelinesQuery } from "../../api/annotations";
import { GuidelineScopeBadge } from "../shared";
import type { ReviewGuideline } from "../../types/reviewGuideline";
import { parts } from "./parts";

export interface GuidelineLinkPickerProps {
  open: boolean;
  runId: string;
  alreadyLinkedIds: string[];
  onLink: (guidelineIds: string[]) => Promise<void> | void;
  onClose: () => void;
}

export function GuidelineLinkPicker({
  open,
  alreadyLinkedIds,
  onLink,
  onClose,
}: GuidelineLinkPickerProps) {
  const { data: guidelines = [], isLoading } = useListGuidelinesQuery({
    status: "active",
  });
  const [search, setSearch] = useState("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const filtered = useMemo(() => {
    if (!search) return guidelines;
    const q = search.toLowerCase();
    return guidelines.filter(
      (g: ReviewGuideline) =>
        g.title.toLowerCase().includes(q) || g.slug.toLowerCase().includes(q),
    );
  }, [guidelines, search]);

  const toggleId = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  const handleLink = async () => {
    setIsSubmitting(true);
    try {
      await onLink(selectedIds);
      setSelectedIds([]);
      setSearch("");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    setSelectedIds([]);
    setSearch("");
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      data-part={parts.guidelineLinkPicker}
    >
      <DialogTitle>Link Guideline to Run</DialogTitle>
      <DialogContent>
        <TextField
          label="Search guidelines"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          size="small"
          fullWidth
          sx={{ mb: 2 }}
        />

        {isLoading && (
          <Box sx={{ textAlign: "center", py: 2 }}>
            <CircularProgress size={24} />
          </Box>
        )}

        {!isLoading && filtered.length === 0 && (
          <Typography variant="body2" color="text.secondary">
            No guidelines found.
          </Typography>
        )}

        {filtered.map((g: ReviewGuideline) => {
          const isLinked = alreadyLinkedIds.includes(g.id);
          const isSelected = selectedIds.includes(g.id);

          return (
            <Stack
              key={g.id}
              direction="row"
              spacing={1}
              alignItems="center"
              sx={{
                py: 0.75,
                px: 1,
                borderRadius: 0.5,
                opacity: isLinked ? 0.5 : 1,
                bgcolor: isSelected ? "action.selected" : "transparent",
                cursor: isLinked ? "default" : "pointer",
                "&:hover": { bgcolor: "action.hover" },
              }}
              onClick={() => !isLinked && toggleId(g.id)}
            >
              <Checkbox
                checked={isSelected || isLinked}
                disabled={isLinked}
                size="small"
                sx={{ p: 0.25 }}
              />
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography variant="body2" noWrap>
                  {g.slug}
                </Typography>
                <Typography variant="caption" color="text.secondary" noWrap>
                  {g.title}
                </Typography>
              </Box>
              <GuidelineScopeBadge scopeKind={g.scopeKind} />
              {isLinked && (
                <Typography variant="caption" color="text.secondary">
                  linked
                </Typography>
              )}
            </Stack>
          );
        })}
      </DialogContent>
      <DialogActions>
        <Button size="small" onClick={handleClose} disabled={isSubmitting}>
          Cancel
        </Button>
        <Button
          size="small"
          variant="contained"
          disabled={selectedIds.length === 0 || isSubmitting}
          onClick={() => void handleLink()}
        >
          Link {selectedIds.length > 0 ? `${selectedIds.length} Guideline${selectedIds.length > 1 ? "s" : ""}` : "Guideline"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

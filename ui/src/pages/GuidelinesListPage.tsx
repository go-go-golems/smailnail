import { useState, useMemo } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Chip from "@mui/material/Chip";
import AddIcon from "@mui/icons-material/Add";
import { useNavigate } from "react-router-dom";
import {
  useListGuidelinesQuery,
  useUpdateGuidelineMutation,
} from "../api/annotations";
import { GuidelineSummaryCard } from "../components/Guidelines";
import type { GuidelineStatus } from "../types/reviewGuideline";

const STATUS_FILTERS: (GuidelineStatus | "all")[] = [
  "all",
  "active",
  "draft",
  "archived",
];

export function GuidelinesListPage() {
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState<GuidelineStatus | "all">("all");
  const [search, setSearch] = useState("");
  const { data: guidelines = [], isLoading } = useListGuidelinesQuery({
    status: statusFilter === "all" ? undefined : statusFilter,
    search: search || undefined,
  });
  const [updateGuideline] = useUpdateGuidelineMutation();

  const filtered = useMemo(() => {
    if (!search) return guidelines;
    const q = search.toLowerCase();
    return guidelines.filter(
      (g) =>
        g.title.toLowerCase().includes(q) ||
        g.slug.toLowerCase().includes(q) ||
        g.bodyMarkdown.toLowerCase().includes(q),
    );
  }, [guidelines, search]);

  const counts = useMemo(() => {
    const c = { all: 0, active: 0, draft: 0, archived: 0 };
    for (const g of guidelines) {
      c.all++;
      c[g.status]++;
    }
    return c;
  }, [guidelines]);

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading guidelines…
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="guidelines-list-page" sx={{ p: 3 }}>
      {/* Header */}
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
        <Typography variant="h2" sx={{ flex: 1 }}>
          Review Guidelines
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate("/annotations/guidelines/new")}
        >
          New Guideline
        </Button>
      </Stack>

      {/* Count summary */}
      <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
        {STATUS_FILTERS.map((s) => (
          <Chip
            key={s}
            label={`${s === "all" ? "All" : s.charAt(0).toUpperCase() + s.slice(1)} (${counts[s]})`}
            onClick={() => setStatusFilter(s)}
            color={statusFilter === s ? "primary" : "default"}
            variant={statusFilter === s ? "filled" : "outlined"}
          />
        ))}
      </Stack>

      {/* Search */}
      <TextField
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search guidelines…"
        size="small"
        fullWidth
        sx={{ mb: 2 }}
      />

      {/* List */}
      {filtered.length === 0 ? (
        <Box sx={{ py: 4, textAlign: "center" }}>
          <Typography variant="body2" color="text.secondary">
            No guidelines found.
          </Typography>
        </Box>
      ) : (
        filtered.map((g) => (
          <GuidelineSummaryCard
            key={g.id}
            guideline={g}
            linkedRunCount={0}
            onEdit={() =>
              navigate(`/annotations/guidelines/${g.id}`)
            }
            onArchive={
              g.status !== "archived"
                ? () =>
                    void updateGuideline({
                      id: g.id,
                      status: "archived",
                    })
                : undefined
            }
            onActivate={
              g.status === "archived"
                ? () =>
                    void updateGuideline({
                      id: g.id,
                      status: "active",
                    })
                : undefined
            }
          />
        ))
      )}
    </Box>
  );
}

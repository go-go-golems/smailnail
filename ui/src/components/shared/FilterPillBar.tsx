import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import { parts } from "./parts";

export interface FilterPill {
  key: string;
  label: string;
  count?: number;
}

export interface FilterPillBarProps {
  pills: FilterPill[];
  /** Currently active pill key, or null for "all" */
  activeKey: string | null;
  /** Called when a pill is clicked. Pass null for the "All" pill. */
  onSelect: (key: string | null) => void;
}

export function FilterPillBar({
  pills,
  activeKey,
  onSelect,
}: FilterPillBarProps) {
  return (
    <Stack
      data-part={parts.filterPills}
      direction="row"
      spacing={0.75}
      sx={{ flexWrap: "wrap", gap: 0.75 }}
    >
      <Chip
        label="All"
        size="small"
        variant={activeKey === null ? "filled" : "outlined"}
        color={activeKey === null ? "primary" : "default"}
        onClick={() => onSelect(null)}
        sx={{ fontWeight: 600 }}
      />
      {pills.map((pill) => (
        <Chip
          key={pill.key}
          label={pill.count != null ? `${pill.label} (${pill.count})` : pill.label}
          size="small"
          variant={activeKey === pill.key ? "filled" : "outlined"}
          color={activeKey === pill.key ? "primary" : "default"}
          onClick={() => onSelect(pill.key)}
          sx={{ fontWeight: activeKey === pill.key ? 600 : 400 }}
        />
      ))}
    </Stack>
  );
}

import Chip from "@mui/material/Chip";
import SmartToyIcon from "@mui/icons-material/SmartToy";
import PersonIcon from "@mui/icons-material/Person";
import AutoFixHighIcon from "@mui/icons-material/AutoFixHigh";
import FileDownloadIcon from "@mui/icons-material/FileDownload";
import type { SourceKind } from "../../types/annotations";
import { parts } from "./parts";

export interface SourceBadgeProps {
  sourceKind: SourceKind;
  sourceLabel: string;
  size?: "small" | "medium";
}

const iconMap: Record<SourceKind, React.ReactElement> = {
  agent: <SmartToyIcon sx={{ fontSize: 14 }} />,
  human: <PersonIcon sx={{ fontSize: 14 }} />,
  heuristic: <AutoFixHighIcon sx={{ fontSize: 14 }} />,
  import: <FileDownloadIcon sx={{ fontSize: 14 }} />,
};

export function SourceBadge({
  sourceKind,
  sourceLabel,
  size = "small",
}: SourceBadgeProps) {
  return (
    <Chip
      data-part={parts.sourceBadge}
      data-source-kind={sourceKind}
      icon={iconMap[sourceKind]}
      label={sourceLabel}
      size={size}
      variant="outlined"
      sx={{
        fontFamily: "monospace",
        fontSize: "0.6875rem",
        fontWeight: 500,
      }}
    />
  );
}

import Chip from "@mui/material/Chip";
import { getTagColor } from "../../theme/tagColors";
import { parts } from "./parts";

export interface TagChipProps {
  /** The tag name, e.g. "newsletter", "important" */
  tag: string;
  /** MUI Chip size */
  size?: "small" | "medium";
  /** Click handler — if provided, chip renders as clickable */
  onClick?: () => void;
}

export function TagChip({ tag, size = "small", onClick }: TagChipProps) {
  const colors = getTagColor(tag);

  return (
    <Chip
      data-part={parts.tagChip}
      data-tag={tag}
      label={tag}
      size={size}
      variant="outlined"
      onClick={onClick}
      sx={{
        bgcolor: colors.bg,
        color: colors.fg,
        borderColor: colors.border,
        fontWeight: 600,
        fontSize: "0.6875rem",
        cursor: onClick ? "pointer" : "default",
        "&:hover": onClick
          ? { bgcolor: colors.border }
          : {},
      }}
    />
  );
}

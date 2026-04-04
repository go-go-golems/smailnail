import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { parts } from "./parts";

export interface StatBoxProps {
  /** The large number or value */
  value: string | number;
  /** Label below the value */
  label: string;
  /** Color for the value text — use MUI palette paths like "primary.main" or hex */
  color?: string;
}

export function StatBox({
  value,
  label,
  color = "text.primary",
}: StatBoxProps) {
  return (
    <Box
      data-part={parts.statBox}
      sx={{
        textAlign: "center",
        px: 2,
        py: 1.5,
      }}
    >
      <Typography
        variant="h3"
        sx={{
          fontWeight: 700,
          fontFamily: "monospace",
          color,
          fontSize: "1.75rem",
          lineHeight: 1.2,
        }}
      >
        {value}
      </Typography>
      <Typography
        variant="caption"
        sx={{
          color: "text.secondary",
          display: "block",
          mt: 0.5,
        }}
      >
        {label}
      </Typography>
    </Box>
  );
}

import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import { parts } from "./parts";

export interface CountItem {
  label: string;
  value: number;
  color?: string;
}

export interface CountSummaryBarProps {
  items: CountItem[];
}

export function CountSummaryBar({ items }: CountSummaryBarProps) {
  return (
    <Stack
      data-part={parts.countSummary}
      direction="row"
      spacing={0.5}
      alignItems="center"
      sx={{ flexWrap: "wrap" }}
    >
      {items.map((item, i) => (
        <Typography
          key={item.label}
          variant="body2"
          component="span"
          sx={{ color: "text.secondary" }}
        >
          {i > 0 && (
            <Typography
              component="span"
              variant="body2"
              sx={{ color: "text.secondary", mx: 0.75 }}
            >
              ·
            </Typography>
          )}
          <Typography
            component="span"
            variant="body2"
            sx={{
              fontWeight: 700,
              fontFamily: "monospace",
              color: item.color ?? "text.primary",
              mr: 0.5,
            }}
          >
            {item.value}
          </Typography>
          {item.label}
        </Typography>
      ))}
    </Stack>
  );
}

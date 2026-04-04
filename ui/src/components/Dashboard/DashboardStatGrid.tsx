import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Divider from "@mui/material/Divider";
import { StatBox } from "../shared";

export interface DashboardStat {
  value: string | number;
  label: string;
  color?: string;
}

export interface DashboardStatGridProps {
  stats: DashboardStat[];
}

export function DashboardStatGrid({ stats }: DashboardStatGridProps) {
  return (
    <Paper data-widget="dashboard-stats">
      <Stack
        direction="row"
        spacing={0}
        divider={<Divider orientation="vertical" flexItem />}
        sx={{
          p: 0.5,
          flexWrap: "wrap",
          justifyContent: "center",
        }}
      >
        {stats.map((stat) => (
          <StatBox
            key={stat.label}
            value={stat.value}
            label={stat.label}
            color={stat.color}
          />
        ))}
      </Stack>
    </Paper>
  );
}

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { useNavigate } from "react-router-dom";
import { useListGroupsQuery } from "../api/annotations";
import { GroupCard } from "../components/GroupCard";

export function GroupsPage() {
  const navigate = useNavigate();
  const { data: groups = [], isLoading } = useListGroupsQuery({});

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading groups…
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="groups-page" sx={{ p: 3 }}>
      <Typography variant="h2" sx={{ mb: 2 }}>
        Target Groups
      </Typography>

      {groups.length === 0 ? (
        <Box sx={{ textAlign: "center", py: 6, color: "text.secondary" }}>
          <Typography variant="h4" sx={{ mb: 1 }}>
            No groups
          </Typography>
          <Typography variant="body2">
            Groups will appear here when agents group related targets.
          </Typography>
        </Box>
      ) : (
        groups.map((group) => (
          <GroupCard
            key={group.id}
            group={group}
            onNavigateTarget={(targetType, targetId) => {
              if (targetType === "sender") {
                navigate(
                  `/annotations/senders/${encodeURIComponent(targetId)}`,
                );
              }
            }}
          />
        ))
      )}
    </Box>
  );
}

import Box from "@mui/material/Box";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import DashboardIcon from "@mui/icons-material/Dashboard";
import RateReviewIcon from "@mui/icons-material/RateReview";
import SmartToyIcon from "@mui/icons-material/SmartToy";
import PeopleIcon from "@mui/icons-material/People";
import FolderSharedIcon from "@mui/icons-material/FolderShared";
import StorageIcon from "@mui/icons-material/Storage";
import { useLocation, useNavigate } from "react-router-dom";

interface NavItem {
  label: string;
  path: string;
  icon: React.ReactElement;
}

const overviewItems: NavItem[] = [
  {
    label: "Dashboard",
    path: "/annotations",
    icon: <DashboardIcon fontSize="small" />,
  },
];

const reviewItems: NavItem[] = [
  {
    label: "Review Queue",
    path: "/annotations/review",
    icon: <RateReviewIcon fontSize="small" />,
  },
  {
    label: "Agent Runs",
    path: "/annotations/runs",
    icon: <SmartToyIcon fontSize="small" />,
  },
];

const browseItems: NavItem[] = [
  {
    label: "Senders",
    path: "/annotations/senders",
    icon: <PeopleIcon fontSize="small" />,
  },
  {
    label: "Groups",
    path: "/annotations/groups",
    icon: <FolderSharedIcon fontSize="small" />,
  },
];

const toolItems: NavItem[] = [
  {
    label: "SQL Workbench",
    path: "/query",
    icon: <StorageIcon fontSize="small" />,
  },
];

function NavSection({
  title,
  items,
}: {
  title: string;
  items: NavItem[];
}) {
  const location = useLocation();
  const navigate = useNavigate();

  return (
    <>
      <Typography
        variant="overline"
        sx={{ px: 2, pt: 1.5, pb: 0.5, display: "block", color: "text.secondary" }}
      >
        {title}
      </Typography>
      <List dense disablePadding>
        {items.map((item) => {
          const isActive =
            item.path === "/annotations"
              ? location.pathname === "/annotations"
              : location.pathname.startsWith(item.path);

          return (
            <ListItemButton
              key={item.path}
              selected={isActive}
              onClick={() => navigate(item.path)}
              sx={{ px: 2, py: 0.75 }}
            >
              <ListItemIcon sx={{ minWidth: 32 }}>{item.icon}</ListItemIcon>
              <ListItemText
                primary={item.label}
                primaryTypographyProps={{
                  variant: "body2",
                  fontWeight: isActive ? 600 : 400,
                }}
              />
            </ListItemButton>
          );
        })}
      </List>
    </>
  );
}

export function AnnotationSidebar() {
  return (
    <Box
      data-part="sidebar"
      sx={{
        width: 220,
        flexShrink: 0,
        borderRight: 1,
        borderColor: "divider",
        bgcolor: "background.paper",
        height: "100%",
        overflow: "auto",
      }}
    >
      <NavSection title="Overview" items={overviewItems} />
      <Divider sx={{ my: 0.5 }} />
      <NavSection title="Review" items={reviewItems} />
      <Divider sx={{ my: 0.5 }} />
      <NavSection title="Browse" items={browseItems} />
      <Divider sx={{ my: 0.5 }} />
      <NavSection title="Tools" items={toolItems} />
    </Box>
  );
}

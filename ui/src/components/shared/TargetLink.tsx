import Link from "@mui/material/Link";
import Stack from "@mui/material/Stack";
import EmailIcon from "@mui/icons-material/Email";
import DomainIcon from "@mui/icons-material/Language";
import MessageIcon from "@mui/icons-material/Chat";
import GroupIcon from "@mui/icons-material/FolderShared";
import { parts } from "./parts";

export interface TargetLinkProps {
  targetType: string;
  targetId: string;
  onClick?: () => void;
}

const typeIcons: Record<string, React.ReactElement> = {
  sender: <EmailIcon sx={{ fontSize: 14, color: "text.secondary" }} />,
  domain: <DomainIcon sx={{ fontSize: 14, color: "text.secondary" }} />,
  message: <MessageIcon sx={{ fontSize: 14, color: "text.secondary" }} />,
  group: <GroupIcon sx={{ fontSize: 14, color: "text.secondary" }} />,
};

export function TargetLink({ targetType, targetId, onClick }: TargetLinkProps) {
  const icon = typeIcons[targetType] ?? null;

  return (
    <Stack
      data-part={parts.targetLink}
      data-target-type={targetType}
      direction="row"
      spacing={0.5}
      alignItems="center"
      sx={{ display: "inline-flex" }}
    >
      {icon}
      <Link
        component="button"
        variant="body2"
        onClick={onClick}
        underline={onClick ? "hover" : "none"}
        sx={{
          fontFamily: "monospace",
          fontSize: "0.8125rem",
          color: onClick ? "info.main" : "text.primary",
          cursor: onClick ? "pointer" : "default",
        }}
      >
        {targetId}
      </Link>
    </Stack>
  );
}

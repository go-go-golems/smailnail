import Chip from "@mui/material/Chip";
import PublicIcon from "@mui/icons-material/Public";
import MailOutlineIcon from "@mui/icons-material/MailOutline";
import PersonIcon from "@mui/icons-material/Person";
import DomainIcon from "@mui/icons-material/Domain";
import SettingsIcon from "@mui/icons-material/Settings";
import type { GuidelineScopeKind } from "../../types/reviewGuideline";
import { parts } from "./parts";

export interface GuidelineScopeBadgeProps {
  scopeKind: GuidelineScopeKind;
  size?: "small" | "medium";
}

const scopeConfig: Record<
  GuidelineScopeKind,
  { label: string; color: "default" | "primary" | "secondary" | "info" | "warning"; icon: React.ReactElement }
> = {
  global: { label: "Global", color: "default", icon: <PublicIcon fontSize="inherit" /> },
  mailbox: { label: "Mailbox", color: "primary", icon: <MailOutlineIcon fontSize="inherit" /> },
  sender: { label: "Sender", color: "secondary", icon: <PersonIcon fontSize="inherit" /> },
  domain: { label: "Domain", color: "info", icon: <DomainIcon fontSize="inherit" /> },
  workflow: { label: "Workflow", color: "warning", icon: <SettingsIcon fontSize="inherit" /> },
};

export function GuidelineScopeBadge({
  scopeKind,
  size = "small",
}: GuidelineScopeBadgeProps) {
  const config = scopeConfig[scopeKind];

  return (
    <Chip
      data-part={parts.guidelineScopeBadge}
      data-state={scopeKind}
      icon={config.icon}
      label={config.label}
      size={size}
      color={config.color}
      variant="outlined"
      sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
    />
  );
}

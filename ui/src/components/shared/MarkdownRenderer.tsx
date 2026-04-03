import Box from "@mui/material/Box";
import Markdown from "react-markdown";
import { parts } from "./parts";

export interface MarkdownRendererProps {
  /** Markdown content to render */
  content: string;
}

export function MarkdownRenderer({ content }: MarkdownRendererProps) {
  return (
    <Box
      data-part={parts.markdownBody}
      sx={{
        fontSize: "0.8125rem",
        lineHeight: 1.65,
        color: "text.primary",
        "& p": { mt: 0, mb: 1 },
        "& p:last-child": { mb: 0 },
        "& strong": { fontWeight: 700, color: "text.primary" },
        "& em": { fontStyle: "italic" },
        "& code": {
          fontFamily: "monospace",
          fontSize: "0.75rem",
          bgcolor: "action.hover",
          px: 0.5,
          py: 0.25,
          borderRadius: 0.5,
        },
        "& pre": {
          bgcolor: "background.default",
          p: 1.5,
          borderRadius: 1,
          overflow: "auto",
          "& code": { bgcolor: "transparent", p: 0 },
        },
        "& ul, & ol": { pl: 2.5, mt: 0, mb: 1 },
        "& li": { mb: 0.25 },
        "& a": { color: "info.main" },
      }}
    >
      <Markdown>{content}</Markdown>
    </Box>
  );
}

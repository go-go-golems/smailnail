import { useState } from "react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import MenuItem from "@mui/material/MenuItem";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import Tab from "@mui/material/Tab";
import Tabs from "@mui/material/Tabs";
import { MarkdownRenderer } from "../shared";
import { GuidelineScopeBadge } from "../shared";
import { parts } from "./parts";
import type {
  ReviewGuideline,
  GuidelineScopeKind,
  CreateGuidelineRequest,
  UpdateGuidelineRequest,
} from "../../types/reviewGuideline";

export interface GuidelineFormProps {
  guideline?: ReviewGuideline;
  mode: "view" | "edit" | "create";
  onSave: (payload: CreateGuidelineRequest | UpdateGuidelineRequest) => void;
  onCancel: () => void;
}

const SCOPE_OPTIONS: GuidelineScopeKind[] = [
  "global",
  "mailbox",
  "sender",
  "domain",
  "workflow",
];

export function GuidelineForm({
  guideline,
  mode,
  onSave,
  onCancel,
}: GuidelineFormProps) {
  const isReadOnly = mode === "view";
  const [tab, setTab] = useState<"edit" | "preview">("edit");

  const [title, setTitle] = useState(guideline?.title ?? "");
  const [slug, setSlug] = useState(guideline?.slug ?? "");
  const [scopeKind, setScopeKind] = useState<GuidelineScopeKind>(
    guideline?.scopeKind ?? "global",
  );
  const [bodyMarkdown, setBodyMarkdown] = useState(
    guideline?.bodyMarkdown ?? "",
  );

  const handleSubmit = () => {
    if (mode === "create") {
      onSave({
        slug,
        title,
        scopeKind,
        bodyMarkdown,
      } satisfies CreateGuidelineRequest);
    } else {
      onSave({
        title,
        scopeKind,
        bodyMarkdown,
      } satisfies UpdateGuidelineRequest);
    }
  };

  const canSave =
    title.trim().length > 0 &&
    bodyMarkdown.trim().length > 0 &&
    (mode === "edit" || slug.trim().length > 0);

  return (
    <Box data-part={parts.guidelineEditor}>
      {/* Header */}
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 2 }}>
        {guideline && (
          <>
            <GuidelineScopeBadge scopeKind={guideline.scopeKind} />
            <Typography variant="caption" color="text.secondary">
              {guideline.slug}
            </Typography>
          </>
        )}
        <Box sx={{ flex: 1 }} />
        {!isReadOnly && (
          <Stack direction="row" spacing={1}>
            <Button onClick={onCancel}>Cancel</Button>
            <Button
              variant="contained"
              onClick={handleSubmit}
              disabled={!canSave}
            >
              {mode === "create" ? "Create Guideline" : "Save Changes"}
            </Button>
          </Stack>
        )}
      </Stack>

      {/* Metadata fields */}
      <Stack spacing={2} sx={{ mb: 3 }}>
        <TextField
          label="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          fullWidth
          size="small"
          slotProps={{
            input: { readOnly: isReadOnly },
          }}
        />

        <Stack direction="row" spacing={2}>
          <TextField
            label="Slug"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            size="small"
            sx={{ flex: 1 }}
            slotProps={{
              input: { readOnly: mode !== "create" },
            }}
            helperText={
              mode !== "create"
                ? "Slug is immutable after creation"
                : "URL-safe identifier (e.g. my-guideline-name)"
            }
          />

          <TextField
            label="Scope"
            value={scopeKind}
            onChange={(e) => setScopeKind(e.target.value as GuidelineScopeKind)}
            size="small"
            sx={{ minWidth: 180 }}
            select={!isReadOnly}
            slotProps={{
              input: { readOnly: isReadOnly },
            }}
          >
            {SCOPE_OPTIONS.map((s) => (
              <MenuItem key={s} value={s}>
                {s}
              </MenuItem>
            ))}
          </TextField>
        </Stack>
      </Stack>

      <Divider sx={{ my: 2 }} />

      {/* Body editor / viewer */}
      {isReadOnly ? (
        <>
          <Typography variant="overline" sx={{ display: "block", mb: 1 }}>
            Body
          </Typography>
          <MarkdownRenderer content={guideline?.bodyMarkdown ?? ""} />
        </>
      ) : (
        <>
          <Tabs
            value={tab}
            onChange={(_, v) => setTab(v)}
            sx={{ mb: 1 }}
          >
            <Tab value="edit" label="Edit" />
            <Tab value="preview" label="Preview" />
          </Tabs>

          {tab === "edit" ? (
            <TextField
              value={bodyMarkdown}
              onChange={(e) => setBodyMarkdown(e.target.value)}
              multiline
              minRows={8}
              maxRows={20}
              fullWidth
              placeholder="Write guideline body in Markdown…"
              slotProps={{
                input: {
                  sx: { fontFamily: "monospace", fontSize: "0.8125rem" },
                },
              }}
            />
          ) : (
            <Box
              sx={{
                border: 1,
                borderColor: "divider",
                borderRadius: 1,
                p: 2,
                minHeight: 200,
              }}
            >
              {bodyMarkdown ? (
                <MarkdownRenderer content={bodyMarkdown} />
              ) : (
                <Typography variant="body2" color="text.secondary">
                  Nothing to preview — start writing in the Edit tab.
                </Typography>
              )}
            </Box>
          )}
        </>
      )}
    </Box>
  );
}

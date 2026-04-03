import { useState, useCallback } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import { useParams, useNavigate } from "react-router-dom";
import {
  useGetSenderQuery,
  useReviewAnnotationMutation,
} from "../api/annotations";
import {
  SenderProfileCard,
  AgentReasoningPanel,
  MessagePreviewTable,
} from "../components/SenderProfile";
import { AnnotationTable } from "../components/AnnotationTable";
import type { Annotation } from "../types/annotations";

export function SenderDetailPage() {
  const { email } = useParams<{ email: string }>();
  const navigate = useNavigate();
  const { data: sender, isLoading } = useGetSenderQuery(email ?? "");
  const [reviewAnnotation] = useReviewAnnotationMutation();

  const [selected, setSelected] = useState<string[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const annotations = sender?.annotations ?? [];
  const logs = sender?.logs ?? [];
  const messages = sender?.recentMessages ?? [];

  const getRelated = useCallback(
    (ann: Annotation) =>
      annotations.filter(
        (a) =>
          a.targetType === ann.targetType &&
          a.targetId === ann.targetId &&
          a.id !== ann.id,
      ),
    [annotations],
  );

  const handleToggleSelect = useCallback((id: string) => {
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }, []);

  const handleToggleAll = useCallback(() => {
    setSelected((prev) =>
      prev.length === annotations.length ? [] : annotations.map((a) => a.id),
    );
  }, [annotations]);

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading sender…
        </Typography>
      </Box>
    );
  }

  if (!sender) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" color="error.main">
          Sender not found: {email}
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="sender-detail-page" sx={{ p: 3 }}>
      {/* Header */}
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
        <Button
          size="small"
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate("/annotations/senders")}
        >
          All Senders
        </Button>
        <Box sx={{ flex: 1 }}>
          <Typography variant="h2">
            {sender.displayName || sender.email}
          </Typography>
          {sender.displayName && (
            <Typography
              variant="body2"
              sx={{ fontFamily: "monospace", color: "text.secondary" }}
            >
              {sender.email}
            </Typography>
          )}
        </Box>
      </Stack>

      {/* Profile card */}
      <SenderProfileCard sender={sender} />

      {/* Annotations */}
      {annotations.length > 0 && (
        <>
          <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
            Annotations ({annotations.length})
          </Typography>
          <AnnotationTable
            annotations={annotations}
            selected={selected}
            expandedId={expandedId}
            onToggleSelect={handleToggleSelect}
            onToggleAll={handleToggleAll}
            onToggleExpand={(id) =>
              setExpandedId((prev) => (prev === id ? null : id))
            }
            onApprove={(id) =>
              void reviewAnnotation({ id, reviewState: "reviewed" })
            }
            onDismiss={(id) =>
              void reviewAnnotation({ id, reviewState: "dismissed" })
            }
            getRelated={getRelated}
          />
          <Divider sx={{ my: 3 }} />
        </>
      )}

      {/* Agent reasoning */}
      {logs.length > 0 && (
        <>
          <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
            Agent Reasoning ({logs.length})
          </Typography>
          <AgentReasoningPanel logs={logs} />
          <Divider sx={{ my: 3 }} />
        </>
      )}

      {/* Recent messages */}
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Recent Messages ({messages.length})
      </Typography>
      <MessagePreviewTable messages={messages} />
    </Box>
  );
}

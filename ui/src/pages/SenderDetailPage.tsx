import { useState, useCallback, useMemo } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import { useParams, useNavigate } from "react-router-dom";
import {
  useGetSenderQuery,
  useGetSenderGuidelinesQuery,
  useListReviewFeedbackQuery,
  useReviewAnnotationMutation,
} from "../api/annotations";
import {
  SenderProfileCard,
  AgentReasoningPanel,
  MessagePreviewTable,
  SenderGuidelinePanel,
} from "../components/SenderProfile";
import { AnnotationTable } from "../components/AnnotationTable";
import { ReviewCommentDrawer } from "../components/ReviewFeedback";
import type { Annotation } from "../types/annotations";
import type { FeedbackKind } from "../types/reviewFeedback";

export function SenderDetailPage() {
  const { email } = useParams<{ email: string }>();
  const navigate = useNavigate();
  const [selected, setSelected] = useState<string[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [commentAnnotation, setCommentAnnotation] = useState<Annotation | null>(null);

  const { data: sender, isLoading } = useGetSenderQuery(email ?? "");
  const { data: senderGuidelineGroups = [] } = useGetSenderGuidelinesQuery(
    email ?? "",
    { skip: !email },
  );
  const { data: annotationFeedback = [] } = useListReviewFeedbackQuery(
    {
      scopeKind: "annotation",
      targetType: "annotation",
      targetId: expandedId ?? "",
    },
    { skip: !expandedId },
  );
  const [reviewAnnotation] = useReviewAnnotationMutation();

  const annotations = sender?.annotations ?? [];
  const logs = sender?.logs ?? [];
  const messages = sender?.recentMessages ?? [];

  const senderMailboxName = useMemo(() => {
    const mailboxNames = Array.from(
      new Set(
        messages
          .map((message) => message.mailboxName)
          .filter((mailboxName): mailboxName is string => mailboxName.length > 0),
      ),
    );
    return mailboxNames.length === 1 ? mailboxNames[0] : undefined;
  }, [messages]);

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

  const handleToggleExpand = useCallback((id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  }, []);

  const handleApprove = useCallback(
    (id: string) => {
      void reviewAnnotation({ id, reviewState: "reviewed" });
    },
    [reviewAnnotation],
  );

  const handleDismiss = useCallback(
    (id: string) => {
      void reviewAnnotation({ id, reviewState: "dismissed" });
    },
    [reviewAnnotation],
  );

  const handleDismissExplain = useCallback(
    (id: string) => {
      const annotation = annotations.find((item) => item.id === id) ?? null;
      setCommentAnnotation(annotation);
    },
    [annotations],
  );

  const handleCommentSubmit = useCallback(
    (payload: {
      feedbackKind: FeedbackKind;
      title: string;
      bodyMarkdown: string;
      guidelineIds: string[];
    }) => {
      if (!commentAnnotation) {
        return;
      }

      void reviewAnnotation({
        id: commentAnnotation.id,
        reviewState: "dismissed",
        comment: {
          feedbackKind: payload.feedbackKind,
          title: payload.title,
          bodyMarkdown: payload.bodyMarkdown,
        },
        guidelineIds:
          payload.guidelineIds.length > 0 ? payload.guidelineIds : undefined,
        mailboxName: senderMailboxName,
      });
      setCommentAnnotation(null);
    },
    [commentAnnotation, reviewAnnotation, senderMailboxName],
  );

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

      <SenderGuidelinePanel
        groups={senderGuidelineGroups}
        onNavigateRun={(runId) => navigate(`/annotations/runs/${encodeURIComponent(runId)}`)}
      />

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
            onToggleExpand={handleToggleExpand}
            onApprove={handleApprove}
            onDismiss={handleDismiss}
            onDismissExplain={handleDismissExplain}
            getRelated={getRelated}
            getFeedback={(annotation) =>
              expandedId === annotation.id ? annotationFeedback : []
            }
          />
          <Divider sx={{ my: 3 }} />
        </>
      )}

      <ReviewCommentDrawer
        open={commentAnnotation !== null}
        mode="single"
        targetCount={1}
        agentRunId={commentAnnotation?.agentRunId}
        mailboxName={senderMailboxName}
        onSubmit={handleCommentSubmit}
        onCancel={() => setCommentAnnotation(null)}
      />

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

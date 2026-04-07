import { useState, useCallback } from "react";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import Button from "@mui/material/Button";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import EditIcon from "@mui/icons-material/Edit";
import { useParams, useNavigate, useSearchParams, useLocation } from "react-router-dom";
import {
  useGetGuidelineQuery,
  useGetGuidelineRunsQuery,
  useCreateGuidelineMutation,
  useUpdateGuidelineMutation,
  useLinkGuidelineToRunMutation,
} from "../api/annotations";
import { GuidelineForm, GuidelineLinkedRuns } from "../components/Guidelines";
import type {
  CreateGuidelineRequest,
  UpdateGuidelineRequest,
} from "../types/reviewGuideline";

export function GuidelineDetailPage() {
  const { guidelineId } = useParams<{ guidelineId: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const runIdParam = searchParams.get("runId");

  const isNew = guidelineId === "new" || !guidelineId;
  const [mode, setMode] = useState<"view" | "edit" | "create">(
    isNew ? "create" : "view",
  );
  const [saveError, setSaveError] = useState<string | null>(
    (location.state as { flashError?: string } | null)?.flashError ?? null,
  );

  const { data: guideline, isLoading } = useGetGuidelineQuery(
    guidelineId ?? "",
    { skip: isNew },
  );
  const {
    data: linkedRuns = [],
    isFetching: linkedRunsLoading,
  } = useGetGuidelineRunsQuery(guidelineId ?? "", {
    skip: isNew || mode !== "view",
  });
  const [createGuideline] = useCreateGuidelineMutation();
  const [updateGuideline] = useUpdateGuidelineMutation();
  const [linkGuidelineToRun] = useLinkGuidelineToRunMutation();

  const handleSave = useCallback(
    async (payload: CreateGuidelineRequest | UpdateGuidelineRequest) => {
      setSaveError(null);
      try {
        if (mode === "create") {
          const created = await createGuideline(
            payload as CreateGuidelineRequest,
          ).unwrap();

          if (runIdParam) {
            try {
              await linkGuidelineToRun({
                runId: runIdParam,
                guidelineId: created.id,
              }).unwrap();
              navigate(`/annotations/runs/${runIdParam}`);
              return;
            } catch {
              navigate(`/annotations/guidelines/${created.id}`, {
                state: {
                  flashError:
                    "Guideline created, but linking it to the run failed. Please link it manually.",
                },
              });
              return;
            }
          }

          navigate(`/annotations/guidelines/${created.id}`);
          return;
        }

        await updateGuideline({
          id: guidelineId ?? "",
          ...(payload as UpdateGuidelineRequest),
        }).unwrap();
        setMode("view");
      } catch {
        setSaveError("Failed to save the guideline. Please try again.");
      }
    },
    [mode, createGuideline, updateGuideline, linkGuidelineToRun, guidelineId, runIdParam, navigate],
  );

  const handleCancel = useCallback(() => {
    if (mode === "create") {
      navigate(-1);
    } else {
      setMode("view");
    }
  }, [mode, navigate]);

  if (!isNew && isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading guideline…
        </Typography>
      </Box>
    );
  }

  if (!isNew && !guideline) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" color="error.main">
          Guideline not found: {guidelineId}
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="guideline-detail-page" sx={{ p: 3 }}>
      {/* Header */}
      <Box sx={{ display: "flex", alignItems: "center", gap: 2, mb: 2 }}>
        <Button
          size="small"
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate("/annotations/guidelines")}
        >
          All Guidelines
        </Button>
        {mode === "view" && guideline && (
          <Button
            size="small"
            startIcon={<EditIcon />}
            onClick={() => setMode("edit")}
          >
            Edit
          </Button>
        )}
      </Box>

      {saveError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {saveError}
        </Alert>
      )}

      {/* Form */}
      <GuidelineForm
        guideline={guideline}
        mode={mode}
        onSave={(payload) => {
          void handleSave(payload);
        }}
        onCancel={handleCancel}
      />

      {/* Linked runs (view mode only) */}
      {mode === "view" && guideline && (
        <>
          <Divider sx={{ my: 3 }} />
          <GuidelineLinkedRuns
            runs={linkedRuns}
            loading={linkedRunsLoading}
            onNavigateRun={(id) =>
              navigate(`/annotations/runs/${encodeURIComponent(id)}`)
            }
          />
        </>
      )}
    </Box>
  );
}

import { useState, useCallback } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import Button from "@mui/material/Button";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import EditIcon from "@mui/icons-material/Edit";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import {
  useGetGuidelineQuery,
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
  const [searchParams] = useSearchParams();
  const runIdParam = searchParams.get("runId");

  const isNew = guidelineId === "new" || !guidelineId;
  const { data: guideline, isLoading } = useGetGuidelineQuery(
    guidelineId ?? "",
    { skip: isNew },
  );
  const [createGuideline] = useCreateGuidelineMutation();
  const [updateGuideline] = useUpdateGuidelineMutation();
  const [linkGuidelineToRun] = useLinkGuidelineToRunMutation();

  const [mode, setMode] = useState<"view" | "edit" | "create">(
    isNew ? "create" : "view",
  );

  const handleSave = useCallback(
    (payload: CreateGuidelineRequest | UpdateGuidelineRequest) => {
      if (mode === "create") {
        void createGuideline(payload as CreateGuidelineRequest).then(
          (result) => {
            if ("data" in result && result.data) {
              // If we came from a run context, link the new guideline
              if (runIdParam) {
                void linkGuidelineToRun({
                  runId: runIdParam,
                  guidelineId: result.data.id,
                });
                navigate(`/annotations/runs/${runIdParam}`);
              } else {
                navigate(`/annotations/guidelines/${result.data.id}`);
              }
            }
          },
        );
      } else {
        void updateGuideline({
          id: guidelineId ?? "",
          ...(payload as UpdateGuidelineRequest),
        }).then(() => {
          setMode("view");
        });
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

      {/* Form */}
      <GuidelineForm
        guideline={guideline}
        mode={mode}
        onSave={handleSave}
        onCancel={handleCancel}
      />

      {/* Linked runs (view mode only) */}
      {mode === "view" && guideline && (
        <>
          <Divider sx={{ my: 3 }} />
          <GuidelineLinkedRuns
            runs={[]}
            onNavigateRun={(id) =>
              navigate(`/annotations/runs/${encodeURIComponent(id)}`)
            }
          />
        </>
      )}
    </Box>
  );
}

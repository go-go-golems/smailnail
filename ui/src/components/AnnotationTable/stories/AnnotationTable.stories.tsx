import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import { AnnotationTable } from "../AnnotationTable";
import { mockAnnotations } from "../../../mocks/annotations";

const meta = {
  title: "Annotations/AnnotationTable",
  component: AnnotationTable,
  args: {
    annotations: mockAnnotations,
    selected: [],
    expandedId: null,
    onToggleSelect: fn(),
    onToggleAll: fn(),
    onToggleExpand: fn(),
    onApprove: fn(),
    onDismiss: fn(),
    onNavigateTarget: fn(),
  },
  decorators: [
    (Story) => (
      <Box sx={{ width: "100%", maxWidth: 1200 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof AnnotationTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  args: { annotations: [] },
};

export const WithSelection: Story = {
  args: { selected: ["ann-001", "ann-002", "ann-006"] },
};

export const WithExpanded: Story = {
  args: { expandedId: "ann-002" },
};

export const WithRelated: Story = {
  args: {
    expandedId: "ann-001",
    getRelated: (ann) =>
      mockAnnotations.filter(
        (a) =>
          a.targetId === ann.targetId &&
          a.targetType === ann.targetType &&
          a.id !== ann.id,
      ),
  },
};

/** Interactive demo with local state */
export const Interactive: Story = {
  render: () => {
    const [selected, setSelected] = useState<string[]>([]);
    const [expandedId, setExpandedId] = useState<string | null>(null);

    return (
      <AnnotationTable
        annotations={mockAnnotations}
        selected={selected}
        expandedId={expandedId}
        onToggleSelect={(id) =>
          setSelected((prev) =>
            prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
          )
        }
        onToggleAll={() =>
          setSelected((prev) =>
            prev.length === mockAnnotations.length
              ? []
              : mockAnnotations.map((a) => a.id),
          )
        }
        onToggleExpand={(id) =>
          setExpandedId((prev) => (prev === id ? null : id))
        }
        onApprove={fn()}
        onDismiss={fn()}
        onNavigateTarget={fn()}
      />
    );
  },
};

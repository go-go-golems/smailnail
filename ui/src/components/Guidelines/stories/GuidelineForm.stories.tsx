import type { Meta, StoryObj } from "@storybook/react";
import { GuidelineForm } from "../GuidelineForm";
import type { GuidelineFormProps } from "../GuidelineForm";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "Guidelines/GuidelineForm",
  component: GuidelineForm,
  tags: ["autodocs"],
} satisfies Meta<typeof GuidelineForm>;

export default meta;
type Story = StoryObj<typeof GuidelineForm>;

export const ViewMode: Story = {
  args: {
    guideline: mockGuidelines[0],
    mode: "view",
    onSave: () => {},
    onCancel: () => console.log("cancel"),
  } satisfies GuidelineFormProps,
};

export const EditMode: Story = {
  args: {
    guideline: mockGuidelines[0],
    mode: "edit",
    onSave: (payload) => console.log("save", payload),
    onCancel: () => console.log("cancel"),
  } satisfies GuidelineFormProps,
};

export const CreateMode: Story = {
  args: {
    mode: "create",
    onSave: (payload) => console.log("create", payload),
    onCancel: () => console.log("cancel"),
  } satisfies GuidelineFormProps,
};

export const CreateWithPreview: Story = {
  render: () => (
    <GuidelineForm
      guideline={{
        id: "preview-id",
        slug: "preview-guideline",
        title: "Preview Guideline",
        scopeKind: "global",
        status: "draft",
        priority: 50,
        bodyMarkdown:
          "This is a **preview** of what the guideline will look like.\n\n- Item one\n- Item two\n\n> A blockquote for emphasis.",
        createdBy: "manuel",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }}
      mode="create"
      onSave={(payload) => console.log("create", payload)}
      onCancel={() => console.log("cancel")}
    />
  ),
};

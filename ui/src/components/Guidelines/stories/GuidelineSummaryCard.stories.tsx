import type { Meta, StoryObj } from "@storybook/react";
import { GuidelineSummaryCard } from "../GuidelineSummaryCard";
import type { GuidelineSummaryCardProps } from "../GuidelineSummaryCard";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "Guidelines/GuidelineSummaryCard",
  component: GuidelineSummaryCard,
  tags: ["autodocs"],
} satisfies Meta<typeof GuidelineSummaryCard>;

export default meta;
type Story = StoryObj<typeof GuidelineSummaryCard>;

const activeGuideline = mockGuidelines[0]!;
const draftGuideline = mockGuidelines[2]!;
const archivedGuideline = mockGuidelines[3]!;

export const Active: Story = {
  args: {
    guideline: activeGuideline,
    linkedRunCount: 3,
    onEdit: () => console.log("edit"),
    onArchive: () => console.log("archive"),
  } satisfies GuidelineSummaryCardProps,
};

export const Draft: Story = {
  args: {
    guideline: draftGuideline,
    linkedRunCount: 0,
    onEdit: () => console.log("edit"),
    onArchive: () => console.log("archive"),
  } satisfies GuidelineSummaryCardProps,
};

export const Archived: Story = {
  args: {
    guideline: archivedGuideline,
    linkedRunCount: 1,
    onEdit: () => console.log("edit"),
    onActivate: () => console.log("activate"),
  } satisfies GuidelineSummaryCardProps,
};

export const AllStates: Story = {
  render: () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
      <GuidelineSummaryCard
        guideline={activeGuideline}
        linkedRunCount={3}
        onEdit={() => {}}
        onArchive={() => {}}
      />
      <GuidelineSummaryCard
        guideline={draftGuideline}
        linkedRunCount={0}
        onEdit={() => {}}
        onArchive={() => {}}
      />
      <GuidelineSummaryCard
        guideline={archivedGuideline}
        linkedRunCount={1}
        onEdit={() => {}}
        onActivate={() => {}}
      />
    </div>
  ),
};

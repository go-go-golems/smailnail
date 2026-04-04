import type { Meta, StoryObj } from "@storybook/react";
import { GuidelineLinkedRuns } from "../GuidelineLinkedRuns";
import type { GuidelineLinkedRunsProps } from "../GuidelineLinkedRuns";
import { mockRuns } from "../../../mocks/annotations";

const meta = {
  title: "Guidelines/GuidelineLinkedRuns",
  component: GuidelineLinkedRuns,
  tags: ["autodocs"],
} satisfies Meta<typeof GuidelineLinkedRuns>;

export default meta;
type Story = StoryObj<typeof GuidelineLinkedRuns>;

const sampleRuns = [mockRuns[0]!, mockRuns[1]!, mockRuns[2]!];

export const MultipleRuns: Story = {
  args: {
    runs: sampleRuns,
    onNavigateRun: (runId: string) => console.log("navigate", runId),
  } satisfies GuidelineLinkedRunsProps,
};

export const SingleRun: Story = {
  args: {
    runs: [sampleRuns[0]!],
    onNavigateRun: (runId: string) => console.log("navigate", runId),
  } satisfies GuidelineLinkedRunsProps,
};

export const Empty: Story = {
  args: {
    runs: [],
  } satisfies GuidelineLinkedRunsProps,
};

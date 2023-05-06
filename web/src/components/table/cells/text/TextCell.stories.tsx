import { Meta, Story } from "@storybook/react";
import TextCellMemo, { TextCellProps } from "./TextCell";

export default {
  title: "components/table/cells/TextCell",
  component: TextCellMemo,
} as Meta;

const Template: Story<TextCellProps> = (args) => <TextCellMemo {...args} />;

export const Primary = Template.bind({});
Primary.args = {
  text: "Some value text value that is longer than the other values",
  link: "https://www.google.com",
  copyText: "Some value text value that is longer than the other values",
  totalCell: false,
};

export const NoData = Template.bind({});
NoData.args = { text: undefined };

export const Total = Template.bind({});
Total.args = {
  text: "Some value text value that is longer than the other values",
  link: "https://www.google.com",
  copyText: "Some value text value that is longer than the other values",
  totalCell: true,
};

import { useState } from "react";
import { ArrowRotateClockwise20Regular as ReBalanceIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "../nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { NumberFormatValues } from "react-number-format";
import { useSelector } from "react-redux";
import { Input, InputSizeVariant, Socket, Form } from "components/forms/forms";

type ReBalanceChannelNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type RebalanceParameters = {
  // There will only be one of either outgoing or incoming channel ID
  outgoingChannelId?: number;
  incomingChannelId?: number;
  // The channels that are at the other side of the re-balance request
  channelIds: Array<number>;

  amountMsat: number;
  maximumCostMsat: number;
  maximumConcurrency?: number;
};

export function ReBalanceChannelNode({ ...wrapperProps }: ReBalanceChannelNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const [parameters, setParameters] = useState<RebalanceParameters>({
    outgoingChannelId: undefined,
    incomingChannelId: undefined,
    channelIds: [],
    amountMsat: 0,
    maximumCostMsat: 0,
    maximumConcurrency: undefined,
    ...wrapperProps.parameters,
  });

  const [amountSat, setAmountSat] = useState<number | undefined>(
    ((wrapperProps.parameters as RebalanceParameters).amountMsat || 0) / 1000
  );
  const [maximumCostSat, setMaximumCostSat] = useState<number | undefined>(
    ((wrapperProps.parameters as RebalanceParameters).maximumCostMsat || 0) / 1000
  );

  function handleAmountSatChange(e: NumberFormatValues) {
    setAmountSat(e.floatValue);
    setParameters((prev) => ({
      ...prev,
      amountMsat: (e.floatValue || 0) * 1000,
    }));
  }

  function handleMaximumCostSatChange(e: NumberFormatValues) {
    setMaximumCostSat(e.floatValue);
    setParameters((prev) => ({
      ...prev,
      maximumCostMsat: (e.floatValue || 0) * 1000,
    }));
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: parameters,
    });
  }

  const { childLinks } = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      stage: wrapperProps.stage,
    })
  );

  const destinationChannelsIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "destinationChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const destinationChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: destinationChannelsIds,
    })
  );

  const sourceChannelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "sourceChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const sourceChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: sourceChannelIds,
    })
  );

  const avoidChannelsIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "avoidChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const avoidChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: avoidChannelsIds,
    })
  );

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      heading={t.channelPolicyConfiguration}
      headerIcon={<ReBalanceIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Destinations}
          selectedNodes={destinationChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"destinationChannels"}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Sources}
          selectedNodes={sourceChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"sourceChannels"}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Avoid}
          selectedNodes={avoidChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"avoidChannels"}
        />
        <Input
          formatted={true}
          value={amountSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={handleAmountSatChange}
          label={t.amountSat}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={maximumCostSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={handleMaximumCostSatChange}
          label={t.maximumCostSat}
          sizeVariant={InputSizeVariant.small}
        />
        {/*<Input*/}
        {/*  formatted={true}*/}
        {/*  value={parameters.maximumConcurrency}*/}
        {/*  thousandSeparator={","}*/}
        {/*  suffix={" sat"}*/}
        {/*  onValueChange={createChangeHandler("maximumConcurrency")}*/}
        {/*  label={t.maximumConcurrency}*/}
        {/*  sizeVariant={InputSizeVariant.small}*/}
        {/*/>*/}
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}

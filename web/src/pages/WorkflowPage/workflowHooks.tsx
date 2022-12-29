import {
  PuzzlePiece20Regular as NodesIcon,
  Play20Regular as DeployIcon,
  Add16Regular as NewWorkflowIcon,
} from "@fluentui/react-icons";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router";
import { useGetWorkflowQuery, useNewWorkflowMutation } from "pages/WorkflowPage/workflowApi";
import { ReactNode } from "react";
import { Workflow, WorkflowStages, WorkflowVersion } from "./workflowTypes";
import ChannelPolicyNode from "components/workflow/nodes/channelPolicy/ChannelPolicy";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import { TriggerNodeTypes } from "./constants";
import nodeStyles from "components/workflow/nodeWrapper/workflow_nodes.module.scss";

export function useNewWorkflowButton(): ReactNode {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const [newWorkflow] = useNewWorkflowMutation();

  function newWorkflowHandler() {
    const response = newWorkflow();
    response
      .then((res) => {
        console.log(res);
        const data = (res as { data: { workflowId: number; version: number } }).data;
        navigate(`/manage/workflows/${data.workflowId}/versions/${data.version}`);
      })
      .catch((err) => {
        // TODO: Handle error and show a toast
        console.log(err);
      });
  }

  return (
    <Button
      buttonColor={ColorVariant.success}
      className={"collapse-tablet"}
      icon={<NewWorkflowIcon />}
      onClick={newWorkflowHandler}
    >
      {t.newWorkflow}
    </Button>
  );
}

export function useWorkflowData(workflowId?: string, version?: string) {
  const { data } = useGetWorkflowQuery(
    {
      workflowId: parseInt(workflowId || ""),
      version: parseInt(version || ""),
    },
    { skip: !workflowId || !version }
  );

  const workflow: Workflow | undefined = data?.workflow;
  const workflowVersion: WorkflowVersion | undefined = data?.version;

  const stages: WorkflowStages = data?.workflowForest?.sortedStageTrees || {}; //.map((s) => parseInt(s));

  return { workflow, workflowVersion, stages };
}

export function useNodes(stages: WorkflowStages, stageNumber: number) {
  const triggerNodes = (stages[stageNumber] || [])
    .filter((node) => {
      // Filter out the trigger nodes
      return TriggerNodeTypes.includes(node.type);
    })
    .map((node) => {
      const nodeId = node.workflowVersionNodeId;
      return <ChannelPolicyNode {...node} key={`node-${nodeId}`} id={`node-${nodeId}`} name={node.name} />;
    })
    .sort((a, b) => a.props.id.localeCompare(b.props.id));

  const actionNodes = (stages[stageNumber] || [])
    .filter((node) => {
      // Filter out the trigger nodes
      return !TriggerNodeTypes.includes(node.type);
    })
    .map((node) => {
      const nodeId = node.workflowVersionNodeId;
      return <ChannelPolicyNode {...node} key={`node-${nodeId}`} id={`node-${nodeId}`} name={node.name} />;
    })
    .sort((a, b) => a.props.id.localeCompare(b.props.id));

  return { triggerNodes, actionNodes };
}

export function useStages(workflowVersionId: number, stages: WorkflowStages, selectedStage: number) {
  return Object.entries(stages).map((stage) => {
    const stageNumber = parseInt(stage[0]);
    const { triggerNodes, actionNodes } = useNodes(stages, stageNumber);
    return (
      <WorkflowCanvas
        active={selectedStage === stageNumber}
        key={`stage-${stageNumber}`}
        workflowVersionId={workflowVersionId}
        stageNumber={stageNumber}
      >
        <div className={nodeStyles.triggerNodeWrapper}>
          <div className={nodeStyles.triggerNodeContainer}>{triggerNodes}</div>
        </div>
        {actionNodes}
      </WorkflowCanvas>
    );
  });
}

export function useWorkflowControls(sidebarExpanded: boolean, setSidebarExpanded: (expanded: boolean) => void) {
  const { t } = useTranslations();
  return (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.success}
            className={"collapse-tablet"}
            icon={<DeployIcon />}
            onClick={() => {
              console.log("Not implemented yet");
            }}
          >
            {t.deploy}
          </Button>
        </TableControlsTabsGroup>
        <Button
          buttonColor={ColorVariant.primary}
          className={"collapse-tablet"}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            setSidebarExpanded(!sidebarExpanded);
          }}
        >
          {t.nodes}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );
}

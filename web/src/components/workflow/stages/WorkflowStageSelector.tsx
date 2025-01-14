import classNames from "classnames";
import { Add16Regular as NewStageIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useDeleteStageMutation } from "pages/WorkflowPage/workflowApi";
import styles from "./workflow_stages.module.scss";
import { ReactComponent as StageArrowBack } from "pages/WorkflowPage/stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "pages/WorkflowPage/stageArrowFront.svg";
import { useAddNodeMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import useTranslations from "services/i18n/useTranslations";
import React from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import { userEvents } from "utils/userEvents";

type StageSelectorProps = {
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  workflowId: number;
  version: number;
  editingDisabled: boolean;
};

export function StageSelector({
  stageNumbers,
  selectedStage,
  setSelectedStage,
  workflowVersionId,
  workflowId,
  version,
  editingDisabled,
}: StageSelectorProps) {
  return (
    <div className={styles.stagesWrapper} data-intercom-target={"workflow-stages-selector"}>
      {stageNumbers.map((stage, index) => {
        return (
          <SelectStageButton
            key={`stage-${stage}`}
            selectedStage={selectedStage}
            setSelectedStage={setSelectedStage}
            stage={stage}
            stageNumbers={stageNumbers}
            buttonIndex={index}
            workflowId={workflowId}
            version={version}
            editingDisabled={editingDisabled}
          />
        );
      })}
      <AddStageButton
        editingDisabled={editingDisabled}
        setSelectedStage={setSelectedStage}
        workflowVersionId={workflowVersionId}
        selectedStage={selectedStage}
        stageNumbers={stageNumbers}
      />
    </div>
  );
}

type SelectStageButtonProps = {
  stage: number;
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  buttonIndex: number;
  workflowId: number;
  version: number;
  editingDisabled: boolean;
};

function SelectStageButton(props: SelectStageButtonProps) {
  const { stage, stageNumbers, selectedStage, setSelectedStage, buttonIndex, workflowId, version } = props;
  const { t } = useTranslations();
  const { track } = userEvents();
  const toastRef = React.useContext(ToastContext);
  const [deleteStage] = useDeleteStageMutation();

  function handleDeleteStage(stage: number) {
    if (props.editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }
    // Ask the user to confirm deletion of the stage
    if (!confirm(t.deleteStageConfirm)) {
      return;
    }
    track("Workflow Delete Stage", {
      workflowId: workflowId,
      workflowVersion: version,
      workflowStage: stage,
    });
    deleteStage({ workflowId, version, stage }).then(() => {
      // On success, select the preceding stage
      const precedingStage = stageNumbers.slice(0, stageNumbers.indexOf(stage)).pop();
      setSelectedStage(precedingStage || 1);
    });
  }

  // NB: The stage is the stage ID used on the nodes. The buttonIndex is used to display the stage number.
  //   This is because the user can delete a stage in between two stages, and then the stage numbers will not be consecutive.

  return (
    <button
      data-intercom-target={`workflow-stage-select-button-${buttonIndex}`}
      className={classNames(styles.stageContainer, { [styles.selected]: stage === selectedStage })}
      onClick={() => setSelectedStage(stage)}
    >
      {buttonIndex !== 0 && <StageArrowBack />}
      <div className={styles.stage}>
        {`${t.stage} ${buttonIndex + 1}`}
        {buttonIndex !== 0 && (
          <div className={classNames(styles.deleteStage)} onClick={() => handleDeleteStage(stage)}>
            <DeleteIcon />
          </div>
        )}
      </div>
      <StageArrowFront />
    </button>
  );
}

type AddStageButtonProps = {
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  editingDisabled: boolean;
};

function AddStageButton(props: AddStageButtonProps) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const toastRef = React.useContext(ToastContext);
  const [addNode] = useAddNodeMutation();
  const nextStage = Math.max(...props.stageNumbers) + 1;

  function handleAddStage() {
    if (props.editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    track("Workflow Add Stage", {
      workflowVersionId: props.workflowVersionId,
      workflowCurrentStage: props.selectedStage,
      workflowNextStage: nextStage,
    });
    addNode({
      type: WorkflowNodeType.StageTrigger,
      name: `${t.stage} ${nextStage}`,
      visibilitySettings: {
        xPosition: 0,
        yPosition: 0,
        collapsed: false,
      },
      workflowVersionId: props.workflowVersionId,
      stage: nextStage,
    }).then(() => {
      // On success, select the new stage
      props.setSelectedStage(nextStage);
    });
  }

  return (
    <button
      data-intercom-target={"workflow-add-stage-button"}
      className={classNames(
        styles.stageContainer,
        props.editingDisabled ? styles.disabledStage : styles.addStageButton
      )}
      onClick={handleAddStage}
    >
      <StageArrowBack />
      <div className={classNames(styles.stage, props.editingDisabled ? styles.disabled : "")}>
        <NewStageIcon />
      </div>
      <StageArrowFront />
    </button>
  );
}

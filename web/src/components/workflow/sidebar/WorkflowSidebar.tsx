import useTranslations from "services/i18n/useTranslations";
import classNames from "classnames";
import styles from "./workflow_sidebar.module.scss";
import Sidebar from "features/sidebar/Sidebar";
import { SectionContainer } from "features/section/SectionContainer";
import {
  Timer20Regular as TriggersIcon,
  Play20Regular as DataSourcesIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import { useState } from "react";
import {
  ChannelPolicyConfiguratorNodeButton,
  ChannelPolicyAutoRunNodeButton,
  ChannelPolicyRunNodeButton,
  IntervalTriggerNodeButton,
  CronTriggerNodeButton,
  ChannelFilterNodeButton,
  RebalanceConfiguratorNodeButton,
  RebalanceAutoRunNodeButton,
  RebalanceRunNodeButton,
  RemoveTagNodeButton,
  BalanceTriggerNodeButton,
  AddTagNodeButton,
  ChannelCloseTriggerNodeButton,
  ChannelOpenTriggerNodeButton,
  DataSourceTorqChannelsNodeButton,
  ChannelBalanceEventFilterNodeButton,
} from "components/workflow/nodes/nodes";
import { userEvents } from "utils/userEvents";

export type WorkflowSidebarProps = {
  expanded: boolean;
  setExpanded: (expanded: boolean) => void;
};

export default function WorkflowSidebar(props: WorkflowSidebarProps) {
  const { track } = userEvents();
  const { expanded, setExpanded } = props;
  const { t } = useTranslations();
  const closeSidebarHandler = () => {
    track("Workflow Toggle Sidebar");
    setExpanded(false);
  };

  const [sectionState, setSectionState] = useState({
    triggers: true,
    dataSources: true,
    actions: true,
    advanced: false,
  });

  const toggleSection = (section: keyof typeof sectionState) => {
    setSectionState({
      ...sectionState,
      [section]: !sectionState[section],
    });
  };

  return (
    <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: expanded })}>
      <Sidebar title={t.actions} closeSidebarHandler={closeSidebarHandler}>
        {" "}
        <SectionContainer
          intercomTarget={"workflow-triggers-section"}
          title={t.triggers}
          icon={TriggersIcon}
          expanded={sectionState.triggers}
          handleToggle={() => toggleSection("triggers")}
        >
          <IntervalTriggerNodeButton />
          <CronTriggerNodeButton />
          <BalanceTriggerNodeButton />
          <ChannelOpenTriggerNodeButton />
          <ChannelCloseTriggerNodeButton />
        </SectionContainer>
        <SectionContainer
          intercomTarget={"workflow-data-sources-section"}
          title={t.dataSources}
          icon={DataSourcesIcon}
          expanded={sectionState.dataSources}
          handleToggle={() => toggleSection("dataSources")}
        >
          <DataSourceTorqChannelsNodeButton />
        </SectionContainer>
        <SectionContainer
          intercomTarget={"workflow-actions-section"}
          title={t.actions}
          icon={ChannelsIcon}
          expanded={sectionState.actions}
          handleToggle={() => toggleSection("actions")}
        >
          <ChannelFilterNodeButton />
          <ChannelPolicyAutoRunNodeButton />
          <RebalanceAutoRunNodeButton />
          <AddTagNodeButton />
          <RemoveTagNodeButton />
        </SectionContainer>
        <SectionContainer
          intercomTarget={"workflow-advanced-actions-section"}
          title={t.AdvancedActions}
          icon={TagsIcon}
          expanded={sectionState.advanced}
          handleToggle={() => toggleSection("advanced")}
        >
          <ChannelBalanceEventFilterNodeButton />
          <ChannelPolicyConfiguratorNodeButton />
          <ChannelPolicyRunNodeButton />
          <RebalanceConfiguratorNodeButton />
          <RebalanceRunNodeButton />
        </SectionContainer>
      </Sidebar>
    </div>
  );
}

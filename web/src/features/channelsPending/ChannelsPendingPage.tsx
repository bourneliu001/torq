import { Link } from "react-router-dom";
import {
  Options20Regular as OptionsIcon,
  ArrowDownload20Regular as DownloadCsvIcon,
  ArrowSync20Regular as RefreshIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { ChannelPending } from "features/channelsPending/channelsPendingTypes";
import * as Routes from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import Table from "features/table/Table";
import {
  ChannelsPendingFilterTemplate,
  ChannelsPendingSortTemplate,
  FilterableChannelsPendingColumns,
  SortableChannelsPendingColumns,
  DefaultPendingChannelsView,
} from "features/channelsPending/channelsPendingDefaults";
import { useGetChannelsPendingQuery } from "apiSlice";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { selectPendingChannelView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { useState } from "react";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import { TableResponses, ViewResponse } from "features/viewManagement/types";
import channelsPendingCellRenderer from "features/channelsPending/channelsPendingCellRenderer";
import { AllChannelPendingColumns } from "features/channelsPending/channelsPendingColumns.generated";
import { createCsvFile } from "utils/JsonTableToCsv";
import Button, { ColorVariant } from "components/buttons/Button";
import { userEvents } from "utils/userEvents";

function useMaximums(data: Array<ChannelPending>): ChannelPending | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: ChannelPending, current: ChannelPending) => {
    return {
      ...prev,
      alias: "Max",
      capacity: Math.max(prev.capacity, current.capacity),
    };
  });
}

function ChannelsPendingPage() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPendingChannelView);
  const channelViews = useAppSelector(selectViews)("channelsPending");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [updateTableView] = useUpdateTableViewMutation();

  const channelsResponse = useGetChannelsPendingQuery<{
    data: Array<ChannelPending>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !isSuccess, pollingInterval: 10000 });

  const filteredData = useFilterData(channelsResponse.data, viewResponse.view.filters);
  const data = useSortData(filteredData, viewResponse.view.sortBy);
  const maxRow = useMaximums(data);

  // Logic for toggling the sidebar
  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    track("Toggle Table Sidebar", { page: "ChannelsPending" });
  };

  function handleNameChange(name: string) {
    const view = channelViews.views[selectedViewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: name },
      });
    }
  }

  const tableControls = (
    <TableControlSection intercomTarget={"table-page-controls"}>
      <div />
      <TableControlsButtonGroup intercomTarget="table-page-controls-right">
        <Button
          intercomTarget="download-table"
          buttonColor={ColorVariant.primary}
          title={t.download}
          hideMobileText={true}
          icon={<DownloadCsvIcon />}
          onClick={() => {
            track("Downloads Table as CSV", {
              downloadTablePage: "Channels Pending",
              downloadTableViewTitle: viewResponse.view.title,
              downloadTableColumns: viewResponse.view.columns,
              downloadTableFilters: viewResponse.view.filters,
              downloadTableSortBy: viewResponse.view.sortBy,
            });
            createCsvFile(data, viewResponse.view.title || "Pending Channels");
          }}
        />
        <Button
          intercomTarget="refresh-table"
          buttonColor={ColorVariant.primary}
          icon={<RefreshIcon />}
          onClick={() => {
            track("Refresh Table", { page: "Channels Pending" });
            channelsResponse.refetch();
          }}
        />
        <Button
          intercomTarget="toggle-table-sidebar"
          onClick={() => {
            track("Toggle Table Sidebar", { page: "Channels Pending" });
            setSidebarExpanded(!sidebarExpanded);
          }}
          icon={<OptionsIcon />}
          id={"tableControlsButton"}
        >
          {t.Options}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllChannelPendingColumns}
      defaultView={DefaultPendingChannelsView}
      filterableColumns={FilterableChannelsPendingColumns}
      filterTemplate={ChannelsPendingFilterTemplate}
      sortableColumns={SortableChannelsPendingColumns}
      sortByTemplate={ChannelsPendingSortTemplate}
    />
  );

  const breadcrumbs = [
    <span key="b1">{t.channels}</span>,
    <Link key="b2" to={`/${Routes.CHANNELS}/${Routes.PENDING_CHANNELS}`}>
      {t.pendingChannels}
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      titleContent={""}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
    >
      <Table
        intercomTarget={"pending-channels-table"}
        cellRenderer={channelsPendingCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={channelsResponse.isLoading || channelsResponse.isFetching || channelsResponse.isUninitialized}
        maxRow={maxRow}
      />
    </TablePageTemplate>
  );
}

export default ChannelsPendingPage;

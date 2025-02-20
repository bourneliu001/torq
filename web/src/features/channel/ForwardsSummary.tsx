import {
  useGetChannelHistoryQuery,
  useGetChannelRebalancingQuery,
  useGetChannelOnChainCostQuery,
  useGetFlowQuery,
} from "apiSlice";
import type { GetChannelHistoryData, GetFlowQueryParams } from "types/api";
import classNames from "classnames";
import * as d3 from "d3";
import { addDays, format } from "date-fns";
import DetailsPageTemplate from "features/templates/detailsPageTemplate/DetailsPageTemplate";
import { useParams } from "react-router";
import { Link, Outlet } from "react-router-dom";
import { useAppSelector } from "store/hooks";
import Select from "components/forms/select/Select";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import styles from "./channel-page.module.scss";
import FlowChart, { FlowChartKeyOptions } from "./flowChart/FlowChart";
import ProfitsChart, { ProfitChartKeyOptions } from "./revenueChart/ProfitsChart";
import { selectActiveNetwork } from "features/network/networkSlice";
import { InputSizeVariant } from "components/forms/forms";
import useLocalStorage from "utils/useLocalStorage";
import { userEvents } from "utils/userEvents";
import { IsStringOption } from "utils/typeChecking";
import { TableControlsButtonGroup, TableControlSection } from "../templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import { ArrowSync20Regular as RefreshIcon } from "@fluentui/react-icons";

const ft = d3.format(",.0f");

function FowardsSummaryPage() {
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const currentPeriod = useAppSelector(selectTimeInterval);
  const { track } = userEvents();
  const [profitChartKey, setProfitChartKey] = useLocalStorage(`profitChartKey`, { value: "amount", label: "Amount" });
  const [flowChartKey, setFlowChartKey] = useLocalStorage(`flowChartKey`, { value: "amount", label: "Amount" });

  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  const { chanId } = useParams();

  const flowQueryParams: GetFlowQueryParams = {
    from: from,
    to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
    chanIds: chanId || "all",
    network: activeNetwork,
  };

  const { data, isLoading, refetch: flowRefetch } = useGetFlowQuery(flowQueryParams);

  const getChannelHistoryData: GetChannelHistoryData = {
    params: { chanId: chanId || "all" },
    queryParams: {
      from: from,
      to: format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd"),
      network: activeNetwork,
    },
  };

  const { data: onChainCost, refetch: onChainRefetch } = useGetChannelOnChainCostQuery(getChannelHistoryData);
  const { data: history, refetch: historyRefetch } = useGetChannelHistoryQuery(getChannelHistoryData);
  const { data: rebalancing, refetch: rebalanceRefetch } = useGetChannelRebalancingQuery(getChannelHistoryData);

  const profit: number =
    history?.revenueOut && onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? history?.revenueOut - onChainCost?.onChainCost - rebalancing?.rebalancingCost / 1000
      : 0;

  const totalCost: number =
    onChainCost?.onChainCost && rebalancing?.rebalancingCost
      ? onChainCost?.onChainCost + rebalancing?.rebalancingCost / 1000
      : 0;
  const historyRevenueOut = history?.revenueOut || 0;
  const rebalancingCostBy1000 = rebalancing?.rebalancingCost ? rebalancing?.rebalancingCost / 1000 : 0;
  const onchainCost = onChainCost?.onChainCost || 0;
  const historyReveueOutMinusCost = history?.revenueOut ? (history?.revenueOut - totalCost) / history?.revenueOut : 0;
  const historyAmountOut = history?.amountOut || 0;
  const historyCountOut = history?.countOut || 0;
  const historyRevenueOutAmountOut =
    history?.revenueOut && history?.amountOut ? (history?.revenueOut / history?.amountOut) * 1000 * 1000 : 0;

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"analyse"}>
      Summary
    </Link>,
  ];

  const refreshData = () => {
    flowRefetch();
    onChainRefetch();
    historyRefetch();
    rebalanceRefetch();
  };

  return (
    <DetailsPageTemplate title={"Forwards Summary"} breadcrumbs={breadcrumbs}>
      <TableControlSection intercomTarget={"table-page-controls"}>
        <TableControlsButtonGroup intercomTarget={"table-page-controls-left"}>
          <TimeIntervalSelect />
        </TableControlsButtonGroup>
        <TableControlsButtonGroup intercomTarget={"table-page-controls-right"}>
          <Button
            intercomTarget="refresh-page-data"
            buttonColor={ColorVariant.primary}
            icon={<RefreshIcon />}
            onClick={() => {
              track("Refresh forwards summary data", { page: "ForwardsSummary" });
              refreshData();
            }}
          />
        </TableControlsButtonGroup>
      </TableControlSection>
      <div className={styles.channelWrapper}>
        <div
          className={classNames(styles.pageRow, styles.channelSummary)}
          data-intercom-target={"forwards-summary-container"}
        >
          <div className={styles.shortColumn} data-intercom-target={"forwards-summary-stats"}>
            <div className={styles.card} data-intercom-target={"forwards-summary-stats-revenue-card"}>
              <div className={styles.heading}>Revenue</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Forwarding fees</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Channel Leases</div>
                <div className={classNames(styles.rowValue, styles.comingSoon)}>(Coming soon)</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
            </div>
            <div className={styles.card} data-intercom-target={"forwards-summary-stats-expenses-card"}>
              <div className={styles.heading}>Expenses</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Rebalancing</div>
                <div className={classNames(styles.rowValue)}>{ft(rebalancingCostBy1000)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Open & Close</div>
                <div className={classNames(styles.rowValue)}>{ft(onchainCost)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={classNames(styles.rowValue)}>{ft(totalCost)}</div>
              </div>
            </div>
            <div className={styles.card} data-intercom-target={"forwards-summary-stats-profit-card"}>
              <div className={styles.heading}>Profit</div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Total</div>
                <div className={classNames(styles.rowValue)}>{ft(profit)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Gross Profit Margin</div>
                <div className={classNames(styles.rowValue)}>{d3.format(".2%")(historyReveueOutMinusCost)}</div>
              </div>
            </div>
            <div className={styles.card} data-intercom-target={"forwards-summary-stats-transaction-card"}>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Revenue</div>
                <div className={styles.rowValue}>{ft(historyRevenueOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Amount</div>
                <div className={classNames(styles.rowValue)}>{ft(historyAmountOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Transactions</div>
                <div className={styles.rowValue}>{ft(historyCountOut)}</div>
              </div>
              <div className={styles.cardRow}>
                <div className={styles.rowLabel}>Average fee</div>
                <div className={classNames(styles.rowValue)}>{d3.format(",.1f")(historyRevenueOutAmountOut)}</div>
              </div>
            </div>
          </div>

          <div
            className={classNames(styles.card, styles.channelSummaryChart)}
            data-intercom-target={"forwards-summary-chart"}
          >
            <div className={styles.profitChartControls}>
              <div className={styles.profitChartLeftControls}>
                <Select
                  intercomTarget={"forwards-summary-chart-select"}
                  sizeVariant={InputSizeVariant.small}
                  value={profitChartKey}
                  onChange={(newValue) => {
                    if (IsStringOption(newValue)) {
                      track("Update ProfitChart Key", { oldKey: profitChartKey.value, key: newValue.value });
                      setProfitChartKey(newValue);
                    }
                  }}
                  options={ProfitChartKeyOptions}
                />
              </div>
              {/*<div className={styles.profitChartRightControls}>*/}
              {/*  <SettingsIcon />*/}
              {/*  Settings*/}
              {/*</div>*/}
            </div>
            <div className={styles.chartContainer}>
              {history && (
                <ProfitsChart
                  dataKey={profitChartKey.value}
                  data={history.history}
                  dashboard={true}
                  from={from}
                  to={to}
                />
              )}
            </div>
          </div>
        </div>

        <div className={styles.pageRow} data-intercom-target={"forwards-summary-flow-chart"}>
          <div className={styles.card}>
            <div className={styles.profitChartControls}>
              <div
                className={styles.profitChartLeftControls}
                data-intercom-target={"forwards-summary-flow-chart-select"}
              >
                <Select
                  intercomTarget={"forwards-summary-flow-chart-key-select"}
                  value={flowChartKey}
                  sizeVariant={InputSizeVariant.small}
                  onChange={(newValue) => {
                    if (IsStringOption(newValue)) {
                      track("Update FlowChart Key", { oldKey: flowChartKey.value, key: newValue.value });
                      setFlowChartKey(newValue);
                    }
                  }}
                  options={FlowChartKeyOptions}
                />
              </div>
              <div className={styles.profitChartRightControls}>
                {/*<Popover*/}
                {/*  button={<Button text={"Settings"} icon={<SettingsIcon />} hideMobileText={true} />}*/}
                {/*  className={"right"}*/}
                {/*>*/}
                {/*  Hello*/}
                {/*</Popover>*/}
              </div>
            </div>
            <div className="legendsContainer">
              <div className="sources">Sources</div>
              <div className="outbound">Outbound</div>
              <div className="inbound">Inbound</div>
              <div className="destinations">Destinations</div>
            </div>
            <div className={classNames(styles.chartWrapper, styles.flowChartWrapper)}>
              {!isLoading && data && <FlowChart flowKey={flowChartKey.value} data={data} />}
            </div>
          </div>
        </div>
      </div>
      <Outlet />
    </DetailsPageTemplate>
  );
}

export default FowardsSummaryPage;

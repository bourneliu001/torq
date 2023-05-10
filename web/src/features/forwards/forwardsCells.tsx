import styles from "components/table/cells/cell.module.scss";
import { ColumnMetaData } from "features/table/types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import { Forward } from "./forwardsTypes";
import TagsCell from "components/table/cells/tags/TagsCell";
import { GroupByOptions } from "features/viewManagement/types";
import ChannelCell from "components/table/cells/channelCell/ChannelCell";
import TextCell from "components/table/cells/text/TextCell";

export default function channelsCellRenderer(
  row: Forward,
  rowIndex: number,
  column: ColumnMetaData<Forward>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Forward,
  groupedBy?: GroupByOptions
): JSX.Element {
  if (column.key === "alias" && !isTotalsRow) {
    return (
      <ChannelCell
        color={row["torqNodeCssColour"] as string}
        alias={row["alias"] as string}
        channelId={row.channelId}
        nodeId={row.peerNodeId}
        open={row["open"]}
        key={"alias" + rowIndex + columnIndex}
        className={column.locked ? styles.locked : ""}
        hideActionButtons={groupedBy === "peer"}
      />
    );
  }

  if (column.key === "alias" && isTotalsRow) {
    return <TextCell text={"Total"} key={"alias" + rowIndex + columnIndex} className={styles.locked} />;
  }

  if (column.key === "tags") {
    return (
      <TagsCell
        channelTags={row.channelTags}
        peerTags={row.peerTags}
        key={"tags" + rowIndex + columnIndex}
        channelId={row.channelId}
        nodeId={row.peerNodeId}
        totalCell={isTotalsRow}
        displayChannelTags={groupedBy !== "peer"}
      />
    );
  }

  // Use the defualt
  return DefaultCellRenderer(row, rowIndex, column, columnIndex, isTotalsRow, maxRow);
}

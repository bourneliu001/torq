import { ColumnMetaData } from "features/table/types";
import { Peer } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import TextCell from "components/table/cells/text/TextCell";
import PeersAliasCell from "components/table/cells/peersCell/PeersAliasCell";
import TagsCell from "components/table/cells/tags/TagsCell";

export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  if (column.key === "peerAlias") {
    return (
      <PeersAliasCell
        key={column.key.toString() + rowIndex}
        alias={row.peerAlias}
        peerNodeId={row.nodeId}
        torqNodeId={row.torqNodeId}
        connectionStatus={row.connectionStatus}
        color={row.nodeCssColour}
      />
    );
  }
  if (column.key === "tags") {
    return (
      <TagsCell
        channelTags={[]}
        peerTags={row.tags}
        key={"tagsCell" + rowIndex}
        nodeId={row.nodeId}
        displayChannelTags={false}
      />
    );
  }

  if (column.key === "connectionStatus") {
    return <TextCell text={row.connectionStatus} key={column.key.toString() + rowIndex} totalCell={isTotalsRow} />;
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}

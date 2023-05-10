import cellStyles from "components/table/cells/cell.module.scss";
import classNames from "classnames";
import { ReactNode } from "react";
import { CellRendererFunction, ColumnMetaData } from "./types";
import FirstTableCellWrapper from "./firstTableCellRenderer";
import { GroupByOptions } from "features/viewManagement/types";

export type RowProp<T> = {
  row: T;
  rowIndex: number;
  columns: Array<ColumnMetaData<T>>;
  selectable?: boolean;
  selected: boolean;
  cellRenderer: CellRendererFunction<T>;
  isTotalsRow?: boolean;
  maxRow?: T;
  groupedBy: GroupByOptions;
};

function Row<T>(props: RowProp<T>) {
  // const totalsRowRenderer = props.totalsRowRenderer ? props.totalsRowRenderer : defaultTotalsRowRenderer;

  // Adds empty cells at the start and end of each row. This is to give the table a buffer at each end.
  const rowContent: Array<ReactNode> = [];
  rowContent.push(<FirstTableCellWrapper key={"first-cell-" + props.rowIndex} />);

  rowContent.push(
    ...props.columns.map((columnMeta: ColumnMetaData<T>, columnIndex) => {
      return props.cellRenderer(
        props.row,
        props.rowIndex,
        columnMeta,
        columnIndex,
        props.isTotalsRow,
        props.maxRow,
        props.groupedBy
      );
    })
  );

  rowContent.push(
    <div
      className={classNames(
        cellStyles.cell,
        cellStyles.empty,
        {
          [cellStyles.lastTotalCell]: props.isTotalsRow,
        },
        cellStyles.lastEmptyCell
      )}
      key={"last-cell-" + props.rowIndex}
    />
  );

  return (
    <div
      className={classNames(cellStyles.tableRow, "torq-row-" + props.rowIndex, {
        [cellStyles.totalsRow]: props.isTotalsRow,
      })}
      key={"torq-row-" + props.rowIndex}
    >
      {rowContent}
    </div>
  );
}

export default Row;

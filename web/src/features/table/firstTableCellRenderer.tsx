import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";

type FirstTableCellRendererProps = {
  isTotalsRow?: boolean;
  children?: JSX.Element;
};

export default function FirstTableCellWrapper(props: FirstTableCellRendererProps): JSX.Element {
  return (
    <div
      className={classNames(cellStyles.cell, cellStyles.empty, cellStyles.firstEmptyCell, cellStyles.locked, {
        [cellStyles.totalCell]: props.isTotalsRow,
        [cellStyles.hasContent]: !!props.children,
      })}
    >
      {props.children}
    </div>
  );
}

import styles from "components/table/cells/cell.module.scss";
import classNames from "classnames";

export default function EmptyCell(index?: number | string, className?: string) {
  return <div className={classNames(styles.cell, styles.empty, className)} key={"last-cell-" + index} />;
}

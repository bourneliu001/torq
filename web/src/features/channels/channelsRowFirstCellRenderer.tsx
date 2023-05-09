import { channel } from "./channelsTypes";
import styles from "./channels.module.scss";

//table has a first cell that is empty, this is to make customise the first cell
export default function channelsRowFirstCellRenderer(row: channel): JSX.Element {
  if (row) {
    return (
      <div
        className={styles.firstCell}
        style={{
          backgroundColor: row.nodeCssColour,
        }}
      ></div>
    );
  }

  return <></>;
}

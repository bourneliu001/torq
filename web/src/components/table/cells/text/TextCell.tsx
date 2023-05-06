import React from "react";
import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./text_cell.module.scss";

export type TextCellProps = {
  text: string | Array<string>;
  link?: string;
  copyText?: string;
  className?: string;
  totalCell?: boolean;
};

const TextCell = (props: TextCellProps) => {
  const textArray = Array.isArray(props.text) ? props.text : [props.text];
  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.textCell,
        { [cellStyles.totalCell]: props.totalCell },
        props.className
      )}
    >
      {!props.totalCell && (
        <div>
          {textArray.map((text, i) => (
            <div key={i}>
              <span className={classNames(styles.content)}>{text}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

const TextCellMemo = React.memo(TextCell);
export default TextCellMemo;

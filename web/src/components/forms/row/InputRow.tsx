import React from "react";
import { Children } from "react";
import classNames from "classnames";
import styles from "./input_row.module.scss";

type InputRowProps = {
  className?: string;
  button?: React.ReactNode;
  children: React.ReactNode;
};

export default function InputRow({ className, button, children }: InputRowProps) {
  const childrenCount = Children.count(children);
  const gridColumns = {
    gridTemplateColumns: `${Array(childrenCount).fill("1fr").join(" ")} ${button ? "min-content" : ""}`,
  };

  return (
    <div className={classNames(styles.inputRowWrapper, className)} style={gridColumns}>
      {(Children.toArray(children) || []).map((child, index) => {
        return (
          <div className={styles.inputRowItem} key={"input-row-item-" + index}>
            {child}
          </div>
        );
      })}
      {button && <div className={styles.inputRowButton}>{button}</div>}
    </div>
  );
}

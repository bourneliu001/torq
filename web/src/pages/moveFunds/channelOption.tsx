import classNames from "classnames";
import { format } from "d3";
import { ArrowStepIn16Regular, ArrowStepOut16Regular } from "@fluentui/react-icons";
import styles from "./channelOption.module.scss";

type ChannelOptionProps = {
  remoteBalance: number;
  capacity: number;
  localBalance: number;
  shortChannelId: string;
};

const formatAmount = (amount: number) => format(",.0f")(amount);

const calculateAvailableBalance = (local: number, remote: number) => {
  const calculate = local / (local + remote);
  return format(".2%")(calculate);
};

export default function ChannelOption(props: ChannelOptionProps) {
  return (
    <div className={styles.channelOptionWrapper}>
      <div className={styles.shortChannelId}>{props.shortChannelId}</div>
      <div className={styles.detailsWrapper}>
        <div className={classNames(styles.balances)}>
          <div className={classNames(styles.amount)}>
            <div className={classNames(styles.icon)}>
              <ArrowStepIn16Regular />
            </div>
            <div className={classNames(styles.value)}>{formatAmount(props.remoteBalance)}</div>
          </div>
          <div className={classNames(styles.amount)}>
            <div className={classNames(styles.value)}>{formatAmount(props.localBalance)}</div>
            <div className={classNames(styles.icon)}>
              <ArrowStepOut16Regular />
            </div>
          </div>
        </div>
        <div className={classNames(styles.bar)}>
          <div
            className={classNames(styles.percentage)}
            style={{ width: calculateAvailableBalance(props.localBalance, props.remoteBalance) }}
          />
        </div>
        <div className={classNames(styles.capacity)}>
          <div className={classNames(styles.remote)}>
            {calculateAvailableBalance(props.remoteBalance, props.localBalance)}
          </div>
          <div className={classNames(styles.total)}>{formatAmount(props.capacity)}</div>
          <div className={classNames(styles.local)}>
            {calculateAvailableBalance(props.localBalance, props.remoteBalance)}
          </div>
        </div>
      </div>
    </div>
  );
}

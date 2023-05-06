import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./newPayments.module.scss";
import { DecodedInvoice } from "types/api";
import classNames from "classnames";
import { InvoiceStatusType, NewPaymentResponse } from "features/transact/Payments/paymentTypes";
import { useEffect, useState } from "react";
import { format } from "d3";
import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";
import useTranslations from "services/i18n/useTranslations";
import {
  DetailsContainer,
  DetailsRow,
  DetailsRowLinkAndCopy,
} from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";

const f = format(",.0f");

type InvoicePaymentResponseProps = {
  responses: Array<NewPaymentResponse>;
  paymentProcessingError: string;
  selectedNodeId: number;
  decodedInvoice: DecodedInvoice;
  destination: string;
  clearPaymentFlow: () => void;
  amount: number;
};

export function InvoicePaymentResponse(props: InvoicePaymentResponseProps) {
  const { t } = useTranslations();
  const [status, setStatus] = useState<InvoiceStatusType>();

  useEffect(() => {
    if (props.paymentProcessingError !== "") {
      setStatus("FAILED");
    } else if (props.responses !== undefined && props.responses.length !== 0) {
      setStatus(props.responses[props.responses.length - 1].paymentStatus);
    } else {
      setStatus("IN_FLIGHT");
    }
  }, [props.responses.length, props.paymentProcessingError]);

  return (
    <ProgressTabContainer>
      {status === "SUCCEEDED" && (
        <div className={classNames(styles.paymentStatusMessage)}>
          <div className={styles.amountPaid}>{`${f(props.amount)} sats`}</div>
          <div className={styles.amountPaidText}>{`Sent to ${props.decodedInvoice.nodeAlias}`}</div>
        </div>
      )}
      <StatusIcon state={status === "SUCCEEDED" ? "success" : status === "FAILED" ? "error" : "processing"} />
      {props.paymentProcessingError && (
        <div className={classNames(styles.paymentStatusMessage)}>{props.paymentProcessingError}</div>
      )}
      {status === "SUCCEEDED" && (
        <div>
          <DetailsContainer>
            <DetailsRow label={"Fee:"}>{props.responses[props.responses.length - 1].feePaidMsat / 1000}</DetailsRow>
          </DetailsContainer>
          <DetailsContainer>
            <DetailsRowLinkAndCopy label={"Preimage:"} copy={props.responses[props.responses.length - 1].preimage}>
              {props.responses[props.responses.length - 1].preimage}
            </DetailsRowLinkAndCopy>
          </DetailsContainer>
        </div>
      )}
      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        rightChildren={
          <Button
            intercomTarget={"new-payment-restart-button"}
            onClick={() => {
              props.clearPaymentFlow();
            }}
            buttonColor={ColorVariant.primary}
          >
            {t.newPayment}
          </Button>
        }
      />
    </ProgressTabContainer>
  );
}

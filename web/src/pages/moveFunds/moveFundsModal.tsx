import { ArrowSwap20Regular as ModalIcon } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery } from "apiSlice";
// import Button, { ButtonPosition, ColorVariant, SizeVariant } from "components/buttons/Button";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { Form, Input, RadioChips } from "components/forms/forms";
import { NumberFormatValues } from "react-number-format";
import useLocalStorage from "utils/useLocalStorage";
// import { useAppSelector } from store/hooks";
// import { selectActiveNetwork } from features/network/networkSlice";
// import Note, { NoteType } from "features/note/Note";
// import ToastContext from "features/toast/context";
// import { toastCategory } from "features/toast/Toasts";
// import Spinny from "features/spinny/Spinny";
// import { ServerErrorType } from "components/errors/errors";
// import ErrorSummary from "components/errors/ErrorSummary";
// import { userEvents } from "utils/userEvents";

function moveFundsModal() {
  const { t } = useTranslations();
  // const { track } = userEvents();
  // const toastRef = useContext(ToastContext);
  // const activeNetwork = useAppSelector(selectActiveNetwork);
  // const { data: nodes } = useGetNodesInformationByCategoryQuery(activeNetwork);
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<Array<SelectOptions>>();
  const [selectedFromNodeId, setSelectedFromNodeId] = useLocalStorage<SelectOptions | undefined>(
    "moveFundsFrom",
    undefined
  );
  const [selectedToNodeId, setSelectedToNodeId] = useLocalStorage<SelectOptions | undefined>("moveFundsTo", undefined);
  const [amount, setAmount] = useLocalStorage<number>("moveFundsAmount", 0);
  const [moveChain, setMoveChain] = useLocalStorage<string>("moveFundsChain", "move-funds-off-chain");
  // const [fromNodeBalance, setFromNodeBalance] = useState(0);
  // const [toNodeBalance, setToNodeBalance] = useState(0);

  interface Option {
    label: string;
    value: number;
  }

  useEffect(() => {
    if (nodeConfigurations) {
      const options = nodeConfigurations.map((node: nodeConfiguration) => {
        return { label: node.name, value: node.nodeId };
      });
      setNodeConfigurationOptions(options);
      setSelectedFromNodeId(options[0].value);
      if (options?.length >= 2) setSelectedToNodeId(options[1].value);
    }
  }, [nodeConfigurations]);

  // useEffect(() => {
  //   // if nodes is missing return
  //   if (!nodes?.length) return;
  //   if (selectedFromNodeId) {
  //     setFromNodeBalance(nodes.find((n) => n.nodeId === selectedFromNodeId).);
  //   }
  //   if (selectedToNodeId) {
  //     setToNodeBalance(0);
  //   }
  // }, [nodes, moveChain, selectedFromNodeId, selectedToNodeId]);

  // const handleClickNext = () => {
  //   console.log("next");
  // };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={t.moveFunds} show={true} onClose={() => navigate(-1)} icon={<ModalIcon />}>
      <Form intercomTarget={"move-funds-form"}>
        <RadioChips
          groupName={"move-funds-chain-select"}
          // helpText={t.channelBalanceEventFilterNode.ignoreWhenEventlessHelpText}
          options={[
            {
              label: t.offChainTx,
              id: "move-funds-off-chain",
              checked: moveChain === "move-funds-off-chain",
              onChange: (e) => {
                setMoveChain(e.target.id);
              },
            },
            {
              label: t.onChainTx,
              id: "move-funds-on-chain",
              checked: moveChain === "move-funds-on-chain",
              onChange: (e) => {
                setMoveChain(e.target.id);
              },
            },
          ]}
        />
        <Select
          intercomTarget={"move-funds-from-input"}
          label={t.from}
          onChange={(newValue: unknown) => {
            const value = newValue as Option;
            if (value && value.value != 0) {
              setSelectedFromNodeId(value.value);
            }
          }}
          options={nodeConfigurationOptions}
          value={nodeConfigurationOptions?.find((option) => option.value === selectedFromNodeId)}
        />
        <Select
          intercomTarget={"move-funds-to-input"}
          label={t.to}
          onChange={(newValue: unknown) => {
            const value = newValue as Option;
            if (value && value.value != 0) {
              setSelectedToNodeId(value.value);
            }
          }}
          options={nodeConfigurationOptions}
          value={nodeConfigurationOptions?.find((option) => option.value === selectedToNodeId)}
        />
        <Input
          label={t.amount}
          intercomTarget={"move-funds-amount-input"}
          formatted={true}
          className={styles.single}
          thousandSeparator={","}
          value={amount}
          suffix={" sat"}
          onValueChange={(values: NumberFormatValues) => {
            setAmount(values.floatValue as number);
          }}
        />
      </Form>
      <div className={styles.addressTypeWrapper}></div>
      <div className={styles.addressResultWrapper}></div>
    </PopoutPageTemplate>
  );
}

export default moveFundsModal;

import {
  Options20Regular as OptionsIcon,
  ArrowSwap20Regular as ModalIcon,
  ArrowExportUp20Regular as MaxIcon,
} from "@fluentui/react-icons";
import { useGetChannelsQuery, useGetNodeConfigurationsQuery, useGetNodesWalletBalancesQuery } from "apiSlice";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { Form, Input, InputRow, RadioChips, Switch } from "components/forms/forms";
import useLocalStorage from "utils/useLocalStorage";
import { channel } from "features/channels/channelsTypes";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import Button, { ButtonWrapper, ColorVariant } from "components/buttons/Button";
import { components, OptionProps, SingleValueProps } from "react-select";
import ChannelOption from "./channelOption";
import { NumberFormatValues } from "react-number-format";
import { format } from "d3";
import { userEvents } from "utils/userEvents";
import ErrorSummary from "components/errors/ErrorSummary";
import { FormErrors, mergeServerError } from "components/errors/errors";
import { useMoveFundsOffChainMutation, useMoveOnChainFundsMutation } from "./moveFundsApi";
import { IsNumericOption, IsServerErrorResult } from "utils/typeChecking";
import { AddressType } from "./moveFundsTypes";
import { SectionContainer } from "features/section/SectionContainer";

const formatAmount = (amount: number) => format(",.0f")(amount);

type ChannelOption = {
  value: number;
  label: string;
  remoteBalance?: number;
  localBalance?: number;
  capacity?: number;
};

function IsChannelOption(result: unknown): result is ChannelOption {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as { value: unknown; label: string }).value === "number"
  );
}

function IsAddressTypeOption(result: unknown): result is SelectOptions {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as { value: unknown; label: string }).value === "number"
  );
}

function moveFundsModal() {
  const { t } = useTranslations();
  const { track } = userEvents();
  // Create options for the address type select
  const addressTypeOptions: Array<SelectOptions> = [
    { label: t.p2wpkh, value: AddressType.P2WPKH }, // Segwit
    { label: t.p2tr, value: AddressType.P2TR }, // Taproot
    { label: t.p2wkh, value: AddressType.P2WKH }, // Wrapped Segwit
  ];
  const [formErrorState, setFormErrorState] = useState<FormErrors>({});
  // const toastRef = useContext(ToastContext);
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { data: nodesWalletBalances } = useGetNodesWalletBalancesQuery(activeNetwork);
  const navigate = useNavigate();
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork });
  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<Array<SelectOptions>>();
  const [selectedFromNodeId, setSelectedFromNodeId] = useLocalStorage<SelectOptions | undefined>(
    "moveFundsFrom",
    undefined
  );
  const [selectedToNodeId, setSelectedToNodeId] = useLocalStorage<SelectOptions | undefined>("moveFundsTo", undefined);
  const [amount, setAmount] = useLocalStorage<number>("moveFundsAmount", 0);
  const [moveChain, setMoveChain] = useLocalStorage<string>("moveFundsChain", "move-funds-off-chain");
  const [channelOptions, setChannelOptions] = useState<Array<ChannelOption>>();
  const [selectedChannelId, setSelectedChannelId] = useLocalStorage<number | undefined>("moveFundsChannel", undefined);
  const [maxAmount, setMaxAmount] = useState<number>(0);
  // OnChain Options
  const [targetConf, setTargetConf] = useLocalStorage<number | undefined>("moveFundsTargetConf", undefined);
  const [satPerVbyte, setSatPerVbyte] = useLocalStorage<number | undefined>("moveFundsSatPerVbyte", undefined);
  const [spendUnconfirmed, setSpendUnconfirmed] = useLocalStorage<boolean>("moveFundsSpendUnconfirmed", false);
  const [minConf, setMinConf] = useLocalStorage<number | undefined>("moveFundsMinConf", undefined);
  const [addressType, setAddressType] = useLocalStorage<AddressType>("moveFundsAddressType", AddressType.P2WPKH);
  const [sendAll, setSendAll] = useLocalStorage<boolean>("moveFundsSendAll", false);

  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState<boolean>(true);

  const [moveOffChainFunds, { error: offChainErrors }] = useMoveFundsOffChainMutation();
  const [moveOnChainFunds, { error: onChainErrors }] = useMoveOnChainFundsMutation();

  interface Option {
    label: string;
    value: number;
  }

  function handleSwapNodes(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
    e.preventDefault();
    const temp = selectedFromNodeId;
    track("Move Funds Node Swap", {
      oldFrom: selectedFromNodeId,
      oldTo: selectedToNodeId,
      newFrom: selectedToNodeId,
      newTo: temp,
    });
    setSelectedFromNodeId(selectedToNodeId);
    setSelectedToNodeId(temp);
  }

  useEffect(() => {
    if (nodeConfigurations) {
      const options = nodeConfigurations.map((node: nodeConfiguration) => {
        return { label: node.name, value: node.nodeId };
      });
      setNodeConfigurationOptions(options);
      if (options?.length >= 1 && !selectedFromNodeId && !options.find((c) => c.value === selectedFromNodeId)) {
        setSelectedFromNodeId(options[0].value);
      }
      if (options?.length >= 2 && !selectedToNodeId && !options.find((c) => c.value === selectedToNodeId)) {
        setSelectedToNodeId(options[1].value);
      }
    }
  }, [nodeConfigurations]);

  useEffect(() => {
    if (channelsResponse?.data) {
      const options: Array<ChannelOption> = channelsResponse.data
        .filter((c) => [selectedFromNodeId].includes(c.peerNodeId))
        .map((channel: channel) => {
          return {
            label: channel.shortChannelId,
            value: channel.channelId,
            remoteBalance: channel.remoteBalance,
            localBalance: channel.localBalance,
            capacity: channel.capacity,
          };
        })
        // Sort by shortChannelId
        .sort((a: Option, b: Option) => {
          if (a.label > b.label) return 1;
          if (a.label < b.label) return -1;
          return 0;
        });
      setChannelOptions(options);
      if (options?.length >= 1 && !selectedChannelId && !options.find((c) => c.value === selectedChannelId)) {
        setSelectedChannelId(options[0].value);
      }
    }
  }, [channelsResponse?.data, selectedFromNodeId]);

  useEffect(() => {
    if (moveChain === "move-funds-off-chain") {
      setMaxAmount(channelOptions?.find((c) => c.value === selectedChannelId)?.localBalance || 0);
    } else {
      const walletBalance =
        nodesWalletBalances?.find((w) => w.request.nodeId === selectedFromNodeId)?.confirmedBalance || 0;
      setMaxAmount(walletBalance);
    }
  }, [selectedChannelId, channelOptions, moveChain, nodesWalletBalances]);

  // Merge the errors from offchain into the form errors using offChainErrors
  useEffect(() => {
    if (IsServerErrorResult(offChainErrors)) {
      setFormErrorState(mergeServerError(offChainErrors.data, formErrorState));
    }
    if (IsServerErrorResult(onChainErrors) && onChainErrors.data) {
      setFormErrorState(mergeServerError(onChainErrors.data, formErrorState));
    }
  }, [offChainErrors, onChainErrors]);

  const SingleValue = (props: SingleValueProps<unknown>) => {
    const channel = props.data as ChannelOption;
    return (
      <components.SingleValue {...props}>
        <ChannelOption
          shortChannelId={channel?.label || ""}
          localBalance={channel?.localBalance || 0}
          remoteBalance={channel?.remoteBalance || 0}
          capacity={channel?.capacity || 0}
        />
      </components.SingleValue>
    );
  };

  const Option = (props: OptionProps) => {
    const channel = props.data as ChannelOption;
    return (
      <div>
        <components.Option {...props}>
          <ChannelOption
            shortChannelId={channel?.label || ""}
            localBalance={channel?.localBalance || 0}
            remoteBalance={channel?.remoteBalance || 0}
            capacity={channel?.capacity || 0}
          />
        </components.Option>
      </div>
    );
  };

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (moveChain === "move-funds-off-chain") {
      track("Move Off-Chain Funds", {
        moveFundsFrom: selectedFromNodeId,
        moveFundsTo: selectedToNodeId,
        moveFundsAmountMsat: amount * 1000,
        moFundsChannel: selectedChannelId,
      });
      moveOffChainFunds({
        outgoingNodeId: selectedFromNodeId,
        incomingNodeId: selectedToNodeId,
        channelId: selectedChannelId,
        amountMsat: amount * 1000,
      }).then(() => {
        setFormErrorState({} as FormErrors);
      });
    } else if (moveChain === "move-funds-on-chain") {
      track("Move On-Chain Funds", {
        moveFundsFrom: selectedFromNodeId,
        moveFundsTo: selectedToNodeId,
        moveFundsAmountMsat: amount,
        moveFundsTargetConf: targetConf,
        moveFundsSatPerByte: satPerVbyte,
        moveFundsSpendUnconfirmed: spendUnconfirmed,
        moveFundsMinConf: minConf,
        moveFundsAddressType: addressType,
        moveFundsSendAll: sendAll,
      });
      moveOnChainFunds({
        outgoingNodeId: selectedFromNodeId,
        incomingNodeId: selectedToNodeId,
        amountMsat: amount * 1000,
        targetConf: targetConf,
        satPerVbyte: satPerVbyte,
        spendUnconfirmed: spendUnconfirmed,
        minConf: minConf,
        addressType: addressType,
        sendAll: sendAll,
      }).then(() => {
        setFormErrorState({} as FormErrors);
      });
    }
  }

  return (
    <PopoutPageTemplate title={t.moveFunds} show={true} onClose={() => navigate(-1)} icon={<ModalIcon />}>
      <Form intercomTarget={"move-funds-form"} onSubmit={handleSubmit}>
        <RadioChips
          groupName={"move-funds-chain-select"}
          options={[
            {
              label: t.offChainTx,
              id: "move-funds-off-chain",
              checked: moveChain === "move-funds-off-chain",
              onChange: (e) => {
                setMoveChain(e.target.id);
                track("Move Funds Chain Selected", { chain: "off-chain" });
              },
            },
            {
              label: t.onChainTx,
              id: "move-funds-on-chain",
              checked: moveChain === "move-funds-on-chain",
              onChange: (e) => {
                setMoveChain(e.target.id);
                track("Move Funds Chain Selected", { chain: "on-chain" });
              },
            },
          ]}
        />
        <InputRow
          button={
            <Button
              intercomTarget={"move-funds-swap-nodes-button"}
              onClick={handleSwapNodes}
              type={"button"}
              icon={<ModalIcon />}
            />
          }
        >
          <Select
            intercomTarget={"move-funds-from-input"}
            label={t.from}
            onChange={(newValue: unknown) => {
              if (IsNumericOption(newValue)) {
                setSelectedFromNodeId(newValue.value);
              }
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions?.find((option) => option.value === selectedFromNodeId)}
          />
          <Select
            intercomTarget={"move-funds-to-input"}
            label={t.to}
            onChange={(newValue: unknown) => {
              if (IsNumericOption(newValue)) {
                setSelectedToNodeId(newValue.value);
              }
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions?.find((option) => option.value === selectedToNodeId)}
          />
        </InputRow>
        {moveChain === "move-funds-off-chain" && (
          <Select
            selectComponents={{ Option, SingleValue }}
            intercomTarget={"move-funds-channel-input"}
            label={t.channel}
            onChange={(newValue: unknown) => {
              if (IsChannelOption(newValue)) {
                setSelectedChannelId(newValue.value);
              }
            }}
            options={channelOptions}
            value={channelOptions?.find((option) => option.value === selectedChannelId)}
          />
        )}
        <Input
          label={t.amount}
          name={"amount"}
          intercomTarget={"move-funds-amount-input"}
          formatted={true}
          className={styles.single}
          disabled={sendAll && moveChain === "move-funds-on-chain"}
          thousandSeparator={","}
          value={amount}
          suffix={" sat"}
          onValueChange={(values: NumberFormatValues) => {
            setAmount(values.floatValue as number);
          }}
          errors={formErrorState}
          infoText={"Maximum amount is: " + formatAmount(maxAmount) + " sat"}
          button={
            <Button
              buttonColor={ColorVariant.primary}
              disabled={sendAll && moveChain === "move-funds-on-chain"}
              intercomTarget={"move-funds-max-button"}
              icon={<MaxIcon />}
              onClick={() => {
                track("Move Funds Max Clicked", { amount: maxAmount, chain: moveChain });
                setAmount(maxAmount);
              }}
            />
          }
        />
        {moveChain === "move-funds-on-chain" && (
          <>
            <Switch
              intercomTarget={"move-funds-send-all"}
              label={t.SendEverything}
              checked={sendAll}
              onChange={(e) => {
                setSendAll(e.target.checked);
                if (e.target.checked) {
                  setAmount(0);
                }
              }}
            />
            <SectionContainer
              title={t.AdvancedOptions}
              icon={OptionsIcon}
              expanded={expandAdvancedOptions}
              handleToggle={() => {
                setExpandAdvancedOptions(!expandAdvancedOptions);
              }}
            >
              <InputRow>
                <Input
                  label={t.SatPerVbyte}
                  name={"satPerVbyte"}
                  formatted={true}
                  intercomTarget={"move-funds-sat-per-vbyte-input"}
                  value={satPerVbyte}
                  disabled={targetConf !== undefined}
                  suffix={" sat/vByte"}
                  onValueChange={(values: NumberFormatValues) => {
                    if (values.floatValue === 0) {
                      setSatPerVbyte(undefined);
                    } else {
                      setSatPerVbyte(values.floatValue);
                    }
                  }}
                  errors={formErrorState}
                />
                <Input
                  label={t.TargetConfirmations}
                  name={"targetConf"}
                  intercomTarget={"move-funds-target-conf-input"}
                  value={targetConf}
                  disabled={satPerVbyte !== undefined}
                  suffix={" blocks"}
                  formatted={true}
                  onValueChange={(values: NumberFormatValues) => {
                    if (values.floatValue === 0) {
                      setTargetConf(undefined);
                    } else {
                      setTargetConf(values.floatValue);
                    }
                  }}
                  errors={formErrorState}
                />
              </InputRow>
              <Select
                intercomTarget={"move-funds-on-chain-select-address-type"}
                label={t.addressType}
                options={addressTypeOptions}
                value={addressTypeOptions?.find((option) => option.value === addressType)}
                onChange={(newValue: unknown) => {
                  if (IsAddressTypeOption(newValue)) {
                    setAddressType(newValue.value);
                  }
                }}
              />
              <Input
                label={t.MinimumConfirmations}
                intercomTarget={"move-funds-min-conf-input"}
                value={minConf}
                type={"number"}
                onChange={(e) => {
                  if (Number(e.target.value) === 0) {
                    setMinConf(undefined);
                  } else {
                    setMinConf(Number(e.target.value));
                  }
                }}
              />
              <Switch
                label={t.SpendUnconfirmed}
                intercomTarget={"move-funds-spend-unconfirmed-switch"}
                value={spendUnconfirmed}
                onChange={(e) => {
                  setSpendUnconfirmed(e.target.checked);
                }}
              />
            </SectionContainer>
          </>
        )}
        <ButtonWrapper
          rightChildren={
            <Button
              buttonColor={ColorVariant.success}
              intercomTarget={"move-funds-confirm-button"}
              type={"submit"}
              // loading={loading}
              // disabled={loading}
            >
              {t.confirm}
            </Button>
          }
        />
        <ErrorSummary errors={formErrorState} />
      </Form>
      <div className={styles.addressTypeWrapper}></div>
      <div className={styles.addressResultWrapper}></div>
    </PopoutPageTemplate>
  );
}

export default moveFundsModal;

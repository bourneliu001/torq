import { ArrowSwap20Regular as ModalIcon, ArrowExportUp20Regular as MaxIcon } from "@fluentui/react-icons";
import { useGetChannelsQuery, useGetNodeConfigurationsQuery, useGetNodesWalletBalancesQuery } from "apiSlice";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { Form, Input, InputRow, RadioChips } from "components/forms/forms";
import useLocalStorage from "utils/useLocalStorage";
import { channel } from "features/channels/channelsTypes";
import { useAppSelector } from "store/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import Button, { ButtonWrapper, ColorVariant } from "components/buttons/Button";
import { components, OptionProps, SingleValueProps } from "react-select";
import ChannelOption from "./channelOption";
import { NumberFormatValues } from "react-number-format";
import { format } from "d3";
import { userEvents } from "../../utils/userEvents";
import ErrorSummary from "../../components/errors/ErrorSummary";
import { FormErrors } from "../../components/errors/errors";

const formatAmount = (amount: number) => format(",.0f")(amount);

type ChannelOption = {
  value: number;
  label: string;
  remoteBalance?: number;
  localBalance?: number;
  capacity?: number;
};

export function IsChannelOption(result: unknown): result is ChannelOption {
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
      setSelectedFromNodeId(options[0].value);
      if (options?.length >= 2) setSelectedToNodeId(options[1].value);
    }
  }, [nodeConfigurations]);

  useEffect(() => {
    if (channelsResponse?.data) {
      const options: Array<ChannelOption> = channelsResponse.data
        .filter((c) => [selectedFromNodeId].includes(c.nodeId))
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
      setSelectedChannelId(options[0].value);
    }
  }, [channelsResponse?.data?.length, selectedFromNodeId]);

  useEffect(() => {
    if (moveChain === "move-funds-off-chain") {
      setMaxAmount(channelOptions?.find((c) => c.value === selectedChannelId)?.localBalance || 0);
      setAmount(0);
    } else {
      const walletBalance =
        nodesWalletBalances?.find((w) => w.request.nodeId === selectedFromNodeId)?.confirmedBalance || 0;
      setMaxAmount(walletBalance);
      setAmount(0);
    }
  }, [selectedChannelId, channelOptions, moveChain, nodesWalletBalances?.length]);

  useEffect(() => {
    if (amount > maxAmount) {
      setFormErrorState({
        server: [{ description: t.amountExceedsMax, attributes: { amount: "" } }],
        fields: {
          amount: [t.amountExceedsMax],
        },
      });
    } else {
      setFormErrorState({});
    }
  }, [amount, maxAmount]);

  const SingleValue = ({ ...props }: SingleValueProps<unknown>) => {
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
    // TODO: Add track event
    console.log("submit");
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
              const value = newValue as Option;
              if (value && value.value != 0) {
                setSelectedFromNodeId(value.value);
              }
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions?.find((option) => option.value === selectedFromNodeId)}
          />
          {/* TODO: Add a flip button */}
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
          intercomTarget={"move-funds-amount-input"}
          formatted={true}
          className={styles.single}
          thousandSeparator={","}
          value={amount}
          suffix={" sat"}
          onValueChange={(values: NumberFormatValues) => {
            setAmount(values.floatValue as number);
          }}
          infoText={"Maximum amount is: " + formatAmount(maxAmount) + " sat"}
          errorText={formErrorState?.fields?.amount.toString()}
          button={
            <Button
              buttonColor={ColorVariant.primary}
              intercomTarget={"move-funds-max-button"}
              icon={<MaxIcon />}
              onClick={() => {
                track("Move Funds Max Clicked", { amount: maxAmount, chain: moveChain });
                setAmount(maxAmount);
              }}
            />
          }
        />
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

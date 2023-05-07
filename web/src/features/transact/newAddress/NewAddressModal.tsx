import { Copy20Regular as CopyIcon, Link20Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery } from "apiSlice";
import Button, { ButtonPosition, ColorVariant, SizeVariant } from "components/buttons/Button";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select, { SelectOptions } from "features/forms/Select";
import { useNewAddressMutation } from "./newAddressApi";
import { AddressType } from "./newAddressTypes";
import Note, { NoteType } from "features/note/Note";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import Spinny from "features/spinny/Spinny";
import { ServerErrorType } from "components/errors/errors";
import ErrorSummary from "components/errors/ErrorSummary";
import { userEvents } from "utils/userEvents";

function NewAddressModal() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const toastRef = useContext(ToastContext);
  const [nodeConfigurationOptions, setNodeConfigurationOptions] = useState<Array<SelectOptions>>();

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

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
      setSelectedNodeId(options[0].value);
    }
  }, [nodeConfigurations]);

  const addressTypeOptions = [
    { label: t.p2wpkh, value: AddressType.P2WPKH }, // Wrapped Segwit
    { label: t.p2wkh, value: AddressType.P2WKH }, // Segwit
    { label: t.p2tr, value: AddressType.P2TR }, // Taproot
  ];

  const [selectedNodeId, setSelectedNodeId] = useState<number | undefined>();

  const [newAddress, { error, data, isLoading, isSuccess, isError, isUninitialized }] = useNewAddressMutation();

  const handleClickNext = (addType: AddressType) => {
    if (selectedNodeId) {
      newAddress({
        nodeId: selectedNodeId,
        type: addType,
        // account: {account},
      });
    }
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate
      title={t.header.newAddress}
      show={true}
      onClose={() => navigate(-1)}
      icon={<TransactionIconModal />}
    >
      <div className={styles.nodeSelectionWrapper}>
        <div className={styles.nodeSelection}>
          <Select
            intercomTarget={"new-address-select-node"}
            label={t.yourNode}
            onChange={(newValue: unknown) => {
              const value = newValue as Option;
              if (value && value.value != 0) {
                setSelectedNodeId(value.value);
              }
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions?.find((option) => option.value === selectedNodeId)}
          />
        </div>
      </div>
      <div className={styles.addressTypeWrapper}>
        <div className={styles.addressTypes}>
          {addressTypeOptions.map((addType, index) => {
            return (
              <Button
                buttonPosition={ButtonPosition.fullWidth}
                intercomTarget={"new-address-" + addType.label}
                disabled={isLoading || selectedNodeId === undefined}
                buttonColor={ColorVariant.primary}
                key={index + addType.label}
                icon={isLoading && <Spinny />}
                onClick={() => {
                  if (selectedNodeId) {
                    handleClickNext(addType.value);
                    track("Select Address Type", { addressType: addType.label });
                  }
                }}
              >
                {addType.label}
              </Button>
            );
          })}
        </div>
      </div>
      <div className={styles.addressResultWrapper}>
        {isUninitialized && (
          <Note
            title={t.newAddress}
            noteType={isLoading ? NoteType.warning : NoteType.info}
            icon={<TransactionIconModal />}
          >
            {isLoading ? t.loading : t.selectAddressType}
          </Note>
        )}
        {data && (
          <Note title={t.newAddress} noteType={NoteType.success} icon={<TransactionIconModal />}>
            {data || t.selectAddressType}
          </Note>
        )}
        {data && isSuccess && (
          <Button
            intercomTarget={"new-address-copy-address"}
            buttonColor={ColorVariant.success}
            buttonSize={SizeVariant.normal}
            buttonPosition={ButtonPosition.fullWidth}
            icon={<CopyIcon />}
            onClick={() => {
              if (data) {
                toastRef?.current?.addToast("Copied to clipboard", toastCategory.success);
                navigator.clipboard.writeText(data);
              }
            }}
          >
            {t.Copy}
          </Button>
        )}
        {isError && <ErrorSummary title={t.error} errors={(error as { data: ServerErrorType })?.data?.errors} />}
      </div>
    </PopoutPageTemplate>
  );
}

export default NewAddressModal;

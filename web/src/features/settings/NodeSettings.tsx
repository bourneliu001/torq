import React, { useState } from "react";
import {
  ChevronDown20Regular as CollapsedIcon,
  Delete20Regular as DeleteIcon,
  Delete24Regular as DeleteIconHeader,
  LineHorizontal120Regular as ExpandedIcon,
  MoreCircle20Regular as MoreIcon,
  Pause16Regular as DisconnectedIcon,
  Pause20Regular as PauseIcon,
  Play16Regular as ConnectedIcon,
  Play20Regular as PlayIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import styles from "./NodeSettings.module.scss";
import Select, { SelectOption, SelectOptions } from "features/forms/Select";
import Spinny from "features/spinny/Spinny";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import File from "components/forms/file/File";
import Input from "components/formsWithValidation/input/InputWithValidation";
import {
  useAddNodeConfigurationMutation,
  useGetNodeConfigurationQuery,
  useUpdateNodeConfigurationMutation,
  useUpdateNodeConfigurationStatusMutation,
  useUpdateCustomSettingsMutation,
} from "apiSlice";
import { nodeConfiguration } from "apiTypes";
import classNames from "classnames";
import Collapse from "features/collapse/Collapse";
import Popover from "features/popover/Popover";
import Button, { ButtonPosition, ColorVariant } from "components/buttons/Button";
import Modal from "features/modal/Modal";
import Switch from "components/forms/switch/Switch";
import useTranslations from "services/i18n/useTranslations";
import Form from "components/forms/form/Form";
import Note, { NoteType } from "features/note/Note";
import ErrorSummary from "components/errors/ErrorSummary";
import { FormErrors, mergeServerError } from "components/errors/errors";
import { format } from "date-fns";
import { torqApi } from "apiSlice";
import { useAppDispatch } from "store/hooks";
import { userEvents } from "utils/userEvents";
import { components, OptionProps, SingleValueProps } from "react-select";
import PropTypes from "prop-types";

interface nodeProps {
  nodeId: number;
  collapsed?: boolean;
  addMode?: boolean;
  onAddSuccess?: () => void;
  onAddFailure?: () => void;
}

const nodeConfigurationTemplate = {
  createdOn: undefined,
  grpcAddress: "",
  macaroonFileName: "",
  name: "",
  tlsFileName: "",
  caCertificateFileName: "",
  certificateFileName: "",
  keyFileName: "",
  updatedOn: undefined,
  implementation: 0,
  nodeId: 0,
  status: 0,
  pingSystem: 0,
  customSettings: 0,
  nodeStartDate: undefined,
};

const importFailedPayments = "importFailedPayments";
const importFailedPaymentsValue = 1;
const importHtlcEvents = "importHtlcEvents";
const importHtlcEventsValue = 2;
// PeerEvents has been removed as an option
const importTransactions = "importTransactions";
const importTransactionsValue = 8;
const importPayments = "importPayments";
const importPaymentsValue = 16;
const importInvoices = "importInvoices";
const importInvoicesValue = 32;
const importForwards = "importForwards";
const importForwardsValue = 64;
const importForwardsHistory = "importForwardsHistory";
const importForwardsHistoryValue = 128;

const customSettingsDefault = {
  importFailedPayments: true,
  importHtlcEvents: true,
  importTransactions: true,
  importPayments: true,
  importInvoices: true,
  importForwards: true,
  importForwardsHistory: true,
};

interface importProps {
  value: number;
  label?: string;
}

type ColourOption = {
  value: string;
};

const colourOptions: ColourOption[] = [
  { value: "#85c4ff" },
  { value: "#354e66" },
  { value: "#6ca09e" },
  { value: "#4b473d" },
  { value: "#e18484" },
  { value: "#c875b6" },
  { value: "#9576c8" },
];

const colourSelectSingleValue = ({ ...props }: SingleValueProps<unknown>) => {
  const colourOption: ColourOption = props.data as ColourOption;
  return (
    <components.SingleValue {...props}>
      <div className={styles.colourSelectContainer}>
        <div className={styles.colourSelect} style={{ backgroundColor: colourOption.value }}></div>
      </div>
    </components.SingleValue>
  );
};

colourSelectSingleValue.propTypes = {
  data: PropTypes.object.isRequired,
};

const colourSelectOption = (props: OptionProps) => {
  const colourOption = props.data as ColourOption;
  return (
    <components.Option {...props}>
      <div className={styles.colourSelectContainer}>
        <div className={styles.colourSelect} style={{ backgroundColor: colourOption.value }}></div>
      </div>
    </components.Option>
  );
};

colourSelectOption.propTypes = {
  data: PropTypes.object.isRequired,
};

const NodeSettings = React.forwardRef(function NodeSettings(
  { nodeId, collapsed, addMode, onAddSuccess }: nodeProps,
  ref
) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const toastRef = React.useContext(ToastContext);
  const popoverRef = React.useRef();

  const { data: nodeConfigurationData } = useGetNodeConfigurationQuery(nodeId, {
    skip: !nodeId || nodeId == 0,
  });
  const [updateNodeConfiguration] = useUpdateNodeConfigurationMutation();
  const [addNodeConfiguration] = useAddNodeConfigurationMutation();
  const [setNodeConfigurationStatus] = useUpdateNodeConfigurationStatusMutation();
  const [setCustomSettings] = useUpdateCustomSettingsMutation();

  const [nodeConfigurationState, setNodeConfigurationState] = useState<nodeConfiguration>(nodeConfigurationTemplate);
  const [collapsedState, setCollapsedState] = useState(collapsed ?? false);
  const [customSettingsCollapsedState, setCustomSettingsCollapsedState] = useState(true);
  const [customSettingsSaveEnabledState, setCustomSettingsSaveEnabledState] = useState(false);
  const [showModalState, setShowModalState] = useState(false);
  const [deleteConfirmationTextInputState, setDeleteConfirmationTextInputState] = useState("");
  const [deleteEnabled, setDeleteEnabled] = useState(false);
  const [saveEnabledState, setSaveEnabledState] = useState(true);
  const [enableEnableButtonState, setEnableEnableButtonState] = useState(true);
  const [customSettingsState, setCustomSettingsState] = React.useState(customSettingsDefault);
  const [formErrorState, setFormErrorState] = React.useState({} as FormErrors);
  const [toggleErrorState, setToggleErrorState] = React.useState({} as FormErrors);
  const [tlsEnabledState, setTlsEnabledState] = useState(true);
  const [macaroonEnabledState, setMacaroonEnabledState] = useState(true);
  const [caCertificateEnabledState, setCaCertificateEnabledState] = useState(false);
  const [certificateEnabledState, setCertificateEnabledState] = useState(false);
  const [keyEnabledState, setKeyEnabledState] = useState(false);

  const dispatch = useAppDispatch();

  const customSettingsSidebarData = new Map<string, importProps>([
    [importTransactions, { value: importTransactionsValue, label: t.importTransactions }],
    [importPayments, { value: importPaymentsValue, label: t.importPayments }],
    [importInvoices, { value: importInvoicesValue, label: t.importInvoices }],
    [importForwards, { value: importForwardsValue, label: t.importForwards }],
    [importForwardsHistory, { value: importForwardsHistoryValue, label: t.importForwardsHistory }],
    [importHtlcEvents, { value: importHtlcEventsValue, label: t.importHtlcEvents }],
    [importFailedPayments, { value: importFailedPaymentsValue, label: undefined }],
  ]);

  React.useImperativeHandle(ref, () => ({
    clear() {
      clear();
    },
  }));

  const clear = () => {
    setCustomSettingsState(customSettingsDefault);
    setNodeConfigurationState({
      grpcAddress: "",
      nodeId: 0,
      status: 0,
      implementation: 0,
      pingSystem: 0,
      name: "",
      customSettings: 0,
      nodeStartDate: undefined,
    } as nodeConfiguration);
  };

  React.useEffect(() => {
    if (collapsed != undefined) {
      setCollapsedState(collapsed);
    }
  }, [collapsed]);

  const handleConfirmationModalClose = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
    setNodeConfigurationState({
      implementation: 0,
      nodeId: 0,
      status: 0,
      pingSystem: 0,
      customSettings: 0,
    });
    setTlsEnabledState(true);
    setMacaroonEnabledState(true);
    setCaCertificateEnabledState(false);
    setCertificateEnabledState(false);
    setKeyEnabledState(false);
  };

  const handleDeleteClick = () => {
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
    setShowModalState(true);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    submitNodeSettings();
  };

  const submitNodeSettings = async () => {
    setSaveEnabledState(false);
    const form = new FormData();
    form.append("implementation", "" + nodeConfigurationState.implementation);
    form.append("name", nodeConfigurationState.name ?? "");
    form.append("nodeId", "" + nodeConfigurationState.nodeId);
    form.append("status", "" + nodeConfigurationState.status);
    form.append("pingSystem", "" + nodeConfigurationState.pingSystem);
    form.append("customSettings", "" + nodeConfigurationState.customSettings);
    form.append("grpcAddress", nodeConfigurationState.grpcAddress ?? "");
    form.append("nodeCssColour", "" + nodeConfigurationState.nodeCssColour);
    if (nodeConfigurationState.tlsFile) {
      form.append("tlsFile", nodeConfigurationState.tlsFile, nodeConfigurationState.tlsFileName);
    }
    if (nodeConfigurationState.macaroonFile) {
      form.append("macaroonFile", nodeConfigurationState.macaroonFile, nodeConfigurationState.macaroonFileName);
    }
    if (nodeConfigurationState.caCertificateFile) {
      form.append(
        "caCertificateFile",
        nodeConfigurationState.caCertificateFile,
        nodeConfigurationState.caCertificateFileName
      );
    }
    if (nodeConfigurationState.certificateFile) {
      form.append(
        "certificateFile",
        nodeConfigurationState.certificateFile,
        nodeConfigurationState.certificateFileName
      );
    }
    if (nodeConfigurationState.keyFile) {
      form.append("keyFile", nodeConfigurationState.keyFile, nodeConfigurationState.keyFileName);
    }
    if (nodeConfigurationState.nodeStartDate) {
      form.append("nodeStartDate", "" + formatDate(nodeConfigurationState.nodeStartDate));
    }
    // we are adding new node
    if (!nodeConfigurationState.nodeId || nodeConfigurationState.nodeId == 0) {
      addNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setFormErrorState({} as FormErrors);
          setSaveEnabledState(true);
          setEnableEnableButtonState(true);
          toastRef?.current?.addToast("Local node added", toastCategory.success);
          if (onAddSuccess) {
            onAddSuccess();
          }
        })
        .catch((error) => {
          setSaveEnabledState(true);
          /* toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error); */
          const mergedErrors = mergeServerError(error.data, formErrorState);
          setFormErrorState(mergedErrors);
        });
      track("Add Local Node");
      return;
    } else {
      updateNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setFormErrorState({} as FormErrors);
          setSaveEnabledState(true);
          toastRef?.current?.addToast("Local node info saved", toastCategory.success);
        })
        .catch((error) => {
          setSaveEnabledState(true);
          /* toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error); */
          const mergedErrors = mergeServerError(error.data, formErrorState);
          setFormErrorState(mergedErrors);
        });
      track("Update Local Node", { nodeId: nodeConfigurationState.nodeId });
    }
    dispatch(torqApi.util.resetApiState());
  };

  const submitCustomSettings = async () => {
    setCustomSettingsSaveEnabledState(false);
    setCustomSettings({
      nodeId: nodeConfigurationState.nodeId,
      customSettings: nodeConfigurationState.customSettings,
      pingSystems: nodeConfigurationState.pingSystem,
    })
      .unwrap()
      .then((_) => {
        setToggleErrorState({} as FormErrors);
        setCustomSettingsSaveEnabledState(true);
        setEnableEnableButtonState(true);
        toastRef?.current?.addToast("Custom Settings Saved", toastCategory.success);
      })
      .catch((error) => {
        setCustomSettingsSaveEnabledState(true);
        /* toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error); */
        const mergedErrors = mergeServerError(error.data, formErrorState);
        setToggleErrorState(mergedErrors);
      });
    track("Save Custom Settings");
    dispatch(torqApi.util.resetApiState());
  };

  React.useEffect(() => {
    setNodeConfigurationState(nodeConfigurationData || nodeConfigurationTemplate);
    if (nodeConfigurationData == undefined) {
      setNodeConfigurationState({ ...nodeConfigurationState, customSettings: 254 });
      setCustomSettingsState(customSettingsDefault);
    } else {
      setCustomSettingsState({
        importFailedPayments:
          nodeConfigurationData.customSettings % (importFailedPaymentsValue * 2) >= importFailedPaymentsValue,
        importHtlcEvents: nodeConfigurationData.customSettings % (importHtlcEventsValue * 2) >= importHtlcEventsValue,
        importTransactions:
          nodeConfigurationData.customSettings % (importTransactionsValue * 2) >= importTransactionsValue,
        importPayments: nodeConfigurationData.customSettings % (importPaymentsValue * 2) >= importPaymentsValue,
        importInvoices: nodeConfigurationData.customSettings % (importInvoicesValue * 2) >= importInvoicesValue,
        importForwards: nodeConfigurationData.customSettings % (importForwardsValue * 2) >= importForwardsValue,
        importForwardsHistory:
          nodeConfigurationData.customSettings % (importForwardsHistoryValue * 2) >= importForwardsHistoryValue,
      });
      setCustomSettingsCollapsedState(false);
    }
    setTlsEnabledState(true);
    setMacaroonEnabledState(true);
    setCaCertificateEnabledState(false);
    setCertificateEnabledState(false);
    setKeyEnabledState(false);
    if (nodeConfigurationData != undefined && nodeConfigurationData.status == 0) {
      setSaveEnabledState(true);
    }
    if (nodeConfigurationData != undefined && nodeConfigurationData.implementation == 1) {
      setTlsEnabledState(false);
      setMacaroonEnabledState(false);
      setCaCertificateEnabledState(true);
      setCertificateEnabledState(true);
      setKeyEnabledState(true);
    }
  }, [nodeConfigurationData]);

  const getCustomSettingsState = (key: string) => {
    const data = customSettingsSidebarData.get(key);
    if (data !== undefined && nodeConfigurationState != undefined) {
      return nodeConfigurationState.customSettings % (data.value * 2) >= data.value;
    }
    return true;
  };

  const toggleCustomSettingsState = (key: string) => {
    const data = customSettingsSidebarData.get(key);
    if (data !== undefined) {
      if (getCustomSettingsState(key)) {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings - data.value,
        });
      } else {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings + data.value,
        });
      }
    }
  };

  const handleTLSFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      tlsFile: file,
      tlsFileName: file ? file.name : undefined,
    });
  };
  const handleMacaroonFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      macaroonFile: file,
      macaroonFileName: file ? file.name : undefined,
    });
  };
  const handleCaCertificateFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      caCertificateFile: file,
      caCertificateFileName: file ? file.name : undefined,
    });
  };
  const handleCertificateFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      certificateFile: file,
      certificateFileName: file ? file.name : undefined,
    });
  };
  const handleKeyFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      keyFile: file,
      keyFileName: file ? file.name : undefined,
    });
  };

  const handleAddressChange = (value: string) => {
    setNodeConfigurationState({ ...nodeConfigurationState, grpcAddress: value });
  };

  const handleNodeNameChange = (value: string) => {
    setNodeConfigurationState({ ...nodeConfigurationState, name: value });
  };

  const handleCollapseClick = () => {
    setCollapsedState(!collapsedState);
  };

  const handleCustomSettingsCollapseClick = () => {
    setCustomSettingsCollapsedState(!customSettingsCollapsedState);
  };

  const handleModalDeleteClick = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
    track("Delete Local Node", { nodeId: nodeConfigurationState.nodeId });
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: 3 });
    dispatch(torqApi.util.resetApiState());
  };

  const handleDeleteConfirmationTextInputChange = (value: string) => {
    setDeleteConfirmationTextInputState(value as string);
    setDeleteEnabled(value.toLowerCase() === "delete");
  };

  const handleStatusClick = () => {
    setEnableEnableButtonState(false);
    setSaveEnabledState(false);
    let statusId = 0;
    if (nodeConfigurationState.status == 0) {
      statusId = 1;
      track("Enable Local Node", { nodeId: nodeConfigurationState.nodeId });
    } else {
      track("Disable Local Node", { nodeId: nodeConfigurationState.nodeId });
    }
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: statusId })
      .unwrap()
      .finally(() => {
        setEnableEnableButtonState(true);
        setSaveEnabledState(true);
      });
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
    dispatch(torqApi.util.resetApiState());
  };

  const handleAmbossPingClick = () => {
    setCustomSettingsSaveEnabledState(true);
    const pingSystem = 1;
    if (nodeConfigurationState.pingSystem % (pingSystem * 2) >= pingSystem) {
      setNodeConfigurationState({
        ...nodeConfigurationState,
        pingSystem: nodeConfigurationState.pingSystem - pingSystem,
      });
    } else {
      setNodeConfigurationState({
        ...nodeConfigurationState,
        pingSystem: nodeConfigurationState.pingSystem + pingSystem,
      });
    }
  };

  const handleVectorPingClick = () => {
    setCustomSettingsSaveEnabledState(true);
    const pingSystem = 2;
    if (nodeConfigurationState.pingSystem % (pingSystem * 2) >= pingSystem) {
      setNodeConfigurationState({
        ...nodeConfigurationState,
        pingSystem: nodeConfigurationState.pingSystem - pingSystem,
      });
    } else {
      setNodeConfigurationState({
        ...nodeConfigurationState,
        pingSystem: nodeConfigurationState.pingSystem + pingSystem,
      });
    }
  };

  const handleNodeStartDateChange = (value: string) => {
    // use Date.parse to get utc date
    const date = new Date(Date.parse(value));
    if (!Number.isNaN(date.valueOf()) && date.valueOf() !== 0) {
      setNodeConfigurationState({ ...nodeConfigurationState, nodeStartDate: date });
    } else {
      setNodeConfigurationState({ ...nodeConfigurationState, nodeStartDate: undefined });
    }
  };

  const toggleCustomSettingsStateNow = (key: string) => {
    setCustomSettingsSaveEnabledState(true);
    const data = customSettingsSidebarData.get(key);
    if (data !== undefined) {
      if (getCustomSettingsState(key)) {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings - data.value,
        });
      } else {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings + data.value,
        });
      }
    }
  };

  const formatDate = (date: Date | undefined) => {
    if (date != undefined && !Number.isNaN(date.valueOf()) && date.valueOf() !== 0) {
      return format(new Date(date.valueOf()), "yyyy-MM-dd");
    }
    return "";
  };

  const implementationOptions: Array<SelectOption> = [
    { value: "0", label: "LND" },
    { value: "1", label: "CLN" },
  ];

  const menuButton = <MoreIcon className={styles.moreIcon} />;

  const handleNodeCssColourChange = (value: string) => {
    setNodeConfigurationState({ ...nodeConfigurationState, nodeCssColour: value });
  };

  return (
    <>
      {!addMode && (
        <div
          className={classNames(styles.header, { [styles.expanded]: !collapsedState })}
          onClick={handleCollapseClick}
        >
          <div
            className={classNames(styles.connectionIcon, {
              [styles.connected]: true,
              [styles.disabled]: nodeConfigurationState.status == 0,
            })}
          >
            {nodeConfigurationState.status == 0 && <DisconnectedIcon />}
            {nodeConfigurationState.status == 1 && <ConnectedIcon />}
          </div>
          <div className={styles.title}>{nodeConfigurationState?.name}</div>
          <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedState })}>
            {collapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
          </div>
        </div>
      )}
      <Collapse collapsed={collapsedState} animate={!addMode}>
        <div className={classNames(styles.collapseContentWrappper, { [styles.addMode]: addMode })}>
          {!addMode && (
            <>
              <div className={styles.borderSection}>
                <div className={styles.detailHeader}>
                  <h4 className={styles.detailsTitle}>Node Details</h4>
                  <Popover button={menuButton} className={classNames("right", styles.moreButton)} ref={popoverRef}>
                    <div className={styles.nodeMenu}>
                      <Button
                        intercomTarget={"node-configuration-disable-button"}
                        buttonPosition={ButtonPosition.fullWidth}
                        buttonColor={ColorVariant.warning}
                        icon={nodeConfigurationState.status == 0 ? <PlayIcon /> : <PauseIcon />}
                        onClick={handleStatusClick}
                        disabled={!enableEnableButtonState}
                      >
                        {nodeConfigurationState.status == 0 ? "Enable node" : "Disable node"}
                      </Button>
                      <Button
                        intercomTarget={"node-configuration-delete-button"}
                        buttonColor={ColorVariant.error}
                        icon={<DeleteIcon />}
                        onClick={handleDeleteClick}
                        buttonPosition={ButtonPosition.fullWidth}
                      >
                        {"Delete node"}
                      </Button>
                    </div>
                  </Popover>
                </div>
              </div>
            </>
          )}
          <div className={""}>
            <Form onSubmit={handleSubmit} intercomTarget={"node-configuration-form"}>
              <Select
                intercomTarget={"node-setting-select-implementation"}
                label={t.implementation}
                onChange={(newValue: unknown) => {
                  const selectOptions = newValue as SelectOptions;
                  if (selectOptions?.value == "0") {
                    setTlsEnabledState(true);
                    setMacaroonEnabledState(true);
                    setCaCertificateEnabledState(false);
                    setCertificateEnabledState(false);
                    setKeyEnabledState(false);
                  } else if (selectOptions?.value == "1") {
                    setTlsEnabledState(false);
                    setMacaroonEnabledState(false);
                    setCaCertificateEnabledState(true);
                    setCertificateEnabledState(true);
                    setKeyEnabledState(true);
                  }
                  setNodeConfigurationState({
                    ...nodeConfigurationState,
                    implementation: selectOptions?.value as number,
                  });
                }}
                options={implementationOptions}
                value={implementationOptions.find((io) => io.value == "" + nodeConfigurationState.implementation)}
              />
              <span id="name">
                <Input
                  intercomTarget={"node-setting-input-name"}
                  label={t.nodeName}
                  value={nodeConfigurationState.name}
                  type={"text"}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleNodeNameChange(e.target.value)}
                  placeholder="Node 1"
                />
              </span>
              <span id="address">
                <Input
                  intercomTarget={"node-setting-input-address"}
                  label={t.grpcAddress}
                  type={"text"}
                  value={nodeConfigurationState.grpcAddress}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleAddressChange(e.target.value)}
                  placeholder="100.100.100.100:10009"
                  name="grpcAddress"
                  errors={formErrorState}
                />
              </span>
              {tlsEnabledState && (
                <span id="tls">
                  <File
                    intercomTarget={"node-setting-input-tls"}
                    label={t.tlsCertificate}
                    onFileChange={handleTLSFileChange}
                    fileName={nodeConfigurationState?.tlsFileName}
                  />
                </span>
              )}
              {macaroonEnabledState && (
                <span id="macaroon">
                  <File
                    intercomTarget={"node-setting-input-macaroon"}
                    label={t.macaroon}
                    onFileChange={handleMacaroonFileChange}
                    fileName={nodeConfigurationState?.macaroonFileName}
                  />
                </span>
              )}
              {caCertificateEnabledState && (
                <span id="caCertificate">
                  <File
                    intercomTarget={"node-setting-input-ca"}
                    label={t.caCertificate}
                    onFileChange={handleCaCertificateFileChange}
                    fileName={nodeConfigurationState?.caCertificateFileName}
                  />
                </span>
              )}
              {certificateEnabledState && (
                <span id="certificate">
                  <File
                    intercomTarget={"node-setting-input-client"}
                    label={t.certificate}
                    onFileChange={handleCertificateFileChange}
                    fileName={nodeConfigurationState?.certificateFileName}
                  />
                </span>
              )}
              {keyEnabledState && (
                <span id="key">
                  <File
                    intercomTarget={"node-setting-input-clientKey"}
                    label={t.key}
                    onFileChange={handleKeyFileChange}
                    fileName={nodeConfigurationState?.keyFileName}
                  />
                </span>
              )}
              <span id="nodeStartDate">
                <Input
                  intercomTarget={"node-setting-input-nodeStartDate"}
                  label={t.nodeStartDate}
                  value={formatDate(nodeConfigurationState.nodeStartDate)}
                  max={formatDate(new Date())}
                  type={"date"}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleNodeStartDateChange(e.target.value)}
                />
              </span>
              <Select
                intercomTarget="settings-node-colour"
                selectComponents={{ Option: colourSelectOption, SingleValue: colourSelectSingleValue }}
                label={t.nodeColour}
                onChange={(newValue: unknown) => {
                  const value = newValue as ColourOption;
                  handleNodeCssColourChange(value.value ?? undefined);
                }}
                options={colourOptions}
                value={colourOptions.find((option) => option.value === nodeConfigurationState.nodeCssColour)}
              />
              {addMode && (
                <div className={styles.customImportSettings}>
                  <div
                    className={classNames(styles.header, { [styles.expanded]: !customSettingsCollapsedState })}
                    onClick={handleCustomSettingsCollapseClick}
                  >
                    <div className={styles.title}>{t.advancedSettings}</div>
                    <div
                      className={classNames(styles.collapseIcon, { [styles.collapsed]: customSettingsCollapsedState })}
                    >
                      {customSettingsCollapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
                    </div>
                  </div>
                  <Collapse collapsed={customSettingsCollapsedState} animate={true}>
                    <div className={styles.customImportSettingsBody}>
                      {Object.keys(customSettingsState).map((key) => {
                        const k = key as keyof typeof customSettingsState;
                        const data = customSettingsSidebarData.get(key);
                        if (data !== undefined && data.label !== undefined) {
                          return (
                            <div className={styles.import} key={key}>
                              <Switch
                                intercomTarget={"node-setting-input-" + key + "-switch"}
                                label={data.label}
                                checked={nodeConfigurationState.customSettings % (data.value * 2) >= data.value}
                                onChange={() => {
                                  toggleCustomSettingsState(k);
                                }}
                              />
                            </div>
                          );
                        }
                      })}
                      <Switch
                        label={t.importFailedPayments}
                        intercomTarget={"node-setting-input-importFailedPayments-switch"}
                        checked={
                          nodeConfigurationState.customSettings % (importFailedPaymentsValue * 2) >=
                          importFailedPaymentsValue
                        }
                        onChange={() => {
                          toggleCustomSettingsState(importFailedPayments);
                        }}
                      />
                      <div className={styles.importFailedPayments}>
                        <Note title={"Failed Payments"} noteType={NoteType.warning}>
                          {t.info.importFailedPayments}
                        </Note>
                      </div>
                    </div>
                  </Collapse>
                </div>
              )}
              <ErrorSummary errors={formErrorState} />
              <Button
                id={"save-node"}
                intercomTarget={"node-setting-save-button"}
                buttonColor={ColorVariant.success}
                icon={saveEnabledState || nodeConfigurationState.status == 1 ? <SaveIcon /> : <Spinny />}
                onClick={submitNodeSettings}
                buttonPosition={ButtonPosition.fullWidth}
                disabled={!saveEnabledState || nodeConfigurationState.status == 1}
              >
                {addMode
                  ? "Add Node"
                  : nodeConfigurationState.status == 1
                  ? "Disable node to update"
                  : saveEnabledState
                  ? "Save node details"
                  : "Saving..."}
              </Button>
              {!addMode && (
                <div className={styles.toggleSettings} data-intercom-target={"node-setting-advanced-toggle-button"}>
                  <div
                    className={classNames(styles.header, { [styles.expanded]: !customSettingsCollapsedState })}
                    onClick={handleCustomSettingsCollapseClick}
                  >
                    <div className={styles.title}>{t.advancedSettings}</div>
                    <div
                      className={classNames(styles.collapseIcon, { [styles.collapsed]: customSettingsCollapsedState })}
                    >
                      {customSettingsCollapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
                    </div>
                  </div>
                  <Collapse
                    collapsed={customSettingsCollapsedState}
                    animate={true}
                    intercomTarget={"node-setting-advanced-options-addmode"}
                  >
                    <div className={styles.customImportSettingsBody}>
                      {Object.keys(customSettingsState).map((key) => {
                        const k = key as keyof typeof customSettingsState;
                        const data = customSettingsSidebarData.get(key);
                        if (data !== undefined && data.label !== undefined) {
                          return (
                            <div className={styles.import} key={key}>
                              <Switch
                                intercomTarget={"node-setting-input-" + key + "-switch-addmode"}
                                label={data.label || ""}
                                checked={customSettingsState[k]}
                                onChange={() => {
                                  setCustomSettingsState({
                                    ...customSettingsState,
                                    [key]: !customSettingsState[k],
                                  });
                                  toggleCustomSettingsStateNow(key);
                                }}
                              />
                            </div>
                          );
                        }
                      })}
                      <Switch
                        intercomTarget={"node-setting-input-importFailedPayments-switch-addmode"}
                        label={t.importFailedPayments}
                        checked={
                          nodeConfigurationState.customSettings % (importFailedPaymentsValue * 2) >=
                          importFailedPaymentsValue
                        }
                        onChange={() => {
                          toggleCustomSettingsStateNow(importFailedPayments);
                        }}
                      />
                      <Switch
                        intercomTarget={"node-setting-input-vectorPing-switch-addmode"}
                        label="Vector Ping"
                        checked={nodeConfigurationState.pingSystem % 4 >= 2}
                        onChange={handleVectorPingClick}
                      />
                      <Switch
                        intercomTarget={"node-setting-input-ambossPing-switch-addmode"}
                        label="Amboss Ping"
                        checked={nodeConfigurationState.pingSystem % 2 >= 1}
                        onChange={handleAmbossPingClick}
                      />
                      <Note
                        title={t.note}
                        noteType={NoteType.info}
                        intercomTarget={"node-setting-advanced-options-note-addmode"}
                      >
                        <p>{t.info.importFailedPayments}</p>
                        <p>{t.pingNote}</p>
                        <p>{t.header.pingSystem}</p>
                        <p>{t.header.vectorPingSystem}</p>
                        <p>{t.header.ambossPingSystem}</p>
                      </Note>
                    </div>
                    <ErrorSummary errors={toggleErrorState} />
                    <Button
                      intercomTarget={"node-setting-advanced-options-save-button-addmode"}
                      id={"customSettings-save-node"}
                      buttonColor={ColorVariant.success}
                      icon={<SaveIcon />}
                      onClick={submitCustomSettings}
                      buttonPosition={ButtonPosition.fullWidth}
                      disabled={!customSettingsSaveEnabledState}
                    >
                      Save toggles
                    </Button>
                  </Collapse>
                </div>
              )}
            </Form>
          </div>
        </div>
      </Collapse>
      <Modal
        title={"Are you sure?"}
        icon={<DeleteIconHeader />}
        onClose={handleConfirmationModalClose}
        show={showModalState}
      >
        <div className={styles.deleteConfirm}>
          <p>
            Deleting the node will prevent you from viewing it&apos;s data in Torq. Alternatively set node to disabled
            to simply stop the data subscription but keep data collected so far.
          </p>
          <p>
            This operation cannot be undone, type &quot;<span className={styles.red}>delete</span>&quot; to confirm.
          </p>
          <Input
            placeholder={t.header.typeDeleteHere}
            value={deleteConfirmationTextInputState}
            type={"text"}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              handleDeleteConfirmationTextInputChange(e.target.value)
            }
          />
          <div className={styles.deleteConfirmButtons}>
            <Button
              intercomTarget={"node-setting-delete-confirm-button"}
              buttonColor={ColorVariant.error}
              buttonPosition={ButtonPosition.fullWidth}
              icon={<DeleteIcon />}
              onClick={handleModalDeleteClick}
              disabled={!deleteEnabled}
            >
              {t.delete}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
});
export default NodeSettings;

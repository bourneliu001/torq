import { useAppDispatch, useAppSelector } from "store/hooks";
import { Globe20Regular as GlobeIcon } from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import Popover from "features/popover/Popover";
import Button, { ButtonPosition, ColorVariant, SizeVariant } from "components/buttons/Button";
import { Network, selectActiveNetwork, setActiveNetwork } from "features/network/networkSlice";
import useTranslations from "services/i18n/useTranslations";
import { userEvents } from "utils/userEvents";

function NetworkSelector() {
  const { t } = useTranslations();
  const dispatch = useAppDispatch();
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const { track } = userEvents();

  return (
    <div className={styles.quickToggles}>
      <Popover
        button={
          <Button
            intercomTarget={"network-selector"}
            buttonColor={ColorVariant.ghost}
            buttonSize={SizeVariant.small}
            icon={<GlobeIcon />}
            hideMobileText={true}
          />
        }
        className={"right"}
      >
        <div className={styles.quickToggleContent}>
          <Button
            buttonPosition={ButtonPosition.fullWidth}
            intercomTarget={"network-select-mainnet"}
            buttonColor={activeNetwork === Network.MainNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.MainNet));
              track("Select Network", { networkSelected: "MainNet" });
            }}
          >
            {t.MainNet}
          </Button>
          <Button
            buttonPosition={ButtonPosition.fullWidth}
            intercomTarget={"network-select-testnet"}
            buttonColor={activeNetwork === Network.TestNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.TestNet));
              track("Select Network", { networkSelected: "TestNet" });
            }}
          >
            {t.TestNet}
          </Button>
          <Button
            buttonPosition={ButtonPosition.fullWidth}
            intercomTarget={"network-select-regtest"}
            buttonColor={activeNetwork === Network.RegTest ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.RegTest));
              track("Select Network", { networkSelected: "RegTest" });
            }}
          >
            {t.RegTest}
          </Button>
          <Button
            buttonPosition={ButtonPosition.fullWidth}
            intercomTarget={"network-select-signet"}
            buttonColor={activeNetwork === Network.SigNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.SigNet));
              track("Select Network", { networkSelected: "SigNet" });
            }}
          >
            {t.SigNet}
          </Button>
          <Button
            buttonPosition={ButtonPosition.fullWidth}
            intercomTarget={"network-select-simnet"}
            buttonColor={activeNetwork === Network.SimNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.SimNet));
              track("Select Network", { networkSelected: "SimNet" });
            }}
          >
            {t.SimNet}
          </Button>
        </div>
      </Popover>
    </div>
  );
}

export default NetworkSelector;

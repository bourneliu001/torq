import React from "react";
import classNames from "classnames";
import {
  WarningRegular as WarningIcon,
  ErrorCircleRegular as ErrorIcon,
  QuestionCircle16Regular as HelpIcon,
} from "@fluentui/react-icons";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import styles from "./slider.module.scss";
import labelStyles from "components/forms/label/label.module.scss";
import { BasicInputType } from "components/forms/formTypes";

export type SliderProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> &
  BasicInputType;

export type FormattedInputProps = {
  label?: string;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
  helpText?: string;
  intercomTarget: string;
};

function Slider({
  label,
  sizeVariant,
  colorVariant,
  leftIcon,
  errorText,
  warningText,
  helpText,
  intercomTarget,
  ...inputProps
}: SliderProps) {
  const inputId = React.useId();
  let inputColorClass = GetColorClass(colorVariant);
  if (warningText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.warning);
  }
  if (errorText != undefined) {
    inputColorClass = GetColorClass(InputColorVaraint.error);
  }
  if (inputProps.disabled === true) {
    inputColorClass = GetColorClass(InputColorVaraint.disabled);
  }

  return (
    <div
      className={classNames(styles.sliderWrapper, GetSizeClass(sizeVariant), inputColorClass)}
      data-intercom-target={intercomTarget}
    >
      {label && (
        <div className={labelStyles.labelWrapper}>
          <label htmlFor={inputProps.id || inputId} className={styles.label} title={"Something"}>
            {label}
          </label>
          {/* Create a div with a circled question mark icon with a data label named data-title */}
          {helpText && (
            <div className={labelStyles.tooltip}>
              <HelpIcon />
              <div className={labelStyles.tooltipTextWrapper}>
                <div className={labelStyles.tooltipTextContainer}>
                  <div className={labelStyles.tooltipHeader}>{label}</div>
                  <div className={labelStyles.tooltipText}>{helpText}</div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
      <div className={classNames(styles.inputFieldContainer, { [styles.hasLeftIcon]: !!leftIcon })}>
        {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
        <input
          type={"range"}
          {...(inputProps as SliderProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputProps.id || inputId}
        />
      </div>
      {errorText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackError)}>
          <div className={styles.feedbackIcon}>
            <ErrorIcon />
          </div>
          <div className={styles.feedbackText}>{errorText}</div>
        </div>
      )}
      {warningText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackWarning)}>
          <div className={styles.feedbackIcon}>
            <WarningIcon />
          </div>
          <div className={styles.feedbackText}>{warningText}</div>
        </div>
      )}
    </div>
  );
}
export default Slider;

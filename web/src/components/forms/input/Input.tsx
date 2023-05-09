import React from "react";
import classNames from "classnames";
import NumberFormat, { NumberFormatProps } from "react-number-format";
import {
  WarningRegular as WarningIcon,
  ErrorCircleRegular as ErrorIcon,
  InfoRegular as InfoIcon,
  QuestionCircle16Regular as HelpIcon,
} from "@fluentui/react-icons";
import { GetColorClass, GetSizeClass, InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import styles from "./textInput.module.scss";
import labelStyles from "components/forms/label/label.module.scss";
import { BasicInputType } from "components/forms/formTypes";
import { FormErrors, replaceMessageMergeTags } from "components/errors/errors";
import useTranslations from "services/i18n/useTranslations";

export type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> &
  BasicInputType;

export type FormattedInputProps = {
  label?: string;
  formatted?: boolean;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
  infoText?: string;
  helpText?: string;
  intercomTarget: string;
  button?: React.ReactNode;
  errors?: FormErrors;
} & NumberFormatProps;

function Input({
  label,
  formatted,
  sizeVariant,
  colorVariant,
  leftIcon,
  errorText,
  warningText,
  infoText,
  helpText,
  intercomTarget,
  button,
  errors,
  ...inputProps
}: InputProps | FormattedInputProps) {
  const { t } = useTranslations();
  const inputId = React.useId();

  if (errors && errors.fields && inputProps.name && errors.fields[inputProps.name]) {
    const codeOrDescription = errors.fields[inputProps.name][0];
    const translatedError = t.errors[codeOrDescription.code];
    const mergedError = replaceMessageMergeTags(translatedError, codeOrDescription.attributes);
    errorText = mergedError ?? codeOrDescription.description;
  }

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

  function renderInput() {
    if (formatted) {
      return (
        <NumberFormat
          {...(inputProps as FormattedInputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputProps.id || inputId}
        />
      );
    } else {
      return (
        <input
          {...(inputProps as InputProps)}
          className={classNames(styles.input, inputProps.className)}
          id={inputProps.id || inputId}
        />
      );
    }
  }

  return (
    <div
      className={classNames(styles.inputWrapper, GetSizeClass(sizeVariant), inputColorClass)}
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
      <div className={classNames({ [styles.inputButtonWrapper]: button })}>
        <div className={classNames(styles.inputFieldContainer, { [styles.hasLeftIcon]: !!leftIcon })}>
          {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
          {renderInput()}
        </div>
        {button && <div className={styles.button}>{button}</div>}
      </div>
      {infoText && (
        <div className={classNames(styles.feedbackWrapper, styles.feedbackInfo)}>
          <div className={styles.feedbackIcon}>
            <InfoIcon />
          </div>
          <div className={styles.feedbackText}>{infoText}</div>
        </div>
      )}
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
export default Input;

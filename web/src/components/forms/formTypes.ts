import { InputColorVaraint, InputSizeVariant } from "./input/variants";
import { ReactNode } from "react";
import { FormErrors } from "components/errors/errors";

export type BasicInputType = {
  label?: string;
  formatted?: never;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
  infoText?: string;
  helpText?: string;
  intercomTarget?: string;
  button?: ReactNode;
  errors?: FormErrors;
};

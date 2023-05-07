import { InputColorVaraint, InputSizeVariant } from "./input/variants";
import { ReactNode } from "react";

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
};

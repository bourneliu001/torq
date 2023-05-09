import { ServerErrorResultType, ServerErrorType } from "../components/errors/errors";

export function IsStringOption(result: unknown): result is { value: string; label: string } {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as { value: unknown; label: unknown }).value === "string"
  );
}

export function IsNumericOption(result: unknown): result is { value: number; label: string } {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as { value: unknown; label: string }).value === "number"
  );
}

// Create generic isOption function
export function IsOption<T>(result: unknown): result is { value: T; label: string } {
  return (
    result !== null &&
    typeof result === "object" &&
    "value" in result &&
    "label" in result &&
    typeof (result as { value: T; label: string }).value === typeof result
  );
}

export function IsServerErrorResult(result: unknown): result is ServerErrorResultType {
  return (
    result !== null &&
    typeof result === "object" &&
    "data" in result &&
    typeof (result as { data: unknown }).data === "object" &&
    "errors" in (result as { data: ServerErrorType }).data &&
    typeof (result as { data: ServerErrorType }).data.errors === "object"
  );
}

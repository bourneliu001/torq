import RawSelect, { SelectOptionType } from "components/forms/select/Select";
import { GroupBase, Props } from "react-select";
import { SelectComponents } from "react-select/dist/declarations/src/components";

type selectProps = {
  label: string;
  intercomTarget: string;
  selectComponents?: Partial<SelectComponents<unknown, boolean, GroupBase<unknown>>> | undefined;
} & Props;

export type SelectOptions = {
  label?: string;
  value: number | string;
  type?: string;
};

function Select(props: selectProps) {
  return (
    <RawSelect
      selectComponents={props.selectComponents}
      intercomTarget={props.intercomTarget}
      label={props.label}
      options={props.options}
      value={props.value}
      onChange={props.onChange}
    />
  );
}

export default Select;
export type SelectOption = SelectOptionType;

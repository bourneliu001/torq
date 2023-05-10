import { GroupByOptions } from "features/viewManagement/types";

export type ColumnMetaData<T> = {
  heading: string;
  key: keyof T;
  key2?: keyof T;
  suffix?: string;
  type?: string;
  width?: number;
  locked?: boolean;
  valueType: string;
  total?: number;
  max?: number;
  percent?: boolean;
  selectOptions?: Array<{ label: string; value: string }>;
};

export type TableProps<T> = {
  activeColumns: Array<ColumnMetaData<T>>;
  data: Array<T>;
  isLoading: boolean;
  showTotals?: boolean;
  totalRow?: T;
  maxRow?: T;
  cellRenderer: CellRendererFunction<T>;
  rowFirstCellRenderer?: CellRendererFunction<T>;
  selectable?: boolean;
  selectedRowIds?: Array<number>;
  intercomTarget?: string;
  groupedBy?: GroupByOptions;
};

export type CellRendererFunction<T> = (
  row: T,
  rowIndex: number,
  columnMeta: ColumnMetaData<T>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxValues?: T,
  groupedBy?: GroupByOptions
) => JSX.Element;

import { ColumnMetaData } from "features/table/types";
import NumericCell from "components/table/cells/numeric/NumericCell";
import BarCell from "components/table/cells/bar/BarCell";
import TextCell from "components/table/cells/text/TextCell";
import DurationCell from "components/table/cells/duration/DurationCell";
import BooleanCell from "components/table/cells/boolean/BooleanCell";
import DateCell from "components/table/cells/date/DateCell";
import EnumCell from "components/table/cells/enum/EnumCell";
import NumericDoubleCell from "components/table/cells/numeric/NumericDoubleCell";
import LongTextCell from "components/table/cells/longText/LongTextCell";
import LinkCell from "components/table/cells/link/LinkCell";

export default function DefaultCellRenderer<T>(
  row: T,
  rowIndex: number,
  column: ColumnMetaData<T>,
  columnIndex: number,
  totalsRow?: boolean,
  maxRow?: T
): JSX.Element {
  const dataKey = column.key as keyof T;
  const dataKey2 = column.key2 as keyof T;
  const suffix = column.suffix as string;
  // const heading = column.heading;
  const percent = column.percent;

  switch (column.valueType) {
    case "string":
      switch (column.type) {
        case "LongTextCell":
          return (
            <LongTextCell
              text={row[dataKey] as string}
              key={dataKey.toString() + rowIndex}
              copyText={row[dataKey] as string}
              totalCell={totalsRow}
            />
          );
        case "TextCell":
          return <TextCell text={row[dataKey] as string} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />;
        case "DurationCell":
          return (
            <DurationCell seconds={row[dataKey] as number} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />
          );
        case "EnumCell":
          return (
            <EnumCell
              value={row[dataKey] as string}
              key={dataKey.toString() + rowIndex + columnIndex}
              totalCell={totalsRow}
            />
          );
      }
      break;
    case "boolean":
      switch (column.type) {
        case "BooleanCell":
          return (
            <BooleanCell
              falseTitle={"Failure"}
              trueTitle={"Success"}
              value={row[dataKey] as boolean}
              key={dataKey.toString() + rowIndex + columnIndex}
              totalCell={totalsRow}
            />
          );
      }
      break;
    case "date":
      return <DateCell value={row[dataKey] as Date} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />;
    case "duration":
      return (
        <DurationCell seconds={row[dataKey] as number} key={dataKey.toString() + rowIndex} totalCell={totalsRow} />
      );
    case "link":
      return (
        <LinkCell
          text={row[dataKey] as string}
          link={row[dataKey] as string}
          key={dataKey.toString() + rowIndex}
          totalCell={totalsRow}
        />
      );
    case "number":
      switch (column.type) {
        case "NumericCell":
          return <NumericCell current={row[dataKey] as number} key={dataKey.toString() + rowIndex + columnIndex} />;
        case "BarCell":
          return (
            <BarCell
              current={row[dataKey] as number}
              max={maxRow ? (maxRow[dataKey] as number) : 0}
              showPercent={percent}
              key={dataKey.toString() + rowIndex + columnIndex}
              suffix={suffix}
            />
          );
        case "NumericDoubleCell":
          return (
            <NumericDoubleCell
              topValue={row[dataKey] as number}
              bottomValue={row[dataKey2] as number}
              suffix={suffix as string}
              className={dataKey.toString()}
              key={dataKey.toString() + rowIndex + columnIndex}
              totalCell={totalsRow}
            />
          );
      }
  }
  return (
    <TextCell text={row[dataKey] as string} key={dataKey.toString() + rowIndex} copyText={row[dataKey] as string} />
  );
}

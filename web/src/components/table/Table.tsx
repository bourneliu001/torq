import React, {Component} from 'react';
import './table.scss'
import tableRow from "./TableRow";
import HeaderCell from "./cells/HeaderCell";
import NameCell from "./cells/NameCell";
import NumericCell from "./cells/NumericCell";
import BarCell from "./cells/BarCell";

export interface ColumnMetaData {
  heading: string;
  type: string;
  width?: string;
  align?: string;
}

interface ColumnMeta {
  heading: string,
  key: string,
  type?: string,
  width?: number,
  locked?: boolean,
}

const columns: ColumnMeta[] = [
  {heading: 'Name', type: 'NameCell', key: 'group_name', locked: true},
  {heading: 'Capacity', type:'NumericCell', key: 'capacity'},
  {heading: 'Turnover', type:'NumericCell', key: 'turnover'},
  {heading: 'Amount out', type:'BarCell', key: 'amount_out'},
  {heading: 'Amount inbound', type:'BarCell', key: 'amount_in'},
  {heading: 'Amount total', type:'BarCell', key: 'amount_total'},
  {heading: 'Revenue out', type:'NumericCell', key: 'revenue_out'},
  {heading: 'Revenue inbound', type:'NumericCell', key: 'revenue_in'},
  {heading: 'Revenue total', type:'NumericCell', key: 'revenue_total'},
  {heading: 'Successful Forwards out', type:'NumericCell', key: 'count_out'},
  {heading: 'Successful Forwards inbound', type:'NumericCell', key: 'count_in'},
  {heading: 'Successful Forwards total', type:'NumericCell', key: 'count_total'},
];

interface RowType {
  group_name: string,
  amount_out: number,
  amount_in: number,
  amount_total: number,
  revenue_out: number,
  revenue_in: number,
  revenue_total: number,
  count_out: number,
  count_in: number,
  count_total: number,
  capacity: number,
  turnover: number,
}

let totalRow: RowType = {
    group_name: "Total",
    amount_out: 1200000,
    amount_in: 1200000,
    amount_total: 1200000,
    revenue_out: 1200000,
    revenue_in: 1200000,
    revenue_total: 1200000,
    count_out: 1200000,
    count_in: 1200000,
    count_total: 1200000,
    capacity: 1200000,
    turnover: 1.42,
}
let pastTotalRow: RowType = {
    group_name: "Total",
    amount_out: 1200000,
    amount_in: 1200000,
    amount_total: 1200000,
    revenue_out: 1200000,
    revenue_in: 1200000,
    revenue_total: 1200000,
    count_out: 1200000,
    count_in: 1200000,
    count_total: 1200000,
    capacity: 1200000,
    turnover: 1.42,
}
let currentRows: RowType[] = [
  {
    group_name: "LNBig",
    amount_out: 1200000,
    amount_in: 1200000,
    amount_total: 1200000,
    revenue_out: 1200000,
    revenue_in: 1200000,
    revenue_total: 1200000,
    count_out: 1200000,
    count_in: 1200000,
    count_total: 1200000,
    capacity: 1200000,
    turnover: 1.42,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },    {
    group_name: "LNBig",
    amount_out: 1,
    amount_in: 1,
    amount_total: 1,
    revenue_out: 1,
    revenue_in: 1,
    revenue_total: 1,
    count_out: 1,
    count_in: 1,
    count_total: 1,
    capacity: 1,
    turnover: 1,
  },
];
let pastRow: RowType[] = [
  {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },    {
    group_name: "LNBig",
    amount_out: 1000,
    amount_in: 1000,
    amount_total: 1000,
    revenue_out: 1000,
    revenue_in: 1000,
    revenue_total: 1000,
    count_out: 1000,
    count_in: 1000,
    count_total: 1000,
    capacity: 1000,
    turnover: 1000,
  },
];



function Table() {
    let key: keyof typeof columns
    let channel: keyof typeof currentRows
    return (
      <div className="table-wrapper">
        <style>
          {".table-content {grid-template-columns: repeat("+(Object.keys(columns).length+2)+",  minmax(min-content, auto))}"}
        </style>
        <div className="table-content">

          {/*Empty header at the start*/}
          {HeaderCell("", "first-empty-header", "empty locked")}

          {columns.map((column) => {
            return  HeaderCell(column.heading, column.key, "", column.locked)
          })}

          {/*Empty header at the end*/}
          {HeaderCell("", "first-empty-header")}

          {currentRows.map((currentRow, index) => {
            let returnedRow = columns.map((column) => {
              let key = column.key as keyof RowType
              let past = pastRow[index][key]
              switch (column.type) {
                case 'NameCell':
                  return NameCell((currentRow[key] as string), key, index)
                case 'NumericCell':
                  return NumericCell((currentRow[key] as number), (past as number), key, index)
                case "BarCell":
                  return BarCell((currentRow[key] as number), (past as number), (past as number), key, index)
                default:
                  return NumericCell((currentRow[key] as number), (past as number), key, index)
              }
            })
            // Add empty cells at the start and end of each row. This is to give the table a buffer at each end.
            returnedRow = [<div className={"cell empty locked"}></div>, ...returnedRow, <div className={"cell empty"}></div>]
            return returnedRow
          })}

          {/*Empty cell at the start*/}
          {<div className={"cell empty total-cell locked"}></div>}

          {columns.map((column) => {
            let key = column.key as keyof RowType
            switch (column.type) {
              case 'NameCell':
                return NameCell((totalRow[key] as string), key, "totals", "total-cell")
              case 'NumericCell':
                return NumericCell((totalRow[key] as number), (pastTotalRow[key] as number), key, "totals", "total-cell")
              case "BarCell":
                return BarCell((totalRow[key] as number), (pastTotalRow[key] as number), (pastTotalRow[key] as number), key, "totals", "total-cell")
              default:
                return NumericCell((totalRow[key] as number), (pastTotalRow[key] as number), key, "totals", "total-cell")
            }
          })}

          {/*Empty cell at the end*/}
          {<div className={"cell empty total-cell"}></div>}


        </div>
      </div>
    );
}

export default Table;
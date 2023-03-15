// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go


import { ColumnMetaData } from "features/table/types";
import { Invoice } from "features/transact/Invoices/invoiceTypes";

export const AllInvoicesColumns: ColumnMetaData<Invoice>[] = [
	{
		heading: "Creation Date",
		type: "DateCell",
		key: "creationDate",
		valueType: "date",
	},
	{
		heading: "Settle Date",
		type: "DateCell",
		key: "settleDate",
		valueType: "date",
	},
	{
		heading: "Settle State",
		type: "TextCell",
		key: "invoiceState",
		valueType: "enum",
		selectOptions: [
			{ label: "Open", value: "OPEN" },
			{ label: "Settled", value: "SETTLED" },
			{ label: "Canceled", value: "CANCELED" },
		],
	},
	{
		heading: "Paid Amount",
		type: "NumericCell",
		key: "amtPaid",
		valueType: "number",
	},
	{
		heading: "Memo",
		type: "TextCell",
		key: "memo",
		valueType: "string",
	},
	{
		heading: "Value",
		type: "NumericCell",
		key: "value",
		valueType: "number",
	},
	{
		heading: "Rebalance",
		type: "BooleanCell",
		key: "isRebalance",
		valueType: "boolean",
	},
	{
		heading: "Keysend",
		type: "BooleanCell",
		key: "isKeysend",
		valueType: "boolean",
	},
	{
		heading: "Destination",
		type: "LongTextCell",
		key: "destinationPubKey",
		valueType: "string",
	},
	{
		heading: "AMP",
		type: "BooleanCell",
		key: "isAmp",
		valueType: "boolean",
	},
	{
		heading: "Fallback Address",
		type: "LongTextCell",
		key: "fallbackAddr",
		valueType: "string",
	},
	{
		heading: "Payment Address",
		type: "LongTextCell",
		key: "paymentAddr",
		valueType: "string",
	},
	{
		heading: "Payment Request",
		type: "LongTextCell",
		key: "paymentRequest",
		valueType: "string",
	},
	{
		heading: "Private",
		type: "BooleanCell",
		key: "private",
		valueType: "boolean",
	},
	{
		heading: "Hash",
		type: "LongTextCell",
		key: "rHash",
		valueType: "string",
	},
	{
		heading: "Preimage",
		type: "LongTextCell",
		key: "rPreimage",
		valueType: "string",
	},
	{
		heading: "Expiry",
		type: "NumericCell",
		key: "expiry",
		valueType: "number",
	},
	{
		heading: "CLTV Expiry",
		type: "NumericCell",
		key: "cltvExpiry",
		valueType: "number",
	},
	{
		heading: "Updated On",
		type: "DateCell",
		key: "updatedOn",
		valueType: "date",
	},
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const InvoicesSortableColumns: Array<keyof Invoice> = [
	"creationDate",
	"settleDate",
	"invoiceState",
	"amtPaid",
	"memo",
	"value",
	"isRebalance",
	"isKeysend",
	"isAmp",
	"private",
	"expiry",
	"updatedOn",
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const InvoicesFilterableColumns: Array<keyof Invoice> = [
	"creationDate",
	"settleDate",
	"invoiceState",
	"amtPaid",
	"memo",
	"value",
	"isRebalance",
	"isKeysend",
	"destinationPubKey",
	"isAmp",
	"fallbackAddr",
	"paymentAddr",
	"paymentRequest",
	"private",
	"rHash",
	"rPreimage",
	"expiry",
	"cltvExpiry",
	"updatedOn",
];
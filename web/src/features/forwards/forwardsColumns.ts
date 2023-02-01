// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go


import { ColumnMetaData } from "features/table/types";
import { Forward } from "features/forwards/forwardsTypes";

export const AllForwardsColumns: ColumnMetaData<Forward>[] = [
	{
		heading: "Name",
		type: "AliasCell",
		key: "alias",
		valueType: "string",
		locked: true,
	},
	{
		heading: "Revenue",
		type: "BarCell",
		key: "revenueOut",
		valueType: "number",
	},
	{
		heading: "Total Forwards",
		type: "BarCell",
		key: "countTotal",
		valueType: "number",
	},
	{
		heading: "Outbound Amount",
		type: "BarCell",
		key: "amountOut",
		valueType: "number",
	},
	{
		heading: "Inbound Amount",
		type: "BarCell",
		key: "amountIn",
		valueType: "number",
	},
	{
		heading: "Total Amount",
		type: "BarCell",
		key: "amountTotal",
		valueType: "number",
	},
	{
		heading: "Tags",
		type: "TagsCell",
		key: "tags",
		valueType: "tag",
	},
	{
		heading: "Turnover Outbound",
		type: "BarCell",
		key: "turnoverOut",
		valueType: "number",
	},
	{
		heading: "Turnover Inbound",
		type: "BarCell",
		key: "turnoverIn",
		valueType: "number",
	},
	{
		heading: "Total Turnover",
		type: "BarCell",
		key: "turnoverTotal",
		valueType: "number",
	},
	{
		heading: "Outbound Forwards",
		type: "BarCell",
		key: "countOut",
		valueType: "number",
	},
	{
		heading: "Inbound Forwards",
		type: "BarCell",
		key: "countIn",
		valueType: "number",
	},
	{
		heading: "Revenue inbound",
		type: "BarCell",
		key: "revenueIn",
		valueType: "number",
	},
	{
		heading: "Revenue total",
		type: "BarCell",
		key: "revenueTotal",
		valueType: "number",
	},
	{
		heading: "Capacity",
		type: "NumericCell",
		key: "capacity",
		valueType: "number",
	},
	{
		heading: "Public key",
		type: "LongTextCell",
		key: "pubKey",
		valueType: "string",
	},
	{
		heading: "Funding Transaction",
		type: "LongTextCell",
		key: "fundingTransactionHash",
		valueType: "string",
	},
	{
		heading: "Funding Tx Output Index",
		type: "TextCell",
		key: "fundingOutputIndex",
		valueType: "string",
	},
	{
		heading: "Short Channel ID",
		type: "LongTextCell",
		key: "shortChannelId",
		valueType: "string",
	},
	{
		heading: "LND Short Channel ID",
		type: "LongTextCell",
		key: "lndShortChannelId",
		valueType: "string",
	},
	{
		heading: "Channel Point",
		type: "LongTextCell",
		key: "channelPoint",
		valueType: "string",
	},
	{
		heading: "Open",
		type: "BooleanCell",
		key: "open",
		valueType: "boolean",
	},
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ForwardsSortableColumns: Array<keyof Forward> = [
	"capacity",
	"shortChannelId",
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ForwardsFilterableColumns: Array<keyof Forward> = [
];
// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go


import { ColumnMetaData } from "features/table/types";
import { channel } from "features/channels/channelsTypes";

export const AllChannelsColumns: ColumnMetaData<channel>[] = [
	{
		heading: "Peer Alias",
		type: "AliasCell",
		key: "peerAlias",
		valueType: "string",
		locked: true,
	},
	{
		heading: "Active",
		type: "BooleanCell",
		key: "active",
		valueType: "boolean",
	},
	{
		heading: "Balance",
		type: "BalanceCell",
		key: "balance",
		valueType: "number",
	},
	{
		heading: "Tags",
		type: "TagsCell",
		key: "tags",
		valueType: "tag",
	},
	{
		heading: "Short Channel ID",
		type: "LongTextCell",
		key: "shortChannelId",
		valueType: "string",
	},
	{
		heading: "Channel Balance (%)",
		type: "BarCell",
		key: "gauge",
		valueType: "number",
	},
	{
		heading: "Remote Balance",
		type: "NumericCell",
		key: "remoteBalance",
		valueType: "number",
	},
	{
		heading: "Local Balance",
		type: "NumericCell",
		key: "localBalance",
		valueType: "number",
	},
	{
		heading: "Capacity",
		type: "NumericCell",
		key: "capacity",
		valueType: "number",
	},
	{
		heading: "Fee rate (PPM)",
		type: "NumericDoubleCell",
		key: "feeRateMilliMsat",
		valueType: "number",
		key2: "remoteFeeRateMilliMsat",
	},
	{
		heading: "Base Fee",
		type: "NumericDoubleCell",
		key: "feeBase",
		valueType: "number",
		key2: "remoteFeeBase",
	},
	{
		heading: "Remote Fee rate (PPM)",
		type: "NumericCell",
		key: "remoteFeeRateMilliMsat",
		valueType: "number",
	},
	{
		heading: "Remote Base Fee",
		type: "NumericCell",
		key: "remoteFeeBase",
		valueType: "number",
	},
	{
		heading: "Minimum HTLC",
		type: "NumericDoubleCell",
		key: "minHtlc",
		valueType: "number",
		key2: "remoteMinHtlc",
	},
	{
		heading: "Maximum HTLC",
		type: "NumericDoubleCell",
		key: "maxHtlc",
		valueType: "number",
		key2: "remoteMaxHtlc",
	},
	{
		heading: "Remote Minimum HTLC",
		type: "NumericCell",
		key: "remoteMinHtlc",
		valueType: "number",
	},
	{
		heading: "Remote Maximum HTLC",
		type: "NumericCell",
		key: "remoteMaxHtlc",
		valueType: "number",
	},
	{
		heading: "Time Lock Delta",
		type: "NumericDoubleCell",
		key: "timeLockDelta",
		valueType: "number",
		key2: "remoteTimeLockDelta",
	},
	{
		heading: "Remote Time Lock Delta",
		type: "NumericCell",
		key: "remoteTimeLockDelta",
		valueType: "number",
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
		heading: "Current BlockHeight",
		type: "NumericCell",
		key: "currentBlockHeight",
		valueType: "number",
	},
	{
		heading: "Funding Transaction",
		type: "LongTextCell",
		key: "fundingTransactionHash",
		valueType: "string",
	},
	{
		heading: "Funding BlockHeight",
		type: "NumericCell",
		key: "fundingBlockHeight",
		valueType: "number",
	},
	{
		heading: "Funding BlockHeight Delta",
		type: "NumericCell",
		key: "fundingBlockHeightDelta",
		valueType: "number",
	},
	{
		heading: "Funding Date",
		type: "DateCell",
		key: "fundedOn",
		valueType: "date",
	},
	{
		heading: "Funding Date Delta (Seconds)",
		type: "DurationCell",
		key: "fundedOnSecondsDelta",
		valueType: "duration",
	},
	{
		heading: "Closing Transaction",
		type: "LongTextCell",
		key: "closingTransactionHash",
		valueType: "string",
	},
	{
		heading: "Closing BlockHeight",
		type: "NumericCell",
		key: "closingBlockHeight",
		valueType: "number",
	},
	{
		heading: "Closing BlockHeight Delta",
		type: "NumericCell",
		key: "closingBlockHeightDelta",
		valueType: "number",
	},
	{
		heading: "Closing Date",
		type: "DateCell",
		key: "closedOn",
		valueType: "date",
	},
	{
		heading: "Closing Date Delta (Seconds)",
		type: "DurationCell",
		key: "closedOnSecondsDelta",
		valueType: "duration",
	},
	{
		heading: "Unsettled Balance",
		type: "NumericCell",
		key: "unsettledBalance",
		valueType: "number",
	},
	{
		heading: "Satoshis Sent",
		type: "NumericCell",
		key: "totalSatoshisSent",
		valueType: "number",
	},
	{
		heading: "Satoshis Received",
		type: "NumericCell",
		key: "totalSatoshisReceived",
		valueType: "number",
	},
	{
		heading: "Pending Forwarding HTLCs count",
		type: "NumericCell",
		key: "pendingForwardingHTLCsCount",
		valueType: "number",
	},
	{
		heading: "Pending Forwarding HTLCs",
		type: "NumericCell",
		key: "pendingForwardingHTLCsAmount",
		valueType: "number",
	},
	{
		heading: "Pending Forwarding HTLCs count",
		type: "NumericCell",
		key: "pendingLocalHTLCsCount",
		valueType: "number",
	},
	{
		heading: "Pending Forwarding HTLCs",
		type: "NumericCell",
		key: "pendingLocalHTLCsAmount",
		valueType: "number",
	},
	{
		heading: "Total Pending Forwarding HTLCs count",
		type: "NumericCell",
		key: "pendingTotalHTLCsCount",
		valueType: "number",
	},
	{
		heading: "Total Pending Forwarding HTLCs",
		type: "NumericCell",
		key: "pendingTotalHTLCsAmount",
		valueType: "number",
	},
	{
		heading: "Local Channel Reserve",
		type: "NumericDoubleCell",
		key: "localChanReserveSat",
		valueType: "number",
		key2: "remoteChanReserveSat",
	},
	{
		heading: "Remote Channel Reserve",
		type: "NumericCell",
		key: "remoteChanReserveSat",
		valueType: "number",
	},
	{
		heading: "Commit Fee",
		type: "NumericCell",
		key: "commitFee",
		valueType: "number",
	},
	{
		heading: "Node Name",
		type: "TextCell",
		key: "nodeName",
		valueType: "string",
	},
	{
		heading: "Mempool",
		type: "LinkCell",
		key: "mempoolSpace",
		valueType: "link",
	},
	{
		heading: "Amboss",
		type: "LinkCell",
		key: "ambossSpace",
		valueType: "link",
	},
	{
		heading: "1ML",
		type: "LinkCell",
		key: "oneMl",
		valueType: "link",
	},
	{
		heading: "Updates",
		type: "NumericCell",
		key: "numUpdates",
		valueType: "number",
	},
	{
		heading: "Peer Total Capacity",
		type: "NumericCell",
		key: "peerChannelCapacity",
		valueType: "number",
	},
	{
		heading: "Peer Channel Count",
		type: "NumericCell",
		key: "peerChannelCount",
		valueType: "number",
	},
	{
		heading: "Peer Total Local Balance",
		type: "NumericCell",
		key: "peerLocalBalance",
		valueType: "number",
	},
	{
		heading: "Peer Public Key",
		type: "LongTextCell",
		key: "remotePubkey",
		valueType: "string",
	},
	{
		heading: "Peer Channel Balance (%)",
		type: "BarCell",
		key: "peerGauge",
		valueType: "number",
	},
	{
		heading: "Private",
		type: "BooleanCell",
		key: "private",
		valueType: "boolean",
	},
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ChannelsSortableColumns: Array<keyof channel> = [
	"peerAlias",
	"active",
	"shortChannelId",
	"gauge",
	"remoteBalance",
	"localBalance",
	"capacity",
	"feeRateMilliMsat",
	"feeBase",
	"remoteFeeRateMilliMsat",
	"remoteFeeBase",
	"minHtlc",
	"maxHtlc",
	"remoteMinHtlc",
	"remoteMaxHtlc",
	"timeLockDelta",
	"remoteTimeLockDelta",
	"lndShortChannelId",
	"channelPoint",
	"currentBlockHeight",
	"fundingTransactionHash",
	"fundingBlockHeight",
	"fundingBlockHeightDelta",
	"fundedOn",
	"fundedOnSecondsDelta",
	"closingTransactionHash",
	"closingBlockHeight",
	"closingBlockHeightDelta",
	"closedOn",
	"closedOnSecondsDelta",
	"unsettledBalance",
	"totalSatoshisSent",
	"totalSatoshisReceived",
	"pendingForwardingHTLCsCount",
	"pendingForwardingHTLCsAmount",
	"pendingLocalHTLCsCount",
	"pendingLocalHTLCsAmount",
	"pendingTotalHTLCsCount",
	"pendingTotalHTLCsAmount",
	"localChanReserveSat",
	"remoteChanReserveSat",
	"commitFee",
	"nodeName",
	"numUpdates",
	"peerChannelCapacity",
	"peerChannelCount",
	"peerLocalBalance",
	"remotePubkey",
	"peerGauge",
	"private",
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ChannelsFilterableColumns: Array<keyof channel> = [
	"peerAlias",
	"active",
	"tags",
	"shortChannelId",
	"gauge",
	"remoteBalance",
	"localBalance",
	"capacity",
	"feeRateMilliMsat",
	"feeBase",
	"remoteFeeRateMilliMsat",
	"remoteFeeBase",
	"minHtlc",
	"maxHtlc",
	"remoteMinHtlc",
	"remoteMaxHtlc",
	"timeLockDelta",
	"remoteTimeLockDelta",
	"lndShortChannelId",
	"channelPoint",
	"currentBlockHeight",
	"fundingTransactionHash",
	"fundingBlockHeight",
	"fundingBlockHeightDelta",
	"fundedOn",
	"fundedOnSecondsDelta",
	"closingTransactionHash",
	"closingBlockHeight",
	"closingBlockHeightDelta",
	"closedOn",
	"closedOnSecondsDelta",
	"unsettledBalance",
	"totalSatoshisSent",
	"totalSatoshisReceived",
	"pendingForwardingHTLCsCount",
	"pendingForwardingHTLCsAmount",
	"pendingLocalHTLCsCount",
	"pendingLocalHTLCsAmount",
	"pendingTotalHTLCsCount",
	"pendingTotalHTLCsAmount",
	"localChanReserveSat",
	"remoteChanReserveSat",
	"commitFee",
	"nodeName",
	"numUpdates",
	"peerChannelCapacity",
	"peerChannelCount",
	"peerLocalBalance",
	"remotePubkey",
	"peerGauge",
	"private",
];
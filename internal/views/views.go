package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/jmoiron/sqlx/types"
	"golang.org/x/exp/slices"
)

type TableViewPage string

const (
	PageForwards            = TableViewPage("forwards")
	PageChannels            = TableViewPage("channel")
	PagePayments            = TableViewPage("payments")
	PageInvoices            = TableViewPage("invoices")
	PageOnChainTransactions = TableViewPage("onChain")
	PageWorkflows           = TableViewPage("workflow")
	PageTags                = TableViewPage("tag")
)

type NewTableView struct {
	View types.JSONText `json:"view" db:"view"`
	Page string         `json:"page" db:"page"`
}

type UpdateTableView struct {
	Id      int            `json:"id" db:"id"`
	View    types.JSONText `json:"view" db:"view"`
	Version string         `json:"version" db:"version"`
}

type TableViewOrder struct {
	Id        int `json:"id" db:"id"`
	ViewOrder int `json:"viewOrder" db:"view_order"`
}

type TableViewLayout struct {
	Id        int            `json:"id" db:"id"`
	View      types.JSONText `json:"view" db:"view"`
	Page      string         `json:"page" db:"page"`
	ViewOrder int            `json:"viewOrder" db:"view_order"`
	Version   string         `json:"version" db:"version"`
}

type TableViewResponses struct {
	Forwards []TableViewLayout `json:"forwards"`
	Channel  []TableViewLayout `json:"channel"`
	Payments []TableViewLayout `json:"payments"`
	Invoices []TableViewLayout `json:"invoices"`
	OnChain  []TableViewLayout `json:"onChain"`
}

type TableViewStructured struct {
	TableViewId int                `json:"tableViewId"`
	Page        string             `json:"page"`
	Title       string             `json:"title"`
	Order       int                `json:"order"`
	UpdateOn    time.Time          `json:"updatedOn"`
	Columns     []TableViewColumn  `json:"columns"`
	Filters     []TableViewFilter  `json:"filters"`
	Sortings    []TableViewSorting `json:"sortings"`
}

type TableView struct {
	TableViewId int       `json:"tableViewId" db:"table_view_id"`
	Page        string    `json:"page" db:"page"`
	Title       string    `json:"title" db:"title"`
	Order       int       `json:"order" db:"order"`
	CreatedOn   time.Time `json:"createdOn" db:"created_on"`
	UpdateOn    time.Time `json:"updatedOn" db:"updated_on"`
}

type TableViewColumn struct {
	TableViewColumnId int       `json:"TableViewColumnId" db:"table_view_column_id"`
	Key               string    `json:"key" db:"key"`
	KeySecond         *string   `json:"keySecond" db:"key_second"`
	Order             int       `json:"order" db:"order"`
	Type              string    `json:"type" db:"type"`
	TableViewId       int       `json:"tableViewId" db:"table_view_id"`
	CreatedOn         time.Time `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time `json:"updatedOn" db:"updated_on"`
}

type TableViewFilter struct {
	TableViewFilterId int            `json:"TableViewFilterId" db:"table_view_filter_id"`
	Filter            types.JSONText `json:"filter" db:"filter"`
	TableViewId       int            `json:"tableViewId" db:"table_view_id"`
	CreatedOn         time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time      `json:"updatedOn" db:"updated_on"`
}

type TableViewSorting struct {
	TableViewSortingId int       `json:"TableViewSortingId" db:"table_view_sorting_id"`
	Key                string    `json:"key" db:"key"`
	Order              int       `json:"order" db:"order"`
	Ascending          bool      `json:"ascending" db:"ascending"`
	TableViewId        int       `json:"tableViewId" db:"table_view_id"`
	CreatedOn          time.Time `json:"createdOn" db:"created_on"`
	UpdateOn           time.Time `json:"updatedOn" db:"updated_on"`
}

type tableViewColumnDefinition struct {
	key           string
	locked        bool
	sortable      bool
	filterable    bool
	heading       string
	visualType    string
	keySecond     string
	valueType     string
	suffix        string
	selectOptions []tableViewSelectOptions
	pages         map[TableViewPage]int
}

type tableViewSelectOptions struct {
	label string
	value string
}

func getTableViewColumnDefinition(key string) tableViewColumnDefinition {
	for _, definition := range getTableViewColumnDefinitions() {
		if definition.key == key {
			return definition
		}
	}
	return tableViewColumnDefinition{}
}

//go:generate go run ../../cmd/torq/internal/generators/gen.go
func GetTableViewColumnDefinitionsForPage(page TableViewPage) string {
	var disclaimer string
	disclaimer = disclaimer + "// DO NOT EDIT THIS FILE...\n"
	disclaimer = disclaimer + "// This File is generated by go:generate\n"
	disclaimer = disclaimer + "// For more information look at cmd/torq/gen.go\n"

	result := disclaimer
	result = result + "\n\nimport { ColumnMetaData } from \"features/table/types\";"
	switch page {
	case PageChannels:
		result = result + "\nimport { channel } from \"features/channels/channelsTypes\";"
		result = result + "\n\nexport const AllChannelsColumns: ColumnMetaData<channel>[] = ["
	case PageForwards:
		result = result + "\nimport { Forward } from \"features/forwards/forwardsTypes\";"
		result = result + "\n\nexport const AllForwardsColumns: ColumnMetaData<Forward>[] = ["
	case PageInvoices:
		result = result + "\nimport { Invoice } from \"features/transact/Invoices/invoiceTypes\";"
		result = result + "\n\nexport const AllInvoicesColumns: ColumnMetaData<Invoice>[] = ["
	case PageOnChainTransactions:
		result = result + "\nimport { OnChainTx } from \"features/transact/OnChain/types\";"
		result = result + "\n\nexport const AllOnChainTransactionsColumns: ColumnMetaData<OnChainTx>[] = ["
	case PagePayments:
		result = result + "\nimport { Payment } from \"features/transact/Payments/types\";"
		result = result + "\n\nexport const AllPaymentsColumns: ColumnMetaData<Payment>[] = ["
	case PageTags:
		result = result + "\nimport { ExpandedTag } from \"pages/tags/tagsTypes\";"
		result = result + "\n\nexport const AllTagsColumns: ColumnMetaData<ExpandedTag>[] = ["
	}
	for _, definition := range getTableViewColumnDefinitionsSorted(page) {
		result = result + fmt.Sprintf(
			"\n\t{\n\t\theading: \"%v\",\n\t\ttype: \"%v\",\n\t\tkey: \"%v\",\n\t\tvalueType: \"%v\",",
			definition.heading, definition.visualType, definition.key, definition.valueType)
		if definition.locked {
			result = result + "\n\t\tlocked: true,"
		}
		if definition.keySecond != "" {
			result = result + fmt.Sprintf("\n\t\tkey2: \"%v\",", definition.keySecond)
		}
		result = result + "\n\t},"
	}
	result = result + "\n];"

	result = result + "\n\n\n" + disclaimer
	switch page {
	case PageChannels:
		result = result + "\nexport const ChannelsSortableColumns: Array<keyof channel> = ["
	case PageForwards:
		result = result + "\nexport const ForwardsSortableColumns: Array<keyof Forward> = ["
	case PageInvoices:
		result = result + "\nexport const InvoicesSortableColumns: Array<keyof Invoice> = ["
	case PageOnChainTransactions:
		result = result + "\nexport const OnChainTransactionsSortableColumns: Array<keyof OnChainTx> = ["
	case PagePayments:
		result = result + "\nexport const PaymentsSortableColumns: Array<keyof Payment> = ["
	case PageTags:
		result = result + "\nexport const TagsSortableColumns: Array<keyof ExpandedTag> = ["
	}
	for _, definition := range getTableViewColumnDefinitionsSorted(page) {
		if definition.sortable {
			result = result + fmt.Sprintf("\n\t\"%v\",", definition.key)
		}
	}
	result = result + "\n];"

	result = result + "\n\n\n" + disclaimer
	switch page {
	case PageChannels:
		result = result + "\nexport const ChannelsFilterableColumns: Array<keyof channel> = ["
	case PageForwards:
		result = result + "\nexport const ForwardsFilterableColumns: Array<keyof Forward> = ["
	case PageInvoices:
		result = result + "\nexport const InvoicesFilterableColumns: Array<keyof Invoice> = ["
	case PageOnChainTransactions:
		result = result + "\nexport const OnChainTransactionsFilterableColumns: Array<keyof OnChainTx> = ["
	case PagePayments:
		result = result + "\nexport const PaymentsFilterableColumns: Array<keyof Payment> = ["
	case PageTags:
		result = result + "\nexport const TagsFilterableColumns: Array<keyof ExpandedTag> = ["
	}
	for _, definition := range getTableViewColumnDefinitionsSorted(page) {
		if definition.filterable {
			result = result + fmt.Sprintf("\n\t\"%v\",", definition.key)
		}
	}
	result = result + "\n];"

	return result
}

func getTableViewColumnDefinitionsSorted(page TableViewPage) []tableViewColumnDefinition {
	pageDefinitions := make(map[int]tableViewColumnDefinition)
	pageDefinitionPriorities := make([]int, 0)
	for _, definition := range getTableViewColumnDefinitions() {
		for definitionPage, priority := range definition.pages {
			if definitionPage == page {
				pageDefinitions[priority] = definition
				if slices.Contains(pageDefinitionPriorities, priority) {
					panic("Development Failure duplicated priority found for page columns of: " + page)
				}
				pageDefinitionPriorities = append(pageDefinitionPriorities, priority)
			}
		}
	}
	sort.Ints(pageDefinitionPriorities)
	var results []tableViewColumnDefinition
	for _, priority := range pageDefinitionPriorities {
		results = append(results, pageDefinitions[priority])
	}
	return results
}
func getTableViewColumnDefinitions() []tableViewColumnDefinition {
	return []tableViewColumnDefinition{
		{
			key:        "peerAlias",
			locked:     true,
			sortable:   true,
			heading:    "Peer Alias",
			visualType: "AliasCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 1,
			},
		},
		{
			key:        "active",
			sortable:   true,
			heading:    "Active",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageChannels: 2,
			},
		},
		{
			key:        "balance",
			heading:    "Balance",
			visualType: "BalanceCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 3,
			},
		},
		{
			key:        "tags",
			heading:    "Tags",
			visualType: "TagsCell",
			valueType:  "tag",
			pages: map[TableViewPage]int{
				PageChannels: 4,
				PageForwards: 7,
			},
		},
		{
			key:        "shortChannelId",
			sortable:   true,
			heading:    "Short Channel ID",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 5,
				PageForwards: 19,
			},
		},
		{
			key:        "gauge",
			sortable:   true,
			heading:    "Channel Balance (%)",
			visualType: "BarCell",
			valueType:  "number",
			suffix:     "%",
			pages: map[TableViewPage]int{
				PageChannels: 6,
			},
		},
		{
			key:        "remoteBalance",
			sortable:   true,
			heading:    "Remote Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 7,
			},
		},
		{
			key:        "localBalance",
			sortable:   true,
			heading:    "Local Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 8,
			},
		},
		{
			key:        "capacity",
			sortable:   true,
			heading:    "Capacity",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 9,
				PageForwards: 15,
			},
		},
		{
			key:        "feeRateMilliMsat",
			sortable:   true,
			heading:    "Fee rate (PPM)",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteFeeRateMilliMsat",
			valueType:  "number",
			suffix:     "ppm",
			pages: map[TableViewPage]int{
				PageChannels: 10,
			},
		},
		{
			key:        "feeBase",
			sortable:   true,
			heading:    "Base Fee",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteFeeBase",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 11,
			},
		},
		{
			key:        "remoteFeeRateMilliMsat",
			sortable:   true,
			heading:    "Remote Fee rate (PPM)",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "ppm",
			pages: map[TableViewPage]int{
				PageChannels: 12,
			},
		},
		{
			key:        "remoteFeeBase",
			sortable:   true,
			heading:    "Remote Base Fee",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 13,
			},
		},
		{
			key:        "minHtlc",
			sortable:   true,
			heading:    "Minimum HTLC",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteMinHtlc",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 14,
			},
		},
		{
			key:        "maxHtlc",
			sortable:   true,
			heading:    "Maximum HTLC",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteMaxHtlc",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 15,
			},
		},
		{
			key:        "remoteMinHtlc",
			sortable:   true,
			heading:    "Remote Minimum HTLC",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 16,
			},
		},
		{
			key:        "remoteMaxHtlc",
			sortable:   true,
			heading:    "Remote Maximum HTLC",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 17,
			},
		},
		{
			key:        "timeLockDelta",
			heading:    "Time Lock Delta",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteTimeLockDelta",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 18,
			},
		},
		{
			key:        "remoteTimeLockDelta",
			heading:    "Remote Time Lock Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 19,
			},
		},
		{
			key:        "lndShortChannelId",
			heading:    "LND Short Channel ID",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 20,
				PageForwards: 20,
			},
		},
		{
			key:        "channelPoint",
			heading:    "Channel Point",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 21,
				PageForwards: 21,
			},
		},
		{
			key:        "currentBlockHeight",
			sortable:   true,
			heading:    "Current BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 22,
			},
		},
		{
			key:        "fundingTransactionHash",
			heading:    "Funding Transaction",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 23,
				PageForwards: 17,
			},
		},
		{
			key:        "fundingBlockHeight",
			sortable:   true,
			heading:    "Funding BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 24,
			},
		},
		{
			key:        "fundingBlockHeightDelta",
			sortable:   true,
			heading:    "Funding BlockHeight Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 25,
			},
		},
		{
			key:        "fundedOn",
			heading:    "Funding Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageChannels: 26,
			},
		},
		{
			key:        "fundedOnSecondsDelta",
			sortable:   true,
			heading:    "Funding Date Delta (Seconds)",
			visualType: "DurationCell",
			valueType:  "duration",
			pages: map[TableViewPage]int{
				PageChannels: 27,
			},
		},
		{
			key:        "closingTransactionHash",
			heading:    "Closing Transaction",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 28,
			},
		},
		{
			key:        "closingBlockHeight",
			sortable:   true,
			heading:    "Closing BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 29,
			},
		},
		{
			key:        "closingBlockHeightDelta",
			sortable:   true,
			heading:    "Closing BlockHeight Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 30,
			},
		},
		{
			key:        "closedOn",
			heading:    "Closing Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageChannels: 31,
			},
		},
		{
			key:        "closedOnSecondsDelta",
			sortable:   true,
			heading:    "Closing Date Delta (Seconds)",
			visualType: "DurationCell",
			valueType:  "duration",
			pages: map[TableViewPage]int{
				PageChannels: 32,
			},
		},
		{
			key:        "unsettledBalance",
			sortable:   true,
			heading:    "Unsettled Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 33,
			},
		},
		{
			key:        "totalSatoshisSent",
			sortable:   true,
			heading:    "Satoshis Sent",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 34,
			},
		},
		{
			key:        "totalSatoshisReceived",
			sortable:   true,
			heading:    "Satoshis Received",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 35,
			},
		},
		{
			key:        "pendingForwardingHTLCsCount",
			heading:    "Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 36,
			},
		},
		{
			key:        "pendingForwardingHTLCsAmount",
			heading:    "Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 37,
			},
		},
		{
			key:        "pendingLocalHTLCsCount",
			heading:    "Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 38,
			},
		},
		{
			key:        "pendingLocalHTLCsAmount",
			heading:    "Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 39,
			},
		},
		{
			key:        "pendingTotalHTLCsCount",
			heading:    "Total Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 40,
			},
		},
		{
			key:        "pendingTotalHTLCsAmount",
			heading:    "Total Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 41,
			},
		},
		{
			key:        "localChanReserveSat",
			sortable:   true,
			heading:    "Local Channel Reserve",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteChanReserveSat",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 42,
			},
		},
		{
			key:        "remoteChanReserveSat",
			sortable:   true,
			heading:    "Remote Channel Reserve",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: map[TableViewPage]int{
				PageChannels: 43,
			},
		},
		{
			key:        "commitFee",
			sortable:   true,
			heading:    "Commit Fee",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageChannels: 44,
			},
		},
		{
			key:        "nodeName",
			sortable:   true,
			heading:    "Node Name",
			visualType: "AliasCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageChannels: 45,
			},
		},
		{
			key:        "mempoolSpace",
			heading:    "Mempool",
			visualType: "LinkCell",
			valueType:  "link",
			pages: map[TableViewPage]int{
				PageChannels: 46,
			},
		},
		{
			key:        "ambossSpace",
			heading:    "Amboss",
			visualType: "LinkCell",
			valueType:  "link",
			pages: map[TableViewPage]int{
				PageChannels: 47,
			},
		},
		{
			key:        "oneMl",
			heading:    "1ML",
			visualType: "LinkCell",
			valueType:  "link",
			pages: map[TableViewPage]int{
				PageChannels: 48,
			},
		},
		{
			key:        "date",
			sortable:   true,
			filterable: true,
			heading:    "Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 1,
				PagePayments:            1,
			},
		},
		{
			key:        "amount",
			sortable:   true,
			filterable: true,
			heading:    "Amount",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 2,
			},
		},
		{
			key:        "totalFees",
			sortable:   true,
			filterable: true,
			heading:    "Fees",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 3,
			},
		},
		{
			key:        "txHash",
			heading:    "Tx Hash",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 4,
			},
		},
		{
			key:        "lndShortChanId",
			sortable:   true,
			filterable: true,
			heading:    "LND Short Channel ID",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 5,
			},
		},
		{
			key:        "lndTxTypeLabel",
			sortable:   true,
			filterable: true,
			heading:    "LND Tx type label",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 6,
			},
		},
		{
			key:        "destAddressesCount",
			sortable:   true,
			filterable: true,
			heading:    "Destination Addresses Count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 7,
			},
		},
		{
			key:        "label",
			sortable:   true,
			filterable: true,
			heading:    "Label",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageOnChainTransactions: 8,
			},
		},
		{
			key:        "alias",
			locked:     true,
			heading:    "Name",
			visualType: "AliasCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageForwards: 1,
			},
		},
		{
			key:        "revenueOut",
			heading:    "Revenue",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 2,
			},
		},
		{
			key:        "countTotal",
			heading:    "Total Forwards",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 3,
			},
		},
		{
			key:        "amountOut",
			heading:    "Outbound Amount",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 4,
			},
		},
		{
			key:        "amountIn",
			heading:    "Inbound Amount",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 5,
			},
		},
		{
			key:        "amountTotal",
			heading:    "Total Amount",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 6,
			},
		},
		{
			key:        "turnoverOut",
			heading:    "Turnover Outbound",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 8,
			},
		},
		{
			key:        "turnoverIn",
			heading:    "Turnover Inbound",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 9,
			},
		},
		{
			key:        "turnoverTotal",
			heading:    "Total Turnover",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 10,
			},
		},
		{
			key:        "countOut",
			heading:    "Outbound Forwards",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 11,
			},
		},
		{
			key:        "countIn",
			heading:    "Inbound Forwards",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 12,
			},
		},
		{
			key:        "revenueIn",
			heading:    "Revenue inbound",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 13,
			},
		},
		{
			key:        "revenueTotal",
			heading:    "Revenue total",
			visualType: "BarCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageForwards: 14,
			},
		},
		{
			key:        "pubKey",
			heading:    "Public key",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageForwards: 16,
			},
		},
		{
			key:        "fundingOutputIndex",
			heading:    "Funding Tx Output Index",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageForwards: 18,
			},
		},
		{
			key:        "open",
			heading:    "Open",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageForwards: 22,
			},
		},
		{
			key:        "creationDate",
			sortable:   true,
			filterable: true,
			heading:    "Creation Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageInvoices: 1,
			},
		},
		{
			key:        "settleDate",
			sortable:   true,
			filterable: true,
			heading:    "Settle Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageInvoices: 2,
			},
		},
		{
			key:        "invoiceState",
			sortable:   true,
			filterable: true,
			heading:    "Settle Date",
			visualType: "TextCell",
			valueType:  "enum",
			selectOptions: []tableViewSelectOptions{
				{label: "Open", value: "OPEN"},
				{label: "Settled", value: "SETTLED"},
				{label: "Canceled", value: "CANCELED"},
			},
			pages: map[TableViewPage]int{
				PageInvoices: 3,
			},
		},
		{
			key:        "amtPaid",
			sortable:   true,
			filterable: true,
			heading:    "Paid Amount",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageInvoices: 4,
			},
		},
		{
			key:        "memo",
			sortable:   true,
			filterable: true,
			heading:    "Memo",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 5,
			},
		},
		{
			key:        "value",
			sortable:   true,
			filterable: true,
			heading:    "Value",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageInvoices: 6,
				PagePayments: 3,
			},
		},
		{
			key:        "isRebalance",
			sortable:   true,
			filterable: true,
			heading:    "Rebalance",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageInvoices: 7,
				PagePayments: 6,
			},
		},
		{
			key:        "isKeysend",
			sortable:   true,
			filterable: true,
			heading:    "Keysend",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageInvoices: 8,
			},
		},
		{
			key:        "destinationPubKey",
			filterable: true,
			heading:    "Destination",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 9,
				PagePayments: 12,
			},
		},
		{
			key:        "isAmp",
			sortable:   true,
			filterable: true,
			heading:    "AMP",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageInvoices: 10,
			},
		},
		{
			key:        "fallbackAddr",
			filterable: true,
			heading:    "Fallback Address",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 11,
			},
		},
		{
			key:        "paymentAddr",
			filterable: true,
			heading:    "Payment Address",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 12,
			},
		},
		{
			key:        "paymentRequest",
			filterable: true,
			heading:    "Payment Request",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 13,
			},
		},
		{
			key:        "private",
			sortable:   true,
			filterable: true,
			heading:    "Private",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageInvoices: 14,
			},
		},
		{
			key:        "rHash",
			filterable: true,
			heading:    "Hash",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 15,
			},
		},
		{
			key:        "rPreimage",
			filterable: true,
			heading:    "Preimage",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageInvoices: 16,
			},
		},
		{
			key:        "expiry",
			sortable:   true,
			filterable: true,
			heading:    "Expiry",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageInvoices: 17,
			},
		},
		{
			key:        "cltvExpiry",
			filterable: true,
			heading:    "CLTV Expiry",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PageInvoices: 18,
			},
		},
		{
			key:        "updatedOn",
			sortable:   true,
			filterable: true,
			heading:    "Updated On",
			visualType: "DateCell",
			valueType:  "date",
			pages: map[TableViewPage]int{
				PageInvoices: 19,
			},
		},
		{
			key:        "status",
			sortable:   true,
			filterable: true,
			heading:    "Status",
			visualType: "TextCell",
			valueType:  "array",
			selectOptions: []tableViewSelectOptions{
				{label: "Succeeded", value: "SUCCEEDED"},
				{label: "In Flight", value: "IN_FLIGHT"},
				{label: "Failed", value: "FAILED"},
			},
			pages: map[TableViewPage]int{
				PagePayments: 2,
			},
		},
		{
			key:        "fee",
			sortable:   true,
			filterable: true,
			heading:    "Fee",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PagePayments: 4,
			},
		},
		{
			key:        "ppm",
			sortable:   true,
			filterable: true,
			heading:    "PPM",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PagePayments: 5,
			},
		},
		{
			key:        "secondsInFlight",
			sortable:   true,
			filterable: true,
			heading:    "Seconds In Flight",
			visualType: "DurationCell",
			valueType:  "duration",
			pages: map[TableViewPage]int{
				PagePayments: 7,
			},
		},
		{
			key:        "failureReason",
			sortable:   true,
			filterable: true,
			heading:    "Failure Reason",
			visualType: "TextCell",
			valueType:  "array",
			selectOptions: []tableViewSelectOptions{
				{value: "FAILURE_REASON_NONE", label: "None"},
				{value: "FAILURE_REASON_TIMEOUT", label: "Timeout"},
				{value: "FAILURE_REASON_NO_ROUTE", label: "No Route"},
				{value: "FAILURE_REASON_ERROR", label: "Error"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_DETAILS", label: "Incorrect Payment Details"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_AMOUNT", label: "Incorrect Payment Amount"},
				{value: "FAILURE_REASON_PAYMENT_HASH_MISMATCH", label: "Payment Hash Mismatch"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_REQUEST", label: "Incorrect Payment Request"},
				{value: "FAILURE_REASON_UNKNOWN", label: "Unknown"},
			},
			pages: map[TableViewPage]int{
				PagePayments: 8,
			},
		},
		{
			key:        "isMpp",
			filterable: true,
			heading:    "MPP",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PagePayments: 9,
			},
		},
		{
			key:        "countFailedAttempts",
			sortable:   true,
			filterable: true,
			heading:    "Failed Attempts",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PagePayments: 10,
			},
		},
		{
			key:        "countSuccessfulAttempts",
			sortable:   true,
			filterable: true,
			heading:    "Successful Attempts",
			visualType: "NumericCell",
			valueType:  "number",
			pages: map[TableViewPage]int{
				PagePayments: 11,
			},
		},
		{
			key:        "paymentHash",
			filterable: true,
			heading:    "Payment Hash",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PagePayments: 13,
			},
		},
		{
			key:        "paymentPreimage",
			filterable: true,
			heading:    "Payment Preimage",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PagePayments: 14,
			},
		},
		{
			key:        "workflowName",
			heading:    "Name",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageWorkflows: 1,
			},
		},
		{
			key:        "workflowStatus",
			heading:    "Active",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: map[TableViewPage]int{
				PageWorkflows: 2,
			},
		},
		{
			key:        "latestVersionName",
			heading:    "Latest Draft",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageWorkflows: 3,
			},
		},
		{
			key:        "activeVersionName",
			heading:    "Active Version",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageWorkflows: 4,
			},
		},
		{
			key:        "name",
			heading:    "Name",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageTags: 1,
			},
		},
		{
			key:        "categoryId",
			heading:    "Category",
			visualType: "TextCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageTags: 2,
			},
		},
		{
			key:        "channels",
			heading:    "Applied to",
			visualType: "NumericCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageTags: 3,
			},
		},
		{
			key:        "edit",
			heading:    "Edit",
			visualType: "EditCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageTags: 4,
			},
		},
		{
			key:        "delete",
			heading:    "Delete",
			visualType: "EditCell",
			valueType:  "string",
			pages: map[TableViewPage]int{
				PageTags: 5,
			},
		},
	}
}

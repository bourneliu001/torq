package channels

import (
	"time"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/vector"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/database"
)

// GetClosureStatus returns Closing when our API is outdated and a new lnrpc.ChannelCloseSummary_ClosureType is added
func GetClosureStatus(lndClosureType lnrpc.ChannelCloseSummary_ClosureType) core.ChannelStatus {
	switch lndClosureType {
	case lnrpc.ChannelCloseSummary_COOPERATIVE_CLOSE:
		return core.CooperativeClosed
	case lnrpc.ChannelCloseSummary_LOCAL_FORCE_CLOSE:
		return core.LocalForceClosed
	case lnrpc.ChannelCloseSummary_REMOTE_FORCE_CLOSE:
		return core.RemoteForceClosed
	case lnrpc.ChannelCloseSummary_BREACH_CLOSE:
		return core.BreachClosed
	case lnrpc.ChannelCloseSummary_FUNDING_CANCELED:
		return core.FundingCancelledClosed
	case lnrpc.ChannelCloseSummary_ABANDONED:
		return core.AbandonedClosed
	}
	return core.Closing
}

type Channel struct {
	// ChannelID A database primary key. NOT a channel_id as specified in BOLT 2
	ChannelID              int                `json:"channelId" db:"channel_id"`
	ShortChannelID         *string            `json:"shortChannelId" db:"short_channel_id"`
	FundingTransactionHash *string            `json:"fundingTransactionHash" db:"funding_transaction_hash"`
	FundingOutputIndex     *int               `json:"fundingOutputIndex" db:"funding_output_index"`
	ClosingTransactionHash *string            `json:"closingTransactionHash" db:"closing_transaction_hash"`
	LNDShortChannelID      *uint64            `json:"lndShortChannelId" db:"lnd_short_channel_id"`
	Capacity               int64              `json:"capacity" db:"capacity"`
	Private                bool               `json:"private" db:"private"`
	FirstNodeId            int                `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId           int                `json:"secondNodeId" db:"second_node_id"`
	InitiatingNodeId       *int               `json:"initiatingNodeId" db:"initiating_node_id"`
	AcceptingNodeId        *int               `json:"acceptingNodeId" db:"accepting_node_id"`
	ClosingNodeId          *int               `json:"closingNodeId" db:"closing_node_id"`
	Status                 core.ChannelStatus `json:"status" db:"status_id"`
	CreatedOn              time.Time          `json:"createdOn" db:"created_on"`
	UpdateOn               *time.Time         `json:"updatedOn" db:"updated_on"`
	FundingBlockHeight     *uint32            `json:"fundingBlockHeight" db:"funding_block_height"`
	FundedOn               *time.Time         `json:"fundedOn" db:"funded_on"`
	ClosingBlockHeight     *uint32            `json:"closingBlockHeight" db:"closing_block_height"`
	ClosedOn               *time.Time         `json:"closedOn" db:"closed_on"`
	Flags                  core.ChannelFlags  `json:"flags" db:"flags"`
}

func (channel *Channel) AddChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.AddChannelFlag(flags)
}
func (channel *Channel) HasChannelFlags(flags core.ChannelFlags) bool {
	return channel.Flags.HasChannelFlag(flags)
}
func (channel *Channel) RemoveChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.RemoveChannelFlag(flags)
}

func AddChannelOrUpdateChannelStatus(db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache,
	channel Channel) (int, error) {

	channel, err := addChannelOrUpdateChannelStatus(db, nodeSettings, channel)
	if err != nil {
		return 0, errors.Wrapf(err, "add channel (or update channel status) for nodeId: %v", nodeSettings.NodeId)
	}
	cache.SetChannel(channel.ChannelID, channel.ShortChannelID, channel.LNDShortChannelID, channel.Status,
		channel.FundingTransactionHash, channel.FundingOutputIndex,
		channel.FundingBlockHeight, channel.FundedOn,
		channel.Capacity, channel.Private, channel.FirstNodeId, channel.SecondNodeId,
		channel.InitiatingNodeId, channel.AcceptingNodeId,
		channel.ClosingTransactionHash, channel.ClosingNodeId, channel.ClosingBlockHeight, channel.ClosedOn,
		channel.Flags)
	if channel.Status >= core.Closing {
		cache.RemoveChannelStateFromCache(channel.ChannelID)
	}
	return channel.ChannelID, nil
}

func addChannelOrUpdateChannelStatus(db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache,
	channel Channel) (Channel, error) {

	channel, existingChannelId, err := createWhenNonExisting(db, channel)
	if err != nil {
		return channel, errors.Wrap(err, "create when non existing")
	}
	if existingChannelId == 0 {
		return channel, nil
	}
	// existingChannelId is known, or it would have returned
	channel.ChannelID = existingChannelId
	existingChannelSettings := cache.GetChannelSettingByChannelId(existingChannelId)
	if existingChannelSettings.ChannelId == 0 {
		existingChannel, err := GetChannel(db, existingChannelId)
		if err != nil {
			return Channel{}, errors.Wrapf(err, "obtaining existing channel for channelId: %v.", existingChannelId)
		}
		existingChannelSettings = cache.ChannelSettingsCache{
			ChannelId:              existingChannel.ChannelID,
			ShortChannelId:         existingChannel.ShortChannelID,
			LndShortChannelId:      existingChannel.LNDShortChannelID,
			FundingTransactionHash: existingChannel.FundingTransactionHash,
			FundingOutputIndex:     existingChannel.FundingOutputIndex,
			FundingBlockHeight:     existingChannel.FundingBlockHeight,
			FundedOn:               existingChannel.FundedOn,
			Capacity:               existingChannel.Capacity,
			FirstNodeId:            existingChannel.FirstNodeId,
			SecondNodeId:           existingChannel.SecondNodeId,
			InitiatingNodeId:       existingChannel.InitiatingNodeId,
			AcceptingNodeId:        existingChannel.AcceptingNodeId,
			Private:                existingChannel.Private,
			Status:                 existingChannel.Status,
			ClosingTransactionHash: existingChannel.ClosingTransactionHash,
			ClosingNodeId:          existingChannel.ClosingNodeId,
			ClosingBlockHeight:     existingChannel.ClosingBlockHeight,
			ClosedOn:               existingChannel.ClosedOn,
			Flags:                  existingChannel.Flags,
		}
	}
	switch channel.Status {
	case core.CooperativeClosed, core.LocalForceClosed, core.RemoteForceClosed, core.BreachClosed:
		if channel.ClosingTransactionHash != nil && *channel.ClosingTransactionHash != "" &&
			!existingChannelSettings.HasChannelFlags(core.ClosedOn) &&
			vector.IsVectorAvailable(nodeSettings) {

			vectorResponse := vector.GetTransactionDetailsFromVector(*channel.ClosingTransactionHash, nodeSettings)
			if vectorResponse.BlockHeight != 0 {
				channel.ClosingBlockHeight = &vectorResponse.BlockHeight
				channel.ClosedOn = &vectorResponse.BlockTimestamp
				channel.AddChannelFlags(core.ClosedOn)
			}
		}
		if existingChannelSettings.ClosingBlockHeight == nil || *existingChannelSettings.ClosingBlockHeight == 0 &&
			(channel.FundingBlockHeight == nil || *channel.FundingBlockHeight == 0) {
			currentBlockHeight := cache.GetBlockHeight()
			channel.ClosingBlockHeight = &currentBlockHeight
			channel.RemoveChannelFlags(core.ClosedOn)
		}
		if existingChannelSettings.ClosedOn == nil && channel.ClosedOn == nil {
			now := time.Now().UTC()
			channel.ClosedOn = &now
			channel.RemoveChannelFlags(core.ClosedOn)
		}
		fallthrough
	case core.Open, core.Closing:
		if channel.FundingTransactionHash != nil &&
			*channel.FundingTransactionHash != "" &&
			!existingChannelSettings.HasChannelFlags(core.FundedOn) &&
			vector.IsVectorAvailable(nodeSettings) {

			vectorResponse := vector.GetTransactionDetailsFromVector(*channel.FundingTransactionHash, nodeSettings)
			if vectorResponse.BlockHeight != 0 {
				channel.FundingBlockHeight = &vectorResponse.BlockHeight
				channel.FundedOn = &vectorResponse.BlockTimestamp
				channel.AddChannelFlags(core.FundedOn)
			}
		}
		if (existingChannelSettings.FundingBlockHeight == nil || *existingChannelSettings.FundingBlockHeight == 0) &&
			(channel.FundingBlockHeight == nil || *channel.FundingBlockHeight == 0) {
			currentBlockHeight := cache.GetBlockHeight()
			channel.FundingBlockHeight = &currentBlockHeight
			channel.RemoveChannelFlags(core.FundedOn)
		}
		if existingChannelSettings.FundedOn == nil && channel.FundedOn == nil {
			now := time.Now().UTC()
			channel.FundedOn = &now
			channel.RemoveChannelFlags(core.FundedOn)
		}
	}
	newShortChannelId := ""
	if channel.ShortChannelID != nil {
		newShortChannelId = *channel.ShortChannelID
	}
	newLndShortChannelId := uint64(0)
	if channel.LNDShortChannelID != nil {
		newLndShortChannelId = *channel.LNDShortChannelID
	}
	oldShortChannelId := ""
	if existingChannelSettings.ShortChannelId != nil {
		oldShortChannelId = *existingChannelSettings.ShortChannelId
	}
	oldLndShortChannelId := uint64(0)
	if existingChannelSettings.LndShortChannelId != nil {
		oldLndShortChannelId = *existingChannelSettings.LndShortChannelId
	}
	if existingChannelSettings.Status != channel.Status ||
		oldShortChannelId != newShortChannelId ||
		oldLndShortChannelId != newLndShortChannelId {
		err = updateChannelStatusAndLndIds(db, existingChannelId, channel.Status, channel.ShortChannelID,
			channel.LNDShortChannelID)
		if err != nil {
			return Channel{}, errors.Wrapf(err,
				"Updating existing channel with FundingTransactionHash %v, FundingOutputIndex %v",
				channel.FundingTransactionHash, channel.FundingOutputIndex)
		}
	}
	if channel.FundingTransactionHash != nil && channel.FundingOutputIndex != nil && (existingChannelSettings.FundingTransactionHash == nil ||
		*existingChannelSettings.FundingTransactionHash != *channel.FundingTransactionHash ||
		existingChannelSettings.FundingOutputIndex == nil ||
		*existingChannelSettings.FundingOutputIndex != *channel.FundingOutputIndex ||
		existingChannelSettings.FundingBlockHeight == nil && channel.FundingBlockHeight != nil ||
		existingChannelSettings.FundedOn == nil && channel.FundedOn != nil ||
		!existingChannelSettings.HasChannelFlags(core.FundedOn) && channel.HasChannelFlags(core.FundedOn)) {
		err := updateChannelFunding(db, existingChannelId, channel.FundingTransactionHash, channel.FundingOutputIndex,
			channel.FundingBlockHeight, channel.FundedOn, channel.Flags)
		if err != nil {
			return Channel{}, errors.Wrapf(err,
				"Updating channel status and closing transaction hash %v.", existingChannelId)
		}
	}
	if channel.ClosingTransactionHash != nil && (existingChannelSettings.ClosingTransactionHash == nil ||
		*existingChannelSettings.ClosingTransactionHash != *channel.ClosingTransactionHash ||
		existingChannelSettings.ClosingBlockHeight == nil && channel.ClosingBlockHeight != nil ||
		existingChannelSettings.ClosedOn == nil && channel.ClosedOn != nil ||
		existingChannelSettings.ClosingNodeId == nil && channel.ClosingNodeId != nil ||
		!existingChannelSettings.HasChannelFlags(core.ClosedOn) && channel.HasChannelFlags(core.ClosedOn)) {
		err := updateChannelClosing(db, existingChannelId,
			*channel.ClosingTransactionHash, channel.ClosingBlockHeight, channel.ClosedOn, channel.ClosingNodeId,
			channel.Flags)
		if err != nil {
			return Channel{}, errors.Wrapf(err,
				"Updating channel status and closing transaction hash %v.", existingChannelId)
		}
	}
	return channel, nil
}

func createWhenNonExisting(db *sqlx.DB, channel Channel) (Channel, int, error) {
	var err error
	var existingChannelId int
	if channel.FundingTransactionHash == nil || *channel.FundingTransactionHash == "" {
		if channel.ShortChannelID == nil || *channel.ShortChannelID == "" || *channel.ShortChannelID == "0x0x0" {
			return Channel{}, 0,
				errors.Wrap(err, "no ShortChannelId nor Funding Transaction information.")
		} else {
			existingChannelId = cache.GetChannelIdByShortChannelId(channel.ShortChannelID)
			if existingChannelId == 0 {
				existingChannelId, err = getChannelIdByShortChannelId(db, channel.ShortChannelID)
				if err != nil {
					return Channel{}, 0,
						errors.Wrapf(err, "getting channelId by ShortChannelId %v", channel.ShortChannelID)
				}
			}
		}
		if existingChannelId == 0 {
			storedChannel, err := AddChannel(db, channel)
			if err != nil {
				return Channel{}, 0,
					errors.Wrapf(err, "adding channel ShortChannelId %v", *channel.ShortChannelID)
			}
			return storedChannel, 0, nil
		}
	} else {
		existingChannelId = cache.GetChannelIdByFundingTransaction(
			channel.FundingTransactionHash, channel.FundingOutputIndex)
		if existingChannelId == 0 {
			existingChannelId, err = getChannelIdByFundingTransaction(db,
				*channel.FundingTransactionHash, *channel.FundingOutputIndex)
			if err != nil {
				return Channel{}, 0,
					errors.Wrapf(err, "Getting channelId by FundingTransactionHash %v, FundingOutputIndex %v",
						channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
		}
		if existingChannelId == 0 {
			storedChannel, err := AddChannel(db, channel)
			if err != nil {
				return Channel{}, 0,
					errors.Wrapf(err, "Adding channel FundingTransactionHash %v, FundingOutputIndex %v",
						channel.FundingTransactionHash, channel.FundingOutputIndex)
			}
			return storedChannel, 0, nil
		}
	}
	return channel, existingChannelId, nil
}

func updateChannelClosing(db *sqlx.DB, channelId int,
	closingTransactionHash string, closingBlockHeight *uint32, closedOn *time.Time, closingNodeId *int,
	flags core.ChannelFlags) error {
	_, err := db.Exec(`
		UPDATE channel
		SET closing_transaction_hash=$1, updated_on=$2, closing_node_id=$4, closing_block_height=$5, closed_on=$6, flags=$7
		WHERE channel_id=$3 AND
		    (
		        (closing_transaction_hash IS NULL OR closing_transaction_hash != $1) OR
		        closing_node_id IS NULL OR closing_node_id != $4 OR
		        closing_block_height IS NULL OR closing_block_height != $5 OR
		        closed_on IS NULL OR closed_on != $6 OR
		        flags != $7
			)`,
		closingTransactionHash, time.Now().UTC(), channelId, closingNodeId, closingBlockHeight, closedOn, flags)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateChannelFunding(db *sqlx.DB, channelId int,
	fundingTransactionHash *string, fundingOutputIndex *int,
	fundingBlockHeight *uint32, fundedOn *time.Time, flags core.ChannelFlags) error {
	_, err := db.Exec(`
		UPDATE channel
		SET updated_on=$1, funding_transaction_hash=$2, funding_output_index=$3, funding_block_height=$4, funded_on=$5, flags=$6
		WHERE channel_id=$7 AND
		    (
		        funding_transaction_hash IS NULL OR funding_transaction_hash!=$2 OR
		        funding_output_index IS NULL OR funding_output_index!=$3 OR
		        funding_block_height IS NULL OR funding_block_height!=$4 OR
		        funded_on IS NULL OR funded_on!=$5 OR
		        flags!=$6
			)`,
		time.Now().UTC(), fundingTransactionHash, fundingOutputIndex, fundingBlockHeight, fundedOn, flags, channelId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateChannelStatusAndLndIds(db *sqlx.DB, channelId int, status core.ChannelStatus, shortChannelId *string,
	lndShortChannelId *uint64) error {
	if shortChannelId != nil && (*shortChannelId == "" || *shortChannelId == "0x0x0") {
		shortChannelId = nil
	}
	if lndShortChannelId != nil && *lndShortChannelId == 0 {
		lndShortChannelId = nil
	}
	_, err := db.Exec(`
		UPDATE channel
		SET status_id=$2, short_channel_id=$3, lnd_short_channel_id=$4, updated_on=$5
		WHERE channel_id=$1 AND (
		    status_id!=$2 OR
		    short_channel_id IS NULL OR short_channel_id!=$3 OR
		    lnd_short_channel_id IS NULL OR lnd_short_channel_id!=$4)`,
		channelId, status, shortChannelId, lndShortChannelId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

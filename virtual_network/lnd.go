package virtual_network

import (
	"bytes"
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

const defautDelayMS = 2000          // 2s
const defaultMaxDurationMS = 120000 // 120s

type channel struct {
	Active                bool          `json:"active,omitempty"`
	RemotePubkey          string        `json:"remote_pubkey,omitempty"`
	ChannelPoint          string        `json:"channel_point,omitempty"`
	ChanId                string        `json:"chan_id,omitempty"`
	Capacity              string        `json:"capacity,omitempty"`
	LocalBalance          string        `json:"local_balance,omitempty"`
	RemoteBalance         string        `json:"remote_balance,omitempty"`
	CommitFee             string        `json:"commit_fee,omitempty"`
	CommitWeight          string        `json:"commit_weight,omitempty"`
	FeePerKw              string        `json:"fee_per_kw,omitempty"`
	UnsettledBalance      string        `json:"unsettled_balance,omitempty"`
	TotalSatoshisSent     string        `json:"total_satoshis_sent,omitempty"`
	TotalSatoshisReceived string        `json:"total_satoshis_received,omitempty"`
	NumUpdates            string        `json:"num_updates,omitempty"`
	PendingHtlcs          []interface{} `json:"pending_htlcs,omitempty"`
	CsvDelay              int           `json:"csv_delay,omitempty"`
	Private               bool          `json:"private,omitempty"`
	Initiator             bool          `json:"initiator,omitempty"`
	ChanStatusFlags       string        `json:"chan_status_flags,omitempty"`
	LocalChanReserveSat   string        `json:"local_chan_reserve_sat,omitempty"`
	RemoteChanReserveSat  string        `json:"remote_chan_reserve_sat,omitempty"`
	StaticRemoteKey       bool          `json:"static_remote_key,omitempty"`
	CommitmentType        string        `json:"commitment_type,omitempty"`
	Lifetime              string        `json:"lifetime,omitempty"`
	Uptime                string        `json:"uptime,omitempty"`
	CloseAddress          string        `json:"close_address,omitempty"`
	PushAmountSat         string        `json:"push_amount_sat,omitempty"`
	ThawHeight            int           `json:"thaw_height,omitempty"`
	LocalConstraints      struct{}      `json:"local_constraints"`
	RemoteConstraints     struct{}      `json:"remote_constraints"`
}

func GetChannelBalance(ctx context.Context, cli *client.Client, containerId string) (balance string, err error) {
	err = Retry(func() error {
		var channelBalance struct {
			Balance string `json:"balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "channelbalance"}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &channelBalance)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if channelBalance.Balance == "" {
			return errors.New("Payment not received")
		}
		if channelBalance.Balance == "0" {
			return errors.New("Payment not received")
		}
		balance = channelBalance.Balance
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Getting channel balance")
	}
	return balance, nil
}

func PayInvoice(ctx context.Context, cli *client.Client, containerId string, invoice string, waitTime *int) error {
	retryWaitTime := 0
	if waitTime != nil {
		retryWaitTime = *waitTime
	}
	if retryWaitTime == 0 {
		retryWaitTime = defaultMaxDurationMS
	}
	err := Retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "sendpayment", "--force", "--pay_req=" + invoice}
		var stderr bytes.Buffer
		stdout, stderr, err := ExecCommand(ctx, cli, containerId, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if len(stderr.Bytes()) > 0 {
			log.Info().Msg("Standard error not empty, retrying")
			return errors.New("Payment not sent")
		}
		if len(stdout.Bytes()) == 0 {
			log.Info().Msg("Standard out is empty, retrying")
			return errors.New("Payment not sent")
		}
		if strings.Contains(strings.ToLower(stdout.String()), "error") {
			log.Info().Msg("Word error was found in stdout, retrying")
			return errors.New("Payment not sent")
		}
		log.Info().Msg("Pay invoice command complete")
		return nil
	}, defautDelayMS, retryWaitTime)
	if err != nil {
		return errors.Wrap(err, "Sending payment")
	}
	return nil
}

func GenerateInvoice(ctx context.Context, cli *client.Client, containerId string, amount string) (encodedInvoice string, err error) {
	err = Retry(func() error {
		var addInvoice struct {
			EncodedPayReq string `json:"payment_request"`
		}
		cmd := []string{"lncli", "--network=simnet", "addinvoice", "--amt=" + amount}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &addInvoice)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if addInvoice.EncodedPayReq == "" {
			return errors.New("Invoice not generated")
		}
		encodedInvoice = addInvoice.EncodedPayReq
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating invoice")
	}
	return encodedInvoice, nil
}

func CheckPeerExists(ctx context.Context, cli *client.Client, containerId string, remotePubkey string) (bool, error) {
	var listPeers struct {
		Peers []struct {
			Pubkey   string `json:"pub_key"`
			SyncType string `json:"sync_type"`
		} `json:"peers"`
	}
	err := Retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "listpeers"}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &listPeers)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", containerId)
		}
		for _, peer := range listPeers.Peers {
			if peer.Pubkey == remotePubkey {
				// peer found, return from retry
				if peer.SyncType != "ACTIVE_SYNC" {
					return noRetryError{}
				}
				return nil
			}
		}
		return errors.New("peer not found")
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return false, errors.Wrap(err, "Checking nodes are peers")
	}
	return true, nil
}

func ConnectPeer(ctx context.Context, cli *client.Client, containerId string, remotePubkey string, remoteIPAddress string) error {
	err := Retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "connect", remotePubkey + "@" + remoteIPAddress}
		var stderr bytes.Buffer
		_, stderr, err := ExecCommand(ctx, cli, containerId, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if len(stderr.Bytes()) > 0 {
			if !strings.Contains(string(stderr.String()), "already connected to peer") {
				return errors.New("Peer didn't connect")
			}
		}

		// immediately check if the peer is connected as sometimes it seems to succeed and didn't
		peerConnected, err := CheckPeerExists(ctx, cli, containerId, remotePubkey)
		if err != nil || !peerConnected {
			return errors.New("Peer didn't connect")
		}

		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return errors.Wrap(err, "Connecting peers")
	}
	return nil
}

func CloseChannel(ctx context.Context, cli *client.Client, containerId string,
	channelPoint string) (closeTxId string, err error) {
	err = Retry(func() error {
		var closeChannel struct {
			ClosingTxId string `json:"closing_txid"`
		}
		fundingTxId := channelPoint[:strings.IndexByte(channelPoint, ':')]
		outputIndex := channelPoint[strings.IndexByte(channelPoint, ':')+1:]
		cmd := []string{"lncli", "--network=simnet", "closechannel", "--funding_txid=" + fundingTxId, "--output_index=" + outputIndex}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &closeChannel)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", containerId)
		}
		if closeChannel.ClosingTxId == "" {
			return errors.New("Channel not closed")
		}
		closeTxId = closeChannel.ClosingTxId
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Closing channel")
	}
	return closeTxId, nil
}

func CreateChannel(ctx context.Context, cli *client.Client, containerId string, remotePubkey string,
	amount string, btcdId string) (channelPoint string, err error) {

	var fundingTxId string
	err = Retry(func() error {
		var openChannel struct {
			FundingTxId string `json:"funding_txid"`
		}
		cmd := []string{"lncli", "--network=simnet", "openchannel", "--node_key=" + remotePubkey, "--local_amt=" + amount}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &openChannel)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if openChannel.FundingTxId == "" {
			return errors.New("Channel not created")
		}
		fundingTxId = openChannel.FundingTxId
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating channel")
	}
	log.Printf("Funding transaction ID: %s\n", fundingTxId)

	log.Info().Msg("Include funding transaction in block thereby opening the channel")

	err = MineBlocks(ctx, cli, btcdId, 30)
	if err != nil {
		return "", errors.Wrap(err, "Mining blocks")
	}

	log.Info().Msg("Blocks mined")
	log.Info().Msg("Checking channel is open")

	err = Retry(func() error {
		var listChannels struct {
			Channels []struct {
				ChannelPoint string `json:"channel_point"`
			} `json:"channels"`
		}
		cmd := []string{"lncli", "--network=simnet", "listchannels"}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &listChannels)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if len(listChannels.Channels) == 0 {
			return errors.New("Channel not open")
		}
		if listChannels.Channels[0].ChannelPoint == "" {
			return errors.New("Channel not open")
		}
		channelPoint = listChannels.Channels[0].ChannelPoint
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating channel")
	}
	log.Printf("Channel point: %s\n", channelPoint)

	return channelPoint, nil

}

func GetNewAddress(ctx context.Context, cli *client.Client, instanceId string) (addr string, err error) {

	err = Retry(func() error {
		var address struct {
			Address string `json:"address"`
		}
		cmd := []string{"lncli", "--network=simnet", "newaddress", "np2wkh"}
		err := ExecJSONReturningCommand(ctx, cli, instanceId, cmd, &address)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", instanceId)
		}
		if address.Address == "" {
			return errors.New("Not a valid address")
		}
		addr = address.Address
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return addr, errors.Wrap(err, "Getting new address")
	}

	return addr, nil
}

func GetPubKey(ctx context.Context, cli *client.Client, containerId string) (pubkey string, err error) {
	var getInfo struct {
		IdentityPubkey string `json:"identity_pubkey"`
	}
	err = Retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "getinfo"}
		err = ExecJSONReturningCommand(ctx, cli, containerId, cmd, &getInfo)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if getInfo.IdentityPubkey == "" {
			return errors.New("Invalid Pubkey")
		}
		pubkey = getInfo.IdentityPubkey
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Calling getinfo")
	}
	return pubkey, nil
}

func GetOnchainBalance(ctx context.Context, cli *client.Client, containerId string) (balance string, err error) {
	err = Retry(func() error {
		var walletBalance struct {
			ConfirmedBalance string `json:"confirmed_balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "walletbalance"}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &walletBalance)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if walletBalance.ConfirmedBalance == "" {
			return errors.New("Balance not confirmed")
		}
		if walletBalance.ConfirmedBalance == "0" {
			return errors.New("Balance not confirmed")
		}
		balance = walletBalance.ConfirmedBalance
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Getting balance")
	}
	return balance, nil

}

func ListNodeChannels(ctx context.Context, cli *client.Client, containerId string, pubkey string) (channels []string, err error) {
	err = Retry(func() error {
		var channelsList struct {
			Channels []channel `json:"channels"`
		}
		cmd := []string{"lncli", "--network=simnet", "listchannels", "--peer=" + pubkey}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &channelsList)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}
		if len(channelsList.Channels) == 0 {
			return nil
		}

		for _, openedChans := range channelsList.Channels {
			channels = append(channels, openedChans.ChannelPoint)
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return nil, errors.Wrap(err, "Channels list")
	}
	return channels, nil
}

func AddressSendCoins(ctx context.Context, cli *client.Client, containerId string, address string, amt string) (txId string, err error) {
	err = Retry(func() error {
		var transID struct {
			TxId string `json:"txId"`
		}
		cmd := []string{"lncli", "--network=simnet", "sendcoins", address, amt}
		err := ExecJSONReturningCommand(ctx, cli, containerId, cmd, &transID)
		if err != nil {
			log.Info().Msgf("Error %v", err)
			return errors.Wrapf(err, "Running exec command on %s", containerId)
		}

		if transID.TxId == "" {
			return errors.New("Invalid Txid")
		}

		txId = transID.TxId

		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Channels list")
	}
	return txId, nil
}

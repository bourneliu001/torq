package move_funds

import (
	"github.com/gin-gonic/gin"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
)

func moveFundsOffChainHandler(c *gin.Context) {
	var request struct {
		OutgoingNodeId int   `json:"outgoingNodeId"`
		AmountMsat     int64 `json:"amountMsat"`
		IncomingNodeId int   `json:"incomingNodeId"`
		ChannelId      int   `json:"channelId"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	memo := "Moving funds between nodes"
	invoiceResponse, err := lightning.NewInvoice(lightning_helpers.NewInvoiceRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: request.IncomingNodeId},
		Memo:                 &memo,
		RPreImage:            nil,
		ValueMsat:            &request.AmountMsat,
	})

	response, err := lightning.MoveFundsOffChain(lightning_helpers.MoveFundsOffChainRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: request.OutgoingNodeId},
		ChannelId:            request.ChannelId,
		OutgoingNodeId:       request.OutgoingNodeId,
		IncomingNodeId:       request.IncomingNodeId,
		AmountMsat:           request.AmountMsat,
		RHash:                invoiceResponse.RHash,
		PaymentAddress:       invoiceResponse.PaymentAddress,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func moveOnChainFundsHandler(c *gin.Context) {
	var request struct {
		OutgoingNodeId   int                           `json:"outgoingNodeId"`
		IncomingNodeId   int                           `json:"incomingNodeId"`
		AmountMsat       int64                         `json:"amountMsat"`
		TargetConf       *int32                        `json:"targetConf"`
		SatPerVbyte      *uint64                       `json:"satPerVbyte"`
		SpendUnconfirmed *bool                         `json:"spendUnconfirmed"`
		MinConf          *int32                        `json:"minConf"`
		AddressType      lightning_helpers.AddressType `json:"addressType"`
		SendAll          *bool                         `json:"sendAll"`
	}

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// If both targetConf and satPerVbyte are set then return an error
	if request.TargetConf != nil && request.SatPerVbyte != nil {
		// return bad request with field error message using the helper
		retError := server_errors.ServerError{}
		retError.AddFieldError("targetConf", "Cannot set both Target Confirmations and Sat Per Vbyte")
		retError.AddFieldError("satPerVbyte", "Cannot set both Target Confirmations and Sat Per Vbyte")
		server_errors.SendBadRequestFieldError(c, &retError)
		return
	}

	// Get the address to send to
	address, err := lightning.NewAddress(lightning_helpers.NewAddressRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: request.IncomingNodeId},
		Type:                 lightning_helpers.P2WKH,
	})

	label := "Moving funds between nodes"
	// Send the funds on chain
	response, err := lightning.OnChainPayment(lightning_helpers.OnChainPaymentRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{NodeId: request.OutgoingNodeId},
		Address:              address,
		AmountSat:            request.AmountMsat / 1000,
		TargetConf:           request.TargetConf,
		SatPerVbyte:          request.SatPerVbyte,
		Label:                &label,
		MinConfs:             request.MinConf,
		SpendUnconfirmed:     request.SpendUnconfirmed,
		SendAll:              request.SendAll,
	})

	c.JSON(http.StatusOK, response)
}

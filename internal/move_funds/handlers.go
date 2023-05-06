package move_funds

import (
	"github.com/gin-gonic/gin"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/lightning_helpers"
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

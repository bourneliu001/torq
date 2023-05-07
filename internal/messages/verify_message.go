package messages

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/internal/lightning"
)

func verifyMessage(ctx context.Context, req VerifyMessageRequest) (VerifyMessageResponse, error) {
	if req.NodeId == 0 {
		return VerifyMessageResponse{}, errors.New("Node Id missing")
	}
	publicKey, valid, err := lightning.SignatureVerification(ctx, req.NodeId, req.Message, req.Signature)
	if err != nil {
		return VerifyMessageResponse{}, errors.Wrapf(err, "Signature Verification (nodeId: %v)", req.NodeId)
	}
	return VerifyMessageResponse{
		Valid:  valid,
		PubKey: publicKey,
	}, nil
}

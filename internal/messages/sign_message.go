package messages

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/internal/lightning"
)

func signMessage(ctx context.Context, req SignMessageRequest) (SignMessageResponse, error) {
	if req.NodeId == 0 {
		return SignMessageResponse{}, errors.New("Node Id missing")
	}
	signature, err := lightning.SignMessage(ctx, req.NodeId, req.Message, req.SingleHash)
	if err != nil {
		return SignMessageResponse{}, errors.Wrapf(err, "Signing message (nodeId: %v)", req.NodeId)
	}
	return SignMessageResponse{
		Signature: signature,
	}, nil
}

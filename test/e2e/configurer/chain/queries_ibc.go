package chain

import (
	"fmt"
	"net/url"

	"github.com/babylonchain/babylon/test/e2e/util"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func (n *NodeConfig) QueryIBCChannels() (*channeltypes.QueryChannelsResponse, error) {
	path := "/ibc/core/channel/v1/channels"
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp channeltypes.QueryChannelsResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (n *NodeConfig) QueryNextSequenceReceive(channelID, portID string) (*channeltypes.QueryNextSequenceReceiveResponse, error) {
	path := fmt.Sprintf("/ibc/core/channel/v1/channels/%s/ports/%s/next_sequence", channelID, portID)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp channeltypes.QueryNextSequenceReceiveResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (n *NodeConfig) QueryNextSequenceSend(channelID, portID string) (*channeltypes.QueryNextSequenceSendResponse, error) {
	path := fmt.Sprintf("/ibc/core/channel/v1/channels/%s/ports/%s/next_sequence_send", channelID, portID)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp channeltypes.QueryNextSequenceSendResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (n *NodeConfig) QueryPacketAcknowledgement(channelID string, portID string, sequence uint64) (*channeltypes.QueryPacketAcknowledgementResponse, error) {
	path := fmt.Sprintf("/ibc/core/channel/v1/channels/%s/ports/%s/packet_acks/%d", channelID, portID, sequence)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp channeltypes.QueryPacketAcknowledgementResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

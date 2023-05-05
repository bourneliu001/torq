export type MoveFundsOffChainRequest = {
  outgoingNodeId: string;
  incomingNodeId: string;
  amountMsat: number;
  channelId: string;
};
export type MoveFundsOffChainResponse = {
  status: string;
};

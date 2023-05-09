export type MoveFundsOffChainRequest = {
  outgoingNodeId: string;
  incomingNodeId: string;
  amountMsat: number;
  channelId: string;
};

export type MoveFundsOffChainResponse = {
  status: string;
};

export enum AddressType {
  P2WPKH = 1,
  P2WKH = 2,
  P2TR = 4,
}

export type MoveOnChainFundsRequest = {
  outgoingNodeId: string;
  incomingNodeId: string;
  amountMsat: number;
  targetConf?: number;
  satPerVbyte?: number;
  spendUnconfirmed?: boolean;
  minConf?: number;
  addressType?: AddressType;
  sendAll?: boolean;
};

export type MoveOnChainFundsResponse = {
  status: string;
  txId: string;
};

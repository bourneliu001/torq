export type MoveFundsOffChainRequest = {
  outgoingNodeId: number;
  incomingNodeId: number;
  amountMsat: number;
  channelId: number;
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
  outgoingNodeId: number;
  incomingNodeId: number;
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

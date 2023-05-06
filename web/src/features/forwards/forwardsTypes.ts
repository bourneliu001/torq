import { Tag } from "pages/tags/tagsTypes";

export type Forward = {
  alias: string;
  channelId: number;
  channelPoint: string;
  pubKey: string;
  shortChannelId: string;
  lndShortChannelId: string;
  fundingOutputIndex: number;
  fundingTransactionHash: string;
  color: string;
  open: boolean;
  capacity: number;
  amountOut: number;
  amountIn: number;
  amountTotal: number;
  revenueOut: number;
  revenueIn: number;
  revenueTotal: number;
  countOut: number;
  countIn: number;
  countTotal: number;
  turnoverOut: number;
  turnoverIn: number;
  turnoverTotal: number;
  localNodeIds: Array<number>;
  tags: Array<Tag>; // this is aggregate of channelTags and peerTags for the frontend filters
  channelTags: Array<Tag>;
  peerTags: Array<Tag>;
  peerNodeId: number;
  torqNodeId: number;
  torqNodeName: string;
};

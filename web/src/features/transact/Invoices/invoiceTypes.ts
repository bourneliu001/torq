import { Pagination } from "types/api";

export type InvoicesResponse = {
  data: Array<Invoice>;
  pagination: Pagination;
};

export type Invoice = {
  creationDate: string;
  settleDate: string;
  addIndex: number;
  settleIndex: number;
  paymentRequest: string;
  destinationPubKey: string;
  rHash: string;
  rPreimage: string;
  memo: string;
  value: number;
  amtPaid: number;
  invoiceState: string;
  isRebalance: boolean;
  isKeysend: boolean;
  isAmp: boolean;
  paymentAddr: string;
  fallbackAddr: string;
  updatedOn: string;
  expiry: number;
  cltvExpiry: number;
  private: boolean;
};

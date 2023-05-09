export interface settings {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
  torqUuid: string;
  mixpanelOptOut: boolean;
  slackOAuthToken: string;
  slackBotAppToken: string;
  telegramHighPriorityCredentials: string;
  telegramLowPriorityCredentials: string;
}
export interface updateSettingsRequest {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
  mixpanelOptOut: boolean;
}

export interface timeZone {
  name: string;
}

export interface nodeConfiguration {
  nodeId: number;
  name?: string;
  implementation: number;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsFile?: File | null;
  macaroonFileName?: string;
  macaroonFile?: File | null;
  caCertificateFileName?: string;
  caCertificateFile?: File | null;
  certificateFileName?: string;
  certificateFile?: File | null;
  keyFileName?: string;
  keyFile?: File | null;
  createdOn?: Date;
  updatedOn?: Date;
  status: number;
  pingSystem: number;
  customSettings: number;
  nodeStartDate?: Date;
  nodeCssColour?: string;
}

export interface stringMap<T> {
  [key: string]: T;
}

export interface services {
  version: string;
  bitcoinNetworks: string[];
  mainService: torqService;
  torqServices: torqService[];
  lndServices: lndService[];
}

export interface torqService {
  status: number;
  statusString: string;
  bootTime: string | null;
  nodeId: number;
  type: number;
  typeString: string;
}

export interface lndService {
  status: number;
  statusString: string;
  bootTime: string | null;
  nodeId: number;
  type: number;
  typeString: string;
}

export interface lightningRequest {
  nodeId: number;
}

export interface nodeWalletBalances {
  request: lightningRequest;
  totalBalance: number;
  confirmedBalance: number;
  unconfirmedBalance: number;
  lockedBalance: number;
  reservedBalanceAnchorChan: number;
}

export interface nodeAddress {
  addr: string;
  network: string;
}
export interface nodeInformation {
  nodeId: number;
  publicKey: string;
  status: nodeStatus;
  addresses: nodeAddress[];
  alias: string;
}

export enum nodeStatus {
  inactive = 0,
  active = 1,
}

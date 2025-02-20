export const ROOT = "/";

export const LOGIN = "login";
export const SERVICES = "services";
export const COOKIELOGIN = "cookie-login";
export const LOGOUT = "logout";

export const ANALYSE = "analyse";
export const MANAGE = "manage";
export const FORWARDS = "forwards";
export const TAGS = "tags";
export const FORWARDS_CUSTOM_VIEW = `${FORWARDS}/:viewId`;
export const CHANNELS = "channels";
export const OPEN_CHANNELS = "open-channels";
export const PENDING_CHANNELS = "pending-channels";
export const CLOSED_CHANNELS = "closed-channels";
export const INSPECT_CHANNEL = "/analyse/inspect/:chanId";
export const FORWARDS_SUMMARY = "forwards-summary";

export const TRANSACTIONS = "transactions";
export const PAYMENTS = "payments";
export const INVOICES = "invoices";
export const ONCHAIN = "onchain";
export const PEERS = "peers";
export const ALL = "all";

export const SETTINGS = "/settings";

// modals
export const NEW_INVOICE = "/new-invoice";
export const NEW_PAYMENT = "/new-payment";
export const NEW_ADDRESS = "/new-address";
export const UPDATE_CHANNEL = "/update-channel";
export const OPEN_CHANNEL = "/open-channel";
export const CLOSE_CHANNEL = "/close-channel";
export const TAG = "/create-tag";
export const UPDATE_TAG = "/update-tag/:tagId";
export const TAG_CHANNEL = "/tag-channel/:channelId";
export const TAG_NODE = "/tag-node/:nodeId";
export const MESSAGE_VERIFICATION = "/message-verification";
export const CONNECT_PEER = "/connect-peer";
export const UPDATE_PEER = "/edit-peer";

// Automation
export const WORKFLOWS = "workflows";
export const WORKFLOW = "workflows/:workflowId/versions/:version";

export const MOVE_FUNDS = "move-funds";

import { useGroupBy } from "./groupBy";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const testData: Array<any> = [
  {
    alias: "Some Node",
    channelId: 32,
    channelPoint: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e288:0",
    pubKey: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2052:0",
    lndShortChannelId: "769235927112613888",
    color: "#68f442",
    open: 1,
    capacity: 10000000,
    countTotal: 20,
    htlcFailAllIn: 10,
    htlcFailAllOut: 150,
    torqNodeId: 1,
  },
  {
    alias: "Another Node",
    channelId: 40,
    channelPoint: "f1c17e33b03bb3722eee187d5cceaaeab7b1e3e72d6efcbebf263747122a770f:0",
    pubKey: "033f405aae705d96d4338efb236645a61c9b0a2303e3185211ed3b02c0803a4a2a",
    shortChannelId: "707781:900:1",
    lndShortChannelId: "778213439477907457",
    color: "#68f4a2",
    open: 1,
    capacity: 2000000,
    countTotal: 10,
    htlcFailAllIn: 5,
    htlcFailAllOut: 50,
    torqNodeId: 1,
  },
  {
    alias: "Some Node",
    channelId: 33,
    channelPoint: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e289:0",
    pubKey: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2053:0",
    lndShortChannelId: "769235927112613889",
    color: "#68f442",
    open: 0,
    capacity: 5000000,
    countTotal: 10,
    htlcFailAllIn: 10,
    htlcFailAllOut: 100,
    torqNodeId: 1,
  },
  {
    alias: "Some Node",
    channelId: 34,
    channelPoint: "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e290:0",
    pubKey: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
    shortChannelId: "699616:2054:0",
    lndShortChannelId: "769235927112613890",
    color: "#68f442",
    open: 0,
    capacity: 5000000,
    countTotal: 0,
    htlcFailAllIn: 0,
    htlcFailAllOut: 0,
    torqNodeId: 1,
  },
];

test("Unknown by param returns exactly what was input", () => {
  const result = useGroupBy(testData, undefined);

  expect(result).toStrictEqual(testData);
});

test("grouping by channels returns exactly what was input", () => {
  const result = useGroupBy(testData, "channel");

  expect(result).toStrictEqual(testData);
});

test("grouping by peers returns correctly grouped channels", () => {
  const result = useGroupBy(testData, "peer");

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const expected: Array<any> = [
    {
      alias: "Some Node",
      channelId: [32, 33, 34],
      channelPoint: [
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e288:0",
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e289:0",
        "448e5a0842cc46ccd16eba686a9ff312ac5f3d27ba67e43b25c91e008a92e290:0",
      ],
      pubKey: "02ab38160e8f24f9cce8d851a091ec927748e78507adc7f7ee01664728a981a597",
      shortChannelId: ["699616:2052:0", "699616:2053:0", "699616:2054:0"],
      lndShortChannelId: ["769235927112613888", "769235927112613889", "769235927112613890"],
      color: "#68f442",
      open: 1,
      capacity: 20000000,
      countTotal: 30,
      htlcFailAllIn: 20,
      htlcFailAllOut: 250,
      torqNodeId: 1,
    },
    {
      alias: "Another Node",
      channelId: 40,
      channelPoint: "f1c17e33b03bb3722eee187d5cceaaeab7b1e3e72d6efcbebf263747122a770f:0",
      pubKey: "033f405aae705d96d4338efb236645a61c9b0a2303e3185211ed3b02c0803a4a2a",
      shortChannelId: "707781:900:1",
      lndShortChannelId: "778213439477907457",
      color: "#68f4a2",
      open: 1,
      capacity: 2000000,
      countTotal: 10,
      htlcFailAllIn: 5,
      htlcFailAllOut: 50,
      torqNodeId: 1,
    },
  ];

  expect(result).toStrictEqual(expected);
});

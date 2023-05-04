import { GroupByOptions } from "features/viewManagement/types";
import { AnyObject } from "utils/types";

const nonSummableFields: Array<string> = [
  "alias",
  "pubKey",
  "color",
  "secondNodeId",
  "firstNodeId",
  "torqNodeId",
  "torqNodeName",
  "peerTags",
];
const arrayAggKeys: Array<string> = [
  "channelId",
  "channelPoint",
  "shortChannelId",
  "lndShortChannelId",
  "channelTags",
  "tags",
];

export function useGroupBy<T extends AnyObject>(data: Array<T>, by: GroupByOptions | undefined): Array<T> {
  if (by !== "peer") {
    return data;
  }

  const summedPubKey: typeof data = [];

  for (const chan of data) {
    const pubKey = String(chan["pubKey" as keyof T]);
    const torqNodeId = String(chan["torqNodeId" as keyof T]);

    console.log(torqNodeId);
    const summedChan = summedPubKey.find(
      (sc) => sc["pubKey" as keyof typeof sc] == pubKey && sc["torqNodeId" as keyof typeof sc] == torqNodeId
    );
    if (!summedChan) {
      summedPubKey.push(chan);
      console.log("pushing");
      continue;
    }

    for (const key of Object.keys(chan)) {
      const value = chan[key as keyof typeof chan];

      if (nonSummableFields.includes(key)) {
        continue;
      }

      console.log(key);
      // Values fround in arrayAggKeys should be converted to an array of values
      if (arrayAggKeys.includes(key)) {
        console.log("it's considered an array");
        let valueArr = [];
        if (Array.isArray(value)) {
          valueArr = [...value];
        } else {
          valueArr = [value];
        }

        // If the previous result is not already an Array, create a new one
        if (!Array.isArray(summedChan[key as keyof typeof summedChan])) {
          (summedChan as { [key: string]: unknown })[key] = [summedChan[key as keyof typeof summedChan], ...valueArr];
          continue;
        }

        (summedChan as { [key: string]: unknown })[key] = [
          ...(summedChan[key as keyof typeof summedChan] as []),
          ...valueArr,
        ];
        continue;
      }

      (summedChan as { [key: string]: unknown })[key] =
        (summedChan[key as keyof typeof summedChan] as number) + (value as number);
    }
  }

  return summedPubKey;
}

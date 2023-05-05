import { torqApi } from "apiSlice";
import { MoveFundsOffChainRequest, MoveFundsOffChainResponse } from "./moveFundsTypes";

// Define a service using a base URL and expected endpoints
export const moveFundsApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    moveFundsOffChain: builder.mutation<MoveFundsOffChainResponse, MoveFundsOffChainRequest>({
      query: (body) => ({
        url: "/move-funds/off-chain",
        method: "POST",
        body,
      }),
      invalidatesTags: ["channels"],
    }),
  }),
});

export const { useMoveFundsOffChainMutation } = moveFundsApi;

import { torqApi } from "apiSlice";
import {
  MoveFundsOffChainRequest,
  MoveFundsOffChainResponse,
  MoveOnChainFundsRequest,
  MoveOnChainFundsResponse,
} from "./moveFundsTypes";

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
    moveOnChainFunds: builder.mutation<MoveOnChainFundsResponse, MoveOnChainFundsRequest>({
      query: (body) => ({
        url: "/move-funds/on-chain",
        method: "POST",
        body,
      }),
    }),
  }),
});

export const { useMoveFundsOffChainMutation, useMoveOnChainFundsMutation } = moveFundsApi;

import { torqApi } from "apiSlice";
import {
  ConnectPeerRequest,
  ConnectPeerResponse,
  DisconnectPeerRequest,
  DisconnectPeerResponse,
  UpdatePeerRequest,
  UpdatePeerResponse,
} from "./peersTypes";

export const peersApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    connectPeer: builder.mutation<ConnectPeerResponse, ConnectPeerRequest>({
      query: (body) => ({
        url: `lightning/peers/connect`,
        method: "POST",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    disconnectPeer: builder.mutation<DisconnectPeerResponse, DisconnectPeerRequest>({
      query: (body) => ({
        url: `lightning/peers/disconnect`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    reconnectPeer: builder.mutation<DisconnectPeerResponse, DisconnectPeerRequest>({
      query: (body) => ({
        url: `lightning/peers/reconnect`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
    updatePeer: builder.mutation<UpdatePeerResponse, UpdatePeerRequest>({
      query: (body) => ({
        url: `lightning/peers/update`,
        method: "PATCH",
        body: body,
      }),
      invalidatesTags: ["peers"],
    }),
  }),
});

export const { useConnectPeerMutation, useDisconnectPeerMutation, useReconnectPeerMutation, useUpdatePeerMutation } =
  peersApi;

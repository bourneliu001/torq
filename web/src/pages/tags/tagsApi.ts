import { torqApi } from "apiSlice";
import { ChannelNode, TagChannelRequest, TagNodeRequest, TagResponse } from "./tagsTypes";

// Define a service using a base URL and expected endpoints
export const tagsApi = torqApi.injectEndpoints({
  endpoints: (builder) => ({
    getTags: builder.query<Array<TagResponse>, void>({
      query: () => `tags`,
      providesTags: ["tags"],
    }),
    getTag: builder.query<TagResponse, number>({
      query: (tagId) => `tags/${tagId}`,
      providesTags: ["tag"],
    }),
    addTag: builder.mutation<TagResponse, TagResponse>({
      query: (tag) => ({
        url: `tags`,
        method: "POST",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    setTag: builder.mutation<TagResponse, TagResponse>({
      query: (tag) => ({
        url: `tags`,
        method: "PUT",
        body: tag,
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    deleteTag: builder.mutation<number, number>({
      query: (tagId) => ({
        url: `tags/${tagId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["tags", "tag"],
    }),
    tagChannel: builder.mutation<void, TagChannelRequest>({
      query: (tagChannel) => ({
        url: `tags/tag`,
        method: "POST",
        body: tagChannel,
      }),
      invalidatesTags: [
        "tags",
        "tag",
        "channels",
        "channelsClosed",
        "channelsPending",
        "tagsForChannel",
        "channelHistory",
        "forwards",
      ],
    }),
    tagNode: builder.mutation<void, TagNodeRequest>({
      query: (tagNode) => ({
        url: `tags/tag`,
        method: "POST",
        body: tagNode,
      }),
      invalidatesTags: [
        "tags",
        "tag",
        "tagsForNode",
        "channels",
        "channelsClosed",
        "channelsPending",
        "channelHistory",
        "forwards",
        "peers",
      ],
    }),
    untagNode: builder.mutation<void, TagNodeRequest>({
      query: (tagNode) => ({
        url: `tags/untag`,
        method: "POST",
        body: tagNode,
      }),
      invalidatesTags: [
        "tags",
        "tag",
        "tagsForNode",
        "channels",
        "channelsClosed",
        "channelsPending",
        "channelHistory",
        "forwards",
        "peers",
      ],
    }),
    untagChannel: builder.mutation<void, TagChannelRequest>({
      query: (tagChannel) => ({
        url: `tags/untag`,
        method: "POST",
        body: tagChannel,
      }),
      invalidatesTags: [
        "tags",
        "tag",
        "channels",
        "channelsClosed",
        "channelsPending",
        "tagsForChannel",
        "channelHistory",
        "forwards",
      ],
    }),
    getChannelTags: builder.query<Array<TagResponse>, number>({
      query: (channelId) => `tags/channel/${channelId}`,
      providesTags: ["tagsForChannel"],
    }),
    getNodeTags: builder.query<Array<TagResponse>, number>({
      query: (nodeId) => `tags/node/${nodeId}`,
      providesTags: ["tagsForNode"],
    }),
    getNodesChannels: builder.query<ChannelNode, void>({
      query: () => ({
        url: `channels/nodes`,
        method: "GET",
      }),
    }),
  }),
});
// Export hooks for usage in functional components, which are
// auto-generated based on the defined endpoints
export const {
  useGetTagsQuery,
  useGetTagQuery,
  useTagChannelMutation,
  useTagNodeMutation,
  useUntagNodeMutation,
  useUntagChannelMutation,
  useAddTagMutation,
  useDeleteTagMutation,
  useSetTagMutation,
  useGetChannelTagsQuery,
  useGetNodeTagsQuery,
  useGetNodesChannelsQuery,
} = tagsApi;

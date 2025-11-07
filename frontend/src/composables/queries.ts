import { apolloClient } from "@/apollo.ts";
import type { GqlImage } from "@/composables/images.ts";
import type { ApolloError } from "@apollo/client";
import { provideApolloClient, useQuery } from "@vue/apollo-composable";
import { createEventHook, type EventHookOn } from "@vueuse/core";
import gql from "graphql-tag";
import { type Reactive, ref, type Ref, toRaw, watch, watchEffect } from "vue";

export const ALL_IMAGE_SUMMARIES = gql`
  {
    getAllImageSummaries
  }
`;

export const GET_DYNAMIC_DATA = gql`
  query getDynamicData($group: String!, $type: String!) {
    getDynamicData(group: $group, type: $type) {
      fileSelectors
      expressions
    }
  }
`;

const GET_IMAGE = gql`
  query getImage($bucket: String!, $name: String!) {
    getImage(bucket: $bucket, name: $name) {
      imageSummary {
        bucket
        key
        name
        group
        type
        geonames {
          objects
          cachedObject {
            lastModified
            cacheKey
          }
        }
        productInfo {
          title
          subtitle
          entries
          summary
        }
        cachedObject {
          lastModified
          cacheKey
        }
        size {
          width
          height
        }
      }
      localization {
        corner
        cachedObject {
          lastModified
          cacheKey
        }
      }
      cachedFileLinks
      signedURLs
      targetFiles
    }
  }
`;

export type ImageQueryVariables = {
  bucket: string;
  name: string;
};

export type ImageQueryResult = {
  loading: Ref<boolean>;
  error: Ref<ApolloError | null>;
  data: Ref<GqlImage | null>;
  onResult: EventHookOn<GqlImage | null>;
};

export function useImageQuery(variables: Reactive<ImageQueryVariables>): ImageQueryResult {
  const loadingRet = ref(true);
  const errorRet = ref<ApolloError | null>(null);
  const dataRet = ref<GqlImage | null>(null);
  const resultEvt = createEventHook<GqlImage | null>();

  const runQuery = () => {
    loadingRet.value = true;
    errorRet.value = null; // reset ?
    dataRet.value = null;

    const { result, loading, error } = provideApolloClient(apolloClient)(() =>
      useQuery(GET_IMAGE, toRaw(variables), () => ({
        fetchPolicy: "network-only",
      }))
    );

    watch(loading, (newLoading) => (loadingRet.value = newLoading));
    watch(error, (newError) => (errorRet.value = newError));
    watch(result, (newResult) => {
      dataRet.value = newResult;
      resultEvt.trigger(newResult);
    });
  };

  watchEffect(() => variables.bucket && runQuery());

  return {
    loading: loadingRet,
    error: errorRet,
    data: dataRet,
    onResult: resultEvt.on,
  };
}

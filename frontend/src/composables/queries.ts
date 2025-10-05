import gql from "graphql-tag";

export const ALL_IMAGE_SUMMARIES = gql`
  {
    getAllImageSummaries
  }
`;

export const GET_IMAGE_SUMMARY = gql`
  query getImage($bucket: String!, $name: String!) {
    getImage(bucket: $bucket, name: $name) {
      imageSummary {
        bucket
        key
        name
        group
        type
        features {
          class
          count
          objects
          cachedObject {
            lastModified
            cacheKey
          }
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
      geonames {
        objects
        cachedObject {
          lastModified
          cacheKey
        }
      }
      localization {
        corner
        cachedObject {
          lastModified
          cacheKey
        }
      }
      additionalFiles
      targetFiles
      fullProductFiles
    }
  }
`;

import gql from "graphql-tag";
import type { DocumentNode } from "graphql/language";

export const ALL_IMAGE_SUMMARIES = gql`
  {
    getAllImageSummaries
  }
`;

export const getImage = (bucket: string, name: string): DocumentNode => {
  return gql(`{
  getImage(
    bucket: "${bucket}"
    name: "${name}"
  ) {
    imageSummary {
      bucket
      name
      group
      type
      cachedObject {
        lastModified
        cacheKey
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
    features {
      class
      objects
      class
      cachedObject {
        lastModified
        cacheKey
      }
    }
    additionalFiles
    targetFiles
    fullProductFiles
  }
}`);
};

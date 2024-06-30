import gql from "graphql-tag";
import type { DocumentNode } from "graphql/language";

export const ALL_IMAGE_SUMMARIES = gql`
  {
    getAllImageSummaries
  }
`;

export const getImageQuery = (bucket: string, name: string): DocumentNode => {
  return gql(`{
  getImage(
    bucket: "${bucket}"
    name: "${name}"
  ) {
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
}`);
};

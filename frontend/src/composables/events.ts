type EventType = "ObjectCreated" | "ObjectRemoved";

type ObjectType =
  | "preview"
  | "geonames"
  | "localization"
  | "additional"
  | "features"
  | "target"
  | "full_product";

export interface EventData {
  eventType: EventType;
  objectType: ObjectType;
  imageBucket: string;
  imageKey: string;
  objectTime: Date;
  object: never | undefined; // @ts-ingore
  error: string | undefined;
}

export function parseEventData(rawData: string): EventData {
  return JSON.parse(rawData, (key, value) => {
    switch (key) {
      case "objectTime":
        return new Date(value);
      default:
        return value;
    }
  });
}

type EventType = "ObjectCreated" | "ObjectRemoved";

type ObjectType = "preview" | "target" | "dynamic_input";

export interface EventData {
  eventType: EventType;
  objectType: ObjectType;
  imageBucket: string;
  imageKey: string;
  objectTime: Date;
  object: never | undefined;
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

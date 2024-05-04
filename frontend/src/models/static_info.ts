export interface ImageType {
  name: string;
  displayName: string;
}

export interface ImageGroup {
  name: string;
  bucket: string;
  types: ImageType[];
}

export interface StaticInfo {
  softwareVersion: string;
  windowTitle: string;
  applicationTitle: string;
  faviconBase64: string;
  logoBase64: string;
  scaleInitialPercentage: number;
  maxImagesDisplayCount: number;
  tileServerURL: string;
  imageGroups: ImageGroup[];
}

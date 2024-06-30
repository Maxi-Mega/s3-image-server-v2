export class ImageType {
  name: string;
  displayName: string;

  constructor(name: string, displayName: string) {
    this.name = name;
    this.displayName = displayName;
  }
}

export class ImageGroup {
  name: string;
  bucket: string;
  types: ImageType[];

  constructor(name: string, bucket: string, types: ImageType[]) {
    this.name = name;
    this.bucket = bucket;
    this.types = types;
  }
}

export class StaticInfo {
  softwareVersion: string;
  windowTitle: string;
  applicationTitle: string;
  faviconBase64: string;
  logoBase64: string;
  scaleInitialPercentage: number;
  maxImagesDisplayCount: number;
  tileServerURL: string;

  constructor(
    softwareVersion: string,
    windowTitle: string,
    applicationTitle: string,
    faviconBase64: string,
    logoBase64: string,
    scaleInitialPercentage: number,
    maxImagesDisplayCount: number,
    tileServerURL: string,
    imageGroups: ImageGroup[]
  ) {
    this.softwareVersion = softwareVersion;
    this.windowTitle = windowTitle;
    this.applicationTitle = applicationTitle;
    this.faviconBase64 = faviconBase64;
    this.logoBase64 = logoBase64;
    this.scaleInitialPercentage = scaleInitialPercentage;
    this.maxImagesDisplayCount = maxImagesDisplayCount;
    this.tileServerURL = tileServerURL;
    this.imageGroups = imageGroups;
  }

  imageGroups: ImageGroup[];
}

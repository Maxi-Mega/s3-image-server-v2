import type { DynamicData } from "@/models/dynamic_data.ts";
import type { StaticInfo } from "@/models/static_info";
import { defineStore } from "pinia";

export const useStaticInfoStore = defineStore("static-info", {
  state: () => {
    return {
      staticInfo: {} as StaticInfo,
      dynamicData: {} as Record<string, Record<string, DynamicData>>,
    };
  },
  actions: {
    setStaticInfo(staticInfo: StaticInfo): void {
      this.staticInfo = staticInfo;
    },
    getDynamicData(group: string, type: string): DynamicData | null {
      if (this.dynamicData[group]) {
        if (this.dynamicData[group][type]) {
          return this.dynamicData[group][type];
        }
      }

      return null;
    },
    setDynamicData(group: string, type: string, data: DynamicData): void {
      if (!this.dynamicData[group]) {
        this.dynamicData[group] = {};
      }

      this.dynamicData[group][type] = data;
    },
  },
});

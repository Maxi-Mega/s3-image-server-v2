import { defineStore } from "pinia";
import { ref } from "vue";

export const useFilterStore = defineStore("filters", {
  state: () => {
    return {
      checkedTypes: ref({} as Record<string, string[]>),
      searchQuery: "",
      globalFontSize: "13px", // TODO:
      globalScaleValue: 20, // calculate values based on initialScalePercentage
    };
  },
  actions: {
    reset() {
      this.searchQuery = "";
      this.checkedTypes = {};
    },
    setCheckedTypes(group: string, types: string[]) {
      this.checkedTypes[group] = types;
    },
    setGlobalSizes(fontSize: string, rawValue: number) {
      this.globalFontSize = fontSize;
      this.globalScaleValue = rawValue;
    },
  },
});

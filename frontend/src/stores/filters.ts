import { defineStore } from "pinia";

export const useFilterStore = defineStore("filters", {
  state: () => {
    return {
      tempFilters: [] as Record<string, string>[] | null,
      checkedFilters: {} as Record<string, Record<string, boolean>>,
      checkedTypes: {} as Record<string, string[]>,
      searchQuery: "",
      globalFontSize: "13px", // TODO:
      globalScaleValue: 20, // calculate values based on initialScalePercentage
    };
  },
  actions: {
    resetTypes() {
      this.searchQuery = "";
      Object.keys(this.checkedTypes).forEach((group) => delete this.checkedTypes[group]);
    },
    initFilterModes(modes: string[]) {
      Object.keys(this.checkedFilters).forEach((filter) => delete this.checkedFilters[filter]);
      for (const mode of modes) {
        this.checkedFilters[mode] = {};
      }

      if (this.tempFilters != null) {
        // Catchup filter options added before initialization
        const tmpFilters = this.tempFilters;
        this.tempFilters = null;
        for (const filterOptions of tmpFilters) {
          this.addFilterOptions(filterOptions);
        }
      }
    },
    addFilterOptions(filterOptions: Record<string, string>) {
      if (this.tempFilters != null) {
        // Filter modes are not initialized yet, store the filter options temporarily
        this.tempFilters.push(filterOptions);
        return;
      }

      for (const [filter, value] of Object.entries(filterOptions)) {
        if (this.checkedFilters[filter] && this.checkedFilters[filter][value] === undefined) {
          this.checkedFilters[filter][value] = true; // checked by default
        }
      }
    },
    toggleFilterValue(filter: string, value: string) {
      if (this.checkedFilters[filter]) {
        this.checkedFilters[filter][value] = !this.checkedFilters[filter][value];
      }
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

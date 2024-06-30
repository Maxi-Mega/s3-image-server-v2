<script setup lang="ts">
import { onMounted, onUnmounted, type Ref, ref } from "vue";
import { Scaler } from "@/composables/scaler";

const props = defineProps<{
  id: string;
  name: string;
  min: number;
  max: number;
  step: number;
  initialScalePercentage: number;
  baseScale: number;
  fullWidth?: boolean;
}>();

const emit = defineEmits<{
  (e: "change", fontSize: string, rawValue: number): void;
}>();

const scaler: Ref<Scaler | undefined> = ref();

onMounted(() => {
  const rangeInput = document.getElementById(props.id);
  if (!rangeInput) {
    throw new Error(`Range input '${props.id}' not found`);
  }

  scaler.value = new Scaler(
    rangeInput as HTMLInputElement,
    props.initialScalePercentage,
    props.baseScale
  );

  scaler.value.onUpdateScale = (fontSize: string, rawValue: number) => {
    emit("change", fontSize, rawValue);
  };
});

onUnmounted(() => {
  if (scaler.value instanceof Scaler) {
    scaler.value.dispose();
  }
});
</script>

<template>
  <label :for="id" class="sr-only">{{ name }}</label>
  <input
    type="range"
    :id="id"
    :min="min"
    :max="max"
    :step="step"
    :class="`${fullWidth ? 'w-full ' : ''}bg-transparent cursor-pointer appearance-none focus:outline-none disabled:pointer-events-none disabled:opacity-50
  [&::-moz-range-thumb]:h-2.5
  [&::-moz-range-thumb]:w-2.5
  [&::-moz-range-thumb]:appearance-none
  [&::-moz-range-thumb]:rounded-full
  [&::-moz-range-thumb]:border-4
  [&::-moz-range-thumb]:border-neutral-500
  [&::-moz-range-thumb]:bg-white
  [&::-moz-range-thumb]:transition-all
  [&::-moz-range-thumb]:duration-150
  [&::-moz-range-thumb]:ease-in-out

  [&::-moz-range-track]:h-2
  [&::-moz-range-track]:w-full
  [&::-moz-range-track]:rounded-full
  [&::-moz-range-track]:bg-gray-50
  [&::-webkit-slider-runnable-track]:h-2
  [&::-webkit-slider-runnable-track]:w-full
  [&::-webkit-slider-runnable-track]:rounded-full
  [&::-webkit-slider-runnable-track]:bg-gray-50
  [&::-webkit-slider-thumb]:-mt-0.5
  [&::-webkit-slider-thumb]:h-2.5

  [&::-webkit-slider-thumb]:w-2.5
  [&::-webkit-slider-thumb]:appearance-none
  [&::-webkit-slider-thumb]:rounded-full
  [&::-webkit-slider-thumb]:bg-white

  [&::-webkit-slider-thumb]:shadow-[0_0_0_4px_rgba(37,99,235,1)]
  [&::-webkit-slider-thumb]:transition-all
  [&::-webkit-slider-thumb]:duration-150
  [&::-webkit-slider-thumb]:ease-in-out`"
  />
</template>

<style scoped></style>

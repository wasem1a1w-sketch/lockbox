<script setup lang="ts">
import { SliderRange, SliderRoot, SliderThumb, SliderTrack } from "reka-ui";
import { cn } from "../../lib/utils";

const props = withDefaults(
  defineProps<{
    min?: number;
    max?: number;
    step?: number;
    disabled?: boolean;
    class?: string;
  }>(),
  { min: 0, max: 100, step: 1, disabled: false },
);

const model = defineModel<number[] | undefined>();
const emit = defineEmits<{ valueCommit: [number[]] }>();
</script>

<template>
  <SliderRoot
    v-model="model"
    :min="min"
    :max="max"
    :step="step"
    :disabled="disabled"
    :class="cn('relative flex w-full touch-none select-none items-center', props.class)"
    @value-commit="emit('valueCommit', $event)"
  >
    <SliderTrack class="relative h-1.5 w-full grow overflow-hidden rounded-full bg-secondary">
      <SliderRange class="absolute h-full bg-primary" />
    </SliderTrack>
    <SliderThumb
      aria-label="Value"
      class="block h-4 w-4 rounded-full border-2 border-primary bg-background shadow transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
    />
  </SliderRoot>
</template>

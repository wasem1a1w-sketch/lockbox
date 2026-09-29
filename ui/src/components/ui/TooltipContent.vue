<script setup lang="ts">
import { TooltipContent, TooltipPortal } from "reka-ui";
import { cn } from "../../lib/utils";

const props = defineProps<{ class?: string; sideOffset?: number }>();
</script>

<template>
  <TooltipPortal>
    <TooltipContent
      :side-offset="sideOffset ?? 4"
      :class="
        cn(
          'z-50 animate-fade-in overflow-hidden rounded-md border bg-popover px-2.5 py-1.5 text-xs text-popover-foreground shadow-md',
          props.class,
        )
      "
    >
      <slot />
    </TooltipContent>
  </TooltipPortal>
</template>

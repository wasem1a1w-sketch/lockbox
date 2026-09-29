<script setup lang="ts">
import { DropdownMenuContent, DropdownMenuPortal } from "reka-ui";
import { cn } from "../../lib/utils";

const props = defineProps<{ class?: string; align?: "start" | "center" | "end" }>();
</script>

<template>
  <DropdownMenuPortal>
    <DropdownMenuContent
      :side-offset="4"
      :align="align ?? 'start'"
      :class="
        cn(
          'z-50 min-w-[10rem] animate-fade-in overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md',
          props.class,
        )
      "
    >
      <slot />
    </DropdownMenuContent>
  </DropdownMenuPortal>
</template>

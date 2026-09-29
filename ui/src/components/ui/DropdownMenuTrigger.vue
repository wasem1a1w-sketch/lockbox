<script setup lang="ts">
import { DropdownMenuTrigger } from "reka-ui";
import { cn } from "../../lib/utils";

const props = defineProps<{ class?: string }>();
</script>

<template>
  <DropdownMenuTrigger
    :class="
      cn(
        'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 ring-offset-background disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground',
        props.class,
      )
    "
  >
    <slot />
  </DropdownMenuTrigger>
</template>

<script setup lang="ts">
import { TooltipRoot } from "reka-ui";

const open = defineModel<boolean>("open", { default: false });
</script>

<template>
  <TooltipRoot v-model:open="open">
    <slot />
  </TooltipRoot>
</template>

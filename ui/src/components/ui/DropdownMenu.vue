<script setup lang="ts">
import { DropdownMenuRoot } from "reka-ui";

const open = defineModel<boolean>("open", { default: false });
</script>

<template>
  <DropdownMenuRoot v-model:open="open">
    <slot />
  </DropdownMenuRoot>
</template>

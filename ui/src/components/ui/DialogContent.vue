<script setup lang="ts">
import { useAttrs } from "vue";
import { DialogClose, DialogContent, DialogOverlay, DialogPortal } from "reka-ui";
import { X } from "lucide-vue-next";
import { cn } from "../../lib/utils";

defineOptions({ inheritAttrs: false });

const attrs = useAttrs();
const props = defineProps<{ class?: string }>();
</script>

<template>
  <DialogPortal>
    <DialogOverlay class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm" />
    <DialogContent
      v-bind="attrs"
      :class="
        cn(
          'fixed top-1/2 left-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 animate-dialog-in rounded-xl border border-border/80 bg-card p-6 shadow-2xl shadow-black/40',
          props.class,
        )
      "
    >
      <slot />
      <DialogClose
        class="absolute top-4 right-4 rounded-sm text-muted-foreground opacity-70 transition-opacity hover:opacity-100 focus:outline-none"
        aria-label="Close"
      >
        <X class="h-4 w-4" />
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>

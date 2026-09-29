<script setup lang="ts">
import { AlertCircle, CheckCircle2, Info, X } from "lucide-vue-next";
import { dismissToast, toasts } from "../composables/useToast";
</script>

<template>
  <div class="fixed top-4 right-4 z-50 flex w-80 flex-col gap-2">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="animate-toast-in pointer-events-auto flex items-start gap-3 rounded-lg border bg-card p-4 shadow-2xl shadow-black/40"
      :class="[t.variant === 'error' && 'border-destructive/50', t.variant === 'success' && 'border-success/50']"
    >
      <CheckCircle2 v-if="t.variant === 'success'" class="mt-0.5 h-4 w-4 shrink-0 text-success" />
      <AlertCircle v-else-if="t.variant === 'error'" class="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
      <Info v-else class="mt-0.5 h-4 w-4 shrink-0 text-primary" />
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium">{{ t.title }}</p>
        <p v-if="t.description" class="mt-0.5 text-xs break-all text-muted-foreground">
          {{ t.description }}
        </p>
      </div>
      <button
        class="text-muted-foreground hover:text-foreground"
        aria-label="Dismiss"
        @click="dismissToast(t.id)"
      >
        <X class="h-4 w-4" />
      </button>
    </div>
  </div>
</template>

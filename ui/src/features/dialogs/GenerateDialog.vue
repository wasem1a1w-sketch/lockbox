<script setup lang="ts">
import { ref, watch } from "vue";
import { Copy, RefreshCw } from "lucide-vue-next";
import Button from "../../components/ui/Button.vue";
import Dialog from "../../components/ui/Dialog.vue";
import DialogContent from "../../components/ui/DialogContent.vue";
import DialogDescription from "../../components/ui/DialogDescription.vue";
import DialogHeader from "../../components/ui/DialogHeader.vue";
import DialogTitle from "../../components/ui/DialogTitle.vue";
import Label from "../../components/ui/Label.vue";
import Slider from "../../components/ui/Slider.vue";
import { toast } from "../../composables/useToast";
import { api } from "../../lib/api";

const open = defineModel<boolean>("open", { default: false });

const length = ref(12);
const password = ref("");

async function generate(len: number) {
  try {
    const res = await api<{ password: string }>("/api/generate", {
      method: "POST",
      body: { length: len },
    });
    password.value = res.password;
  } catch (err) {
    toast("Generate failed", { variant: "error", description: String(err) });
  }
}

watch(open, (o) => {
  if (o) generate(length.value);
});

function onSlide(v: number[] | undefined) {
  if (v && v[0] != null) length.value = v[0];
}

function onCommit(v: number[]) {
  if (v && v[0] != null) generate(v[0]);
}

async function copy() {
  if (!password.value) return;
  try {
    await navigator.clipboard.writeText(password.value);
    toast("Copied to clipboard", { variant: "success" });
  } catch {
    toast("Clipboard unavailable", { variant: "error" });
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Password generator</DialogTitle>
        <DialogDescription>
          Cryptographically strong random password (letters, digits, symbols).
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-5">
        <div class="rounded-lg border bg-background p-4">
          <p class="select-all break-all font-mono text-lg tracking-wide">
            {{ password || "…" }}
          </p>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label html-for="gen-length">Length</Label>
            <span class="text-sm tabular-nums text-muted-foreground">{{ length }}</span>
          </div>
          <Slider
            id="gen-length"
            :model-value="[length]"
            :min="8"
            :max="64"
            :step="1"
            @update:model-value="onSlide"
            @value-commit="onCommit"
          />
        </div>

        <div class="flex gap-2">
          <Button variant="outline" class="flex-1" @click="generate(length)">
            <RefreshCw class="h-4 w-4" />
            Regenerate
          </Button>
          <Button class="flex-1" :disabled="!password" @click="copy">
            <Copy class="h-4 w-4" />
            Copy
          </Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

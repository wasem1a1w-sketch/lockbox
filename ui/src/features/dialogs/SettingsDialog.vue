<script setup lang="ts">
import { ref, watch } from "vue";
import Button from "../../components/ui/Button.vue";
import Dialog from "../../components/ui/Dialog.vue";
import DialogContent from "../../components/ui/DialogContent.vue";
import DialogDescription from "../../components/ui/DialogDescription.vue";
import DialogFooter from "../../components/ui/DialogFooter.vue";
import DialogHeader from "../../components/ui/DialogHeader.vue";
import DialogTitle from "../../components/ui/DialogTitle.vue";
import Input from "../../components/ui/Input.vue";
import Label from "../../components/ui/Label.vue";
import { toast } from "../../composables/useToast";
import { ApiError, api, type ConfigInfo } from "../../lib/api";

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{ "session-invalidated": [] }>();

const info = ref<ConfigInfo | null>(null);
const vaultPath = ref("");
const iterations = ref("");
const genLength = ref("");
const busy = ref(false);
const error = ref("");

watch(open, (o) => {
  if (!o) return;
  api<ConfigInfo>("/api/config")
    .then((cfg) => {
      info.value = cfg;
      vaultPath.value = cfg.vaultPath;
      iterations.value = String(cfg.kdfIterations);
      genLength.value = String(cfg.defaultGenLength);
      error.value = "";
    })
    .catch((err) => (error.value = String(err)));
});

async function save() {
  busy.value = true;
  error.value = "";
  try {
    const changes: Array<[string, string]> = [];
    if (info.value && vaultPath.value !== info.value.vaultPath) {
      changes.push(["vault_path", vaultPath.value]);
    }
    if (info.value && Number(iterations.value) !== info.value.kdfIterations) {
      changes.push(["kdf_iterations", iterations.value]);
    }
    if (info.value && Number(genLength.value) !== info.value.defaultGenLength) {
      changes.push(["default_gen_length", genLength.value]);
    }

    if (changes.length === 0) {
      toast("No changes");
      open.value = false;
      return;
    }

    let invalidated = false;
    for (const [key, value] of changes) {
      const res = await api<{ sessionInvalidated?: boolean }>("/api/config", {
        method: "POST",
        body: { key, value },
      });
      invalidated = invalidated || Boolean(res.sessionInvalidated);
    }

    toast("Settings saved", { variant: "success" });
    open.value = false;
    if (invalidated) {
      toast("Vault path changed", { description: "Unlock again to continue." });
      emit("session-invalidated");
    }
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Save failed.";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Settings</DialogTitle>
        <DialogDescription>
          {{ info ? `Config file: ${info.configPath}` : "Loading configuration…" }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div class="space-y-2">
          <Label html-for="set-vault-path">Vault file path</Label>
          <Input id="set-vault-path" v-model="vaultPath" />
          <p v-if="info && !info.vaultExists" class="text-xs text-muted-foreground">
            No vault at this path yet.
          </p>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label html-for="set-iterations">KDF iterations</Label>
            <Input
              id="set-iterations"
              v-model="iterations"
              type="number"
              :min="1000"
            />
          </div>
          <div class="space-y-2">
            <Label html-for="set-gen-length">Default password length</Label>
            <Input
              id="set-gen-length"
              v-model="genLength"
              type="number"
              :min="4"
              :max="256"
            />
          </div>
        </div>

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
      </div>

      <DialogFooter class="mt-4">
        <Button variant="ghost" @click="open = false">Cancel</Button>
        <Button :disabled="busy || !info" @click="save">
          {{ busy ? "Saving…" : "Save settings" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

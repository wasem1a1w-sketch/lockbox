<script setup lang="ts">
import { ref } from "vue";
import Button from "../../components/ui/Button.vue";
import Dialog from "../../components/ui/Dialog.vue";
import DialogContent from "../../components/ui/DialogContent.vue";
import DialogDescription from "../../components/ui/DialogDescription.vue";
import DialogFooter from "../../components/ui/DialogFooter.vue";
import DialogHeader from "../../components/ui/DialogHeader.vue";
import DialogTitle from "../../components/ui/DialogTitle.vue";
import { toast } from "../../composables/useToast";
import { api } from "../../lib/api";

const open = defineModel<boolean>("open", { default: false });

const props = defineProps<{ credentialId: string | null; credentialLabel: string }>();
const emit = defineEmits<{ deleted: [] }>();

const busy = ref(false);

async function confirm() {
  if (!props.credentialId) return;
  busy.value = true;
  try {
    await api(`/api/vault/${props.credentialId}`, { method: "DELETE" });
    toast("Credential deleted", { variant: "success", description: props.credentialLabel });
    open.value = false;
    emit("deleted");
  } catch (err) {
    toast("Delete failed", { variant: "error", description: String(err) });
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>Delete credential?</DialogTitle>
        <DialogDescription>
          <span class="font-medium text-foreground">{{ credentialLabel }}</span> will be
          permanently removed from your vault. This cannot be undone.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="ghost" @click="open = false">Cancel</Button>
        <Button variant="destructive" :disabled="busy" @click="confirm">
          {{ busy ? "Deleting…" : "Delete" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

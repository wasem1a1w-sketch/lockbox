<script setup lang="ts">
import { ref } from "vue";
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
import { ApiError, api } from "../../lib/api";

const open = defineModel<boolean>("open", { default: false });

const password = ref("");
const confirm = ref("");
const error = ref("");
const busy = ref(false);

async function submit() {
  if (!password.value) {
    error.value = "New password cannot be blank.";
    return;
  }
  if (password.value !== confirm.value) {
    error.value = "Passwords do not match.";
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    await api("/api/change-master", {
      method: "POST",
      body: { newPassword: password.value },
    });
    toast("Master password changed", {
      variant: "success",
      description: "Vault re-encrypted with a fresh salt.",
    });
    password.value = "";
    confirm.value = "";
    open.value = false;
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Change failed.";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Change master password</DialogTitle>
        <DialogDescription>
          The whole vault is re-encrypted with a fresh salt using the new password.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <Label html-for="new-master">New master password</Label>
          <Input
            id="new-master"
            v-model="password"
            type="password"
            autofocus
            autocomplete="new-password"
          />
        </div>
        <div class="space-y-2">
          <Label html-for="confirm-master">Confirm new password</Label>
          <Input
            id="confirm-master"
            v-model="confirm"
            type="password"
            autocomplete="new-password"
          />
        </div>

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

        <DialogFooter>
          <Button type="button" variant="ghost" @click="open = false">Cancel</Button>
          <Button type="submit" :disabled="busy">
            {{ busy ? "Re-encrypting…" : "Change password" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

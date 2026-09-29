<script setup lang="ts">
import { ref, watch } from "vue";
import { Dice6, Eye, EyeOff } from "lucide-vue-next";
import Button from "../../components/ui/Button.vue";
import Dialog from "../../components/ui/Dialog.vue";
import DialogContent from "../../components/ui/DialogContent.vue";
import DialogDescription from "../../components/ui/DialogDescription.vue";
import DialogHeader from "../../components/ui/DialogHeader.vue";
import DialogTitle from "../../components/ui/DialogTitle.vue";
import Input from "../../components/ui/Input.vue";
import Label from "../../components/ui/Label.vue";
import { toast } from "../../composables/useToast";
import { ApiError, api, type Credential } from "../../lib/api";

const open = defineModel<boolean>("open", { default: false });

const props = defineProps<{ credential: Credential | null }>();
const emit = defineEmits<{ saved: [] }>();

const isEdit = () => props.credential !== null;

const account = ref("");
const username = ref("");
const password = ref("");
const show = ref(false);
const busy = ref(false);
const error = ref("");

watch(open, (o) => {
  if (o) {
    account.value = props.credential?.account ?? "";
    username.value = props.credential?.username ?? "";
    password.value = props.credential?.password ?? "";
    error.value = "";
    show.value = false;
  }
});

async function generate() {
  try {
    const res = await api<{ password: string }>("/api/generate", {
      method: "POST",
      body: { length: 0 },
    });
    password.value = res.password;
    show.value = true;
  } catch (err) {
    toast("Generate failed", { variant: "error", description: String(err) });
  }
}

async function submit() {
  if (!account.value || !username.value || !password.value) {
    error.value = "Account, username and password are all required.";
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    if (isEdit()) {
      await api(`/api/vault/${props.credential!.id}`, {
        method: "PUT",
        body: { account: account.value, username: username.value, password: password.value },
      });
      toast("Credential updated", { variant: "success", description: account.value });
    } else {
      await api("/api/vault", {
        method: "POST",
        body: { account: account.value, username: username.value, password: password.value },
      });
      toast("Credential added", { variant: "success", description: account.value });
    }
    open.value = false;
    emit("saved");
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
        <DialogTitle>{{ isEdit() ? "Edit credential" : "Add credential" }}</DialogTitle>
        <DialogDescription>
          {{
            isEdit()
              ? "Update this entry. All fields are saved."
              : "Store a new login in your encrypted vault."
          }}
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <Label html-for="cred-account">Account / website</Label>
          <Input
            id="cred-account"
            v-model="account"
            placeholder="example.com"
            autofocus
          />
        </div>
        <div class="space-y-2">
          <Label html-for="cred-username">Username</Label>
          <Input id="cred-username" v-model="username" placeholder="alice" />
        </div>
        <div class="space-y-2">
          <Label html-for="cred-password">Password</Label>
          <div class="relative flex gap-2">
            <Input
              id="cred-password"
              v-model="password"
              :type="show ? 'text' : 'password'"
              placeholder="p@ssw0rd"
              class="pr-10"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              :aria-label="show ? 'Hide password' : 'Show password'"
              @click="show = !show"
            >
              <EyeOff v-if="show" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon"
              title="Generate password"
              aria-label="Generate password"
              @click="generate"
            >
              <Dice6 class="h-4 w-4" />
            </Button>
          </div>
        </div>

        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

        <div class="flex justify-end gap-2">
          <Button type="button" variant="ghost" @click="open = false">Cancel</Button>
          <Button type="submit" :disabled="busy">
            {{ busy ? "Saving…" : isEdit() ? "Save changes" : "Add credential" }}
          </Button>
        </div>
      </form>
    </DialogContent>
  </Dialog>
</template>

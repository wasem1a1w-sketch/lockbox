<script setup lang="ts">
import { computed, ref } from "vue";
import { Eye, EyeOff, Lock, ShieldCheck } from "lucide-vue-next";
import Button from "../../components/ui/Button.vue";
import Input from "../../components/ui/Input.vue";
import Label from "../../components/ui/Label.vue";
import { ApiError, api } from "../../lib/api";
import { cn } from "../../lib/utils";

const props = defineProps<{ vaultExists: boolean }>();
const emit = defineEmits<{ unlocked: [] }>();

const password = ref("");
const confirm = ref("");
const show = ref(false);
const error = ref("");
const shake = ref(false);
const busy = ref(false);

const isNew = computed(() => !props.vaultExists);

function triggerShake() {
  shake.value = true;
  window.setTimeout(() => (shake.value = false), 450);
}

async function submit() {
  error.value = "";
  if (!isNew.value && password.value === "") {
    error.value = "Master password cannot be blank.";
    triggerShake();
    return;
  }
  if (isNew.value && password.value !== confirm.value) {
    error.value = "Passwords do not match.";
    triggerShake();
    return;
  }
  busy.value = true;
  try {
    await api("/api/unlock", {
      method: "POST",
      body: isNew.value
        ? { password: password.value, confirm: confirm.value }
        : { password: password.value },
    });
    password.value = "";
    confirm.value = "";
    emit("unlocked");
  } catch (err) {
    error.value =
      err instanceof ApiError ? err.message : "Could not reach the lockbox server.";
    triggerShake();
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <div
      class="pointer-events-none absolute -top-32 left-1/2 h-64 w-[36rem] -translate-x-1/2 animate-glow rounded-full bg-primary/25 blur-3xl"
    />

    <div
      :class="
        cn(
          'relative w-full max-w-sm rounded-2xl border border-black/5 bg-card/70 p-8 shadow-2xl shadow-black/50 backdrop-blur-xl dark:border-white/10',
          shake && 'animate-shake',
        )
      "
    >
      <div class="mb-6 flex flex-col items-center text-center">
        <div class="relative mb-4">
          <div class="absolute inset-0 rounded-2xl bg-primary/40 blur-xl" />
          <div
            class="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-indigo-500 text-white shadow-lg shadow-primary/30 ring-1 ring-black/10 dark:ring-white/15"
          >
            <ShieldCheck v-if="isNew" class="h-7 w-7" />
            <Lock v-else class="h-7 w-7" />
          </div>
        </div>
        <h1 class="text-xl font-semibold tracking-tight">
          {{ isNew ? "Create your vault" : "lockbox" }}
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          {{
            isNew
              ? "Choose a master password. It cannot be recovered."
              : "Enter your master password to unlock."
          }}
        </p>
      </div>

      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <Label html-for="master-password">Master password</Label>
          <div class="relative">
            <Input
              id="master-password"
              v-model="password"
              :type="show ? 'text' : 'password'"
              autofocus
              autocomplete="current-password"
              class="bg-background/60 pr-10"
            />
            <button
              type="button"
              class="absolute top-1/2 right-2 -translate-y-1/2 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              :aria-label="show ? 'Hide password' : 'Show password'"
              @click="show = !show"
            >
              <EyeOff v-if="show" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div v-if="isNew" class="space-y-2">
          <Label html-for="confirm-password">Confirm password</Label>
          <Input
            id="confirm-password"
            v-model="confirm"
            :type="show ? 'text' : 'password'"
            autocomplete="new-password"
            class="bg-background/60"
          />
        </div>

        <p
          v-if="error"
          class="animate-fade-in rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ error }}
        </p>

        <Button
          type="submit"
          class="w-full shadow-lg shadow-primary/25"
          :disabled="busy"
        >
          {{ busy ? "Please wait…" : isNew ? "Create vault" : "Unlock" }}
        </Button>
      </form>

      <div
        class="mt-6 flex items-center justify-center gap-1.5 border-t border-border/60 pt-4 text-xs text-muted-foreground"
      >
        <Lock class="h-3 w-3" />
        Local only · session auto-locks after inactivity
      </div>
    </div>
  </div>
</template>

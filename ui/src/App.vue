<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { TooltipProvider } from "reka-ui";
import { Loader2 } from "lucide-vue-next";
import Toasts from "./components/Toasts.vue";
import UnlockScreen from "./features/unlock/UnlockScreen.vue";
import VaultScreen from "./features/vault/VaultScreen.vue";
import { toast } from "./composables/useToast";
import { api, initToken, LOCKED_EVENT, type SessionInfo } from "./lib/api";

type Phase = "boot" | "locked" | "ready";

const phase = ref<Phase>("boot");
const vaultExists = ref(true);

function lock(notify: boolean) {
  phase.value = "locked";
  if (notify) {
    toast("Session locked", { description: "Vault locked after inactivity." });
  }
}

async function boot() {
  try {
    await initToken();
    const info = await api<SessionInfo>("/api/session");
    vaultExists.value = info.vaultExists;
    phase.value = info.unlocked ? "ready" : "locked";
  } catch {
    toast("Cannot reach server", { variant: "error" });
    phase.value = "locked";
  }
}

const onLocked = () => lock(true);

onMounted(() => {
  boot();
  window.addEventListener(LOCKED_EVENT, onLocked);
});
onUnmounted(() => window.removeEventListener(LOCKED_EVENT, onLocked));
</script>

<template>
  <TooltipProvider>
    <Toasts />
    <div
      v-if="phase === 'boot'"
      class="flex min-h-screen items-center justify-center"
    >
      <Loader2 class="h-8 w-8 animate-spin text-primary" />
    </div>
    <UnlockScreen
      v-else-if="phase === 'locked'"
      :vault-exists="vaultExists"
      @unlocked="
        () => {
          vaultExists = true;
          phase = 'ready';
        }
      "
    />
    <VaultScreen
      v-else
      @locked="lock(false)"
      @vault-created="vaultExists = true"
    />
  </TooltipProvider>
</template>

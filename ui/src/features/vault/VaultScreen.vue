<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue";
import Draggable from "vuedraggable";
import {
  Eye,
  EyeOff,
  KeyRound,
  Loader2,
  Lock,
  LockOpen,
  Moon,
  Plus,
  Search,
  Settings,
  Sparkles,
  Sun,
} from "lucide-vue-next";
import Row from "../../components/Row.vue";
import Button from "../../components/ui/Button.vue";
import Input from "../../components/ui/Input.vue";
import Tooltip from "../../components/ui/Tooltip.vue";
import TooltipContent from "../../components/ui/TooltipContent.vue";
import TooltipTrigger from "../../components/ui/TooltipTrigger.vue";
import CredentialDialog from "../dialogs/CredentialDialog.vue";
import ChangeMasterDialog from "../dialogs/ChangeMasterDialog.vue";
import DeleteDialog from "../dialogs/DeleteDialog.vue";
import GenerateDialog from "../dialogs/GenerateDialog.vue";
import SettingsDialog from "../dialogs/SettingsDialog.vue";
import { toast } from "../../composables/useToast";
import { api, type Credential } from "../../lib/api";
import { currentTheme, toggleTheme, type Theme } from "../../lib/theme";

const emit = defineEmits<{ locked: []; "vault-created": [] }>();

const creds = ref<Credential[]>([]);
const loading = ref(true);
const search = ref("");
const query = ref("");
const showAll = ref(false);

const addOpen = ref(false);
const editTarget = ref<Credential | null>(null);
const deleteTarget = ref<Credential | null>(null);
const genOpen = ref(false);
const masterOpen = ref(false);
const settingsOpen = ref(false);

const theme = ref<Theme>(currentTheme());
const searchRef = ref<{ focus: () => void } | null>(null);

let previous: Credential[] | null = null;

async function load(q: string) {
  try {
    const res = await api<{ credentials: Credential[] }>(
      `/api/vault?search=${encodeURIComponent(q)}`,
    );
    creds.value = res.credentials;
  } catch {
    toast("Failed to load vault", { variant: "error" });
  } finally {
    loading.value = false;
  }
}

// Debounced search.
let searchTimer = 0;
watch(search, (value) => {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(() => {
    query.value = value;
    load(value);
  }, 250);
});

// "/" focuses search.
function onKey(e: KeyboardEvent) {
  const target = e.target as HTMLElement;
  const typing =
    target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable;
  if (e.key === "/" && !typing) {
    e.preventDefault();
    searchRef.value?.focus();
  }
}

onMounted(() => {
  load("");
  window.addEventListener("keydown", onKey);
});
onUnmounted(() => {
  window.removeEventListener("keydown", onKey);
  window.clearTimeout(searchTimer);
});

async function saveOrder(id: string, to: number, rollback: () => void) {
  try {
    await api("/api/vault/reorder", { method: "POST", body: { id, to } });
    toast("Order saved", { variant: "success" });
  } catch {
    rollback();
    toast("Reorder failed", { variant: "error" });
  }
}

function onDragStart() {
  previous = [...creds.value];
}

function onDragEnd(evt: { oldIndex?: number; newIndex?: number }) {
  const { oldIndex, newIndex } = evt;
  if (oldIndex == null || newIndex == null || oldIndex === newIndex) return;
  const id = creds.value[newIndex]?.id;
  if (!id || !previous) return;
  const before = previous;
  previous = null;
  saveOrder(id, newIndex + 1, () => {
    creds.value = before;
  });
}

async function keyboardMove(index: number, delta: number) {
  const target = index + delta;
  if (target < 0 || target >= creds.value.length) return;
  const before = [...creds.value];
  const item = before[index];
  const arr = [...before];
  arr.splice(index, 1);
  arr.splice(target, 0, item);
  creds.value = arr;
  await saveOrder(item.id, target + 1, () => {
    creds.value = before;
  });
}

async function lockNow() {
  try {
    await api("/api/lock", { method: "POST" });
  } finally {
    emit("locked");
  }
}

function toggleThemeMode() {
  theme.value = toggleTheme();
}

function setEditOpen(o: boolean) {
  if (!o) editTarget.value = null;
}

function setDeleteOpen(o: boolean) {
  if (!o) deleteTarget.value = null;
}
</script>

<template>
  <div class="min-h-screen">
    <!-- Header -->
    <header class="sticky top-0 z-30 border-b border-border/70 bg-background/80 backdrop-blur-xl">
      <div class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
        <div class="flex items-center gap-2.5">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-indigo-500 text-white shadow-md shadow-primary/25 ring-1 ring-black/10 dark:ring-white/15"
          >
            <Lock class="h-4 w-4" />
          </div>
          <span class="font-semibold tracking-tight">lockbox</span>
          <span
            v-if="!loading && !query"
            class="ml-1 hidden rounded-full border border-border/60 bg-secondary/50 px-2 py-0.5 text-[11px] tabular-nums text-muted-foreground sm:inline-block"
          >
            {{ creds.length }} {{ creds.length === 1 ? "credential" : "credentials" }}
          </span>
        </div>
        <div class="flex items-center gap-1">
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                :aria-label="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
                @click="toggleThemeMode"
              >
                <Sun v-if="theme === 'dark'" class="h-4 w-4" />
                <Moon v-else class="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{{ theme === "dark" ? "Light mode" : "Dark mode" }}</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Change master password"
                @click="masterOpen = true"
              >
                <KeyRound class="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Change master password</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" aria-label="Settings" @click="settingsOpen = true">
                <Settings class="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Settings</TooltipContent>
          </Tooltip>
          <Button variant="outline" size="sm" class="ml-2" @click="lockNow">
            <LockOpen class="h-3.5 w-3.5" />
            Lock
          </Button>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-5xl px-4 py-8">
      <!-- Toolbar -->
      <div class="mb-5 flex flex-wrap items-center gap-3">
        <div class="relative min-w-56 flex-1">
          <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            ref="searchRef"
            v-model="search"
            placeholder="Search accounts, users, passwords…  (press /)"
            class="rounded-lg border-border/70 bg-card/60 pr-12 pl-9 transition-colors focus-visible:bg-card"
          />
          <kbd
            v-if="!search"
            class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 rounded border border-border/60 bg-secondary/60 px-1.5 py-0.5 text-[10px] text-muted-foreground"
          >
            /
          </kbd>
        </div>

        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="outline"
              size="icon"
              :aria-label="showAll ? 'Hide passwords' : 'Show passwords'"
              @click="showAll = !showAll"
            >
              <EyeOff v-if="showAll" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{{ showAll ? "Hide" : "Show" }} all passwords</TooltipContent>
        </Tooltip>

        <Button variant="outline" @click="genOpen = true">
          <Sparkles class="h-4 w-4" />
          Generate
        </Button>
        <Button class="shadow-lg shadow-primary/25" @click="addOpen = true">
          <Plus class="h-4 w-4" />
          Add credential
        </Button>
      </div>

      <!-- Table -->
      <div v-if="loading" class="flex justify-center py-20">
        <Loader2 class="h-6 w-6 animate-spin text-primary" />
      </div>
      <div
        v-else-if="creds.length === 0"
        class="flex flex-col items-center rounded-xl border border-dashed border-border/70 bg-card/30 py-20 text-center"
      >
        <div class="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-primary/20">
          <Lock class="h-6 w-6 text-primary/70" />
        </div>
        <p class="font-medium">
          {{ query ? "No credentials matched" : "Your vault is empty" }}
        </p>
        <p class="mt-1 mb-5 text-sm text-muted-foreground">
          {{
            query
              ? `Nothing matches “${query}”.`
              : "Add your first credential to get started."
          }}
        </p>
        <Button v-if="!query" class="shadow-lg shadow-primary/25" @click="addOpen = true">
          <Plus class="h-4 w-4" />
          Add credential
        </Button>
      </div>
      <div
        v-else
        class="overflow-hidden rounded-xl border border-border/70 bg-card shadow-xl shadow-black/20"
      >
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border/70 bg-secondary/40 text-left text-xs uppercase tracking-wider text-muted-foreground">
              <th class="w-10 px-3 py-2.5" />
              <th class="px-3 py-2.5 font-medium">Account</th>
              <th class="px-3 py-2.5 font-medium">Username</th>
              <th class="px-3 py-2.5 font-medium">Password</th>
              <th class="hidden px-3 py-2.5 font-medium sm:table-cell">Saved</th>
              <th class="w-12 px-3 py-2.5" />
            </tr>
          </thead>
          <Draggable
            v-model="creds"
            tag="tbody"
            item-key="id"
            handle=".grip"
            :animation="150"
            @start="onDragStart"
            @end="onDragEnd"
          >
            <template #item="{ element, index }">
              <Row
                :credential="element"
                :index="index"
                :show-password="showAll"
                @edit="editTarget = element"
                @delete="deleteTarget = element"
                @move="keyboardMove"
              />
            </template>
          </Draggable>
        </table>
      </div>
    </main>

    <!-- Dialogs -->
    <CredentialDialog v-model:open="addOpen" :credential="null" @saved="load(query)" />
    <CredentialDialog
      :open="editTarget !== null"
      :credential="editTarget"
      @update:open="setEditOpen"
      @saved="load(query)"
    />
    <DeleteDialog
      :open="deleteTarget !== null"
      :credential-id="deleteTarget?.id ?? null"
      :credential-label="
        deleteTarget ? `${deleteTarget.account} (${deleteTarget.username})` : ''
      "
      @update:open="setDeleteOpen"
      @deleted="load(query)"
    />
    <GenerateDialog v-model:open="genOpen" />
    <ChangeMasterDialog v-model:open="masterOpen" />
    <SettingsDialog
      v-model:open="settingsOpen"
      @session-invalidated="
        () => {
          emit('vault-created');
          emit('locked');
        }
      "
    />
  </div>
</template>

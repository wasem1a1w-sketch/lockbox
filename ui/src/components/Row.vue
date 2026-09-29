<script setup lang="ts">
import { Copy, GripVertical, MoreHorizontal, Pencil, Trash2 } from "lucide-vue-next";
import DropdownMenu from "./ui/DropdownMenu.vue";
import DropdownMenuContent from "./ui/DropdownMenuContent.vue";
import DropdownMenuItem from "./ui/DropdownMenuItem.vue";
import DropdownMenuTrigger from "./ui/DropdownMenuTrigger.vue";
import { toast } from "../composables/useToast";
import type { Credential } from "../lib/api";
import { cn } from "../lib/utils";

const props = defineProps<{
  credential: Credential;
  index: number;
  showPassword: boolean;
}>();

const emit = defineEmits<{
  edit: [];
  delete: [];
  move: [index: number, delta: number];
}>();

function hueOf(seed: string) {
  let h = 0;
  for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) % 360;
  return h;
}

function initials(name: string) {
  const clean = name.replace(/[^a-zA-Z0-9]/g, "");
  return (clean.slice(0, 2) || "?").toUpperCase();
}

const h = hueOf(props.credential.account);
const masked = "•".repeat(Math.min(Math.max(props.credential.password.length, 6), 16));

async function copyPassword() {
  try {
    await navigator.clipboard.writeText(props.credential.password);
    toast("Password copied", { variant: "success", description: props.credential.account });
  } catch {
    toast("Clipboard unavailable", { variant: "error" });
  }
}

function onGripKeydown(e: KeyboardEvent) {
  if (e.key === "ArrowUp") {
    e.preventDefault();
    emit("move", props.index, -1);
  } else if (e.key === "ArrowDown") {
    e.preventDefault();
    emit("move", props.index, 1);
  }
}
</script>

<template>
  <tr class="group border-b border-border/50 transition-colors last:border-0 hover:bg-accent/50">
    <td class="px-3 py-2.5">
      <button
        class="grip cursor-grab touch-none text-muted-foreground/50 opacity-0 transition-opacity group-hover:opacity-100 active:cursor-grabbing hover:text-foreground"
        :aria-label="`Reorder ${credential.account} (use arrow keys)`"
        @keydown="onGripKeydown"
      >
        <GripVertical class="h-4 w-4" />
      </button>
    </td>
    <td class="px-3 py-2.5">
      <div class="flex items-center gap-3">
        <div
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white ring-1 ring-black/10 dark:ring-white/15"
          :style="{
            background: `linear-gradient(135deg, hsl(${h} 65% 55%), hsl(${(h + 45) % 360} 65% 42%))`,
          }"
          aria-hidden
        >
          {{ initials(credential.account) }}
        </div>
        <span class="font-medium">{{ credential.account }}</span>
      </div>
    </td>
    <td class="px-3 py-2.5 text-muted-foreground">{{ credential.username }}</td>
    <td class="px-3 py-2.5">
      <div class="flex items-center gap-1.5">
        <span
          :class="
            cn(
              'select-all rounded px-1.5 py-0.5 font-mono text-xs tracking-wider',
              showPassword ? 'bg-secondary/70 text-foreground' : 'text-muted-foreground/80',
            )
          "
        >
          {{ showPassword ? credential.password : masked }}
        </span>
        <button
          class="rounded-md p-1 text-muted-foreground opacity-0 transition-all group-hover:opacity-100 hover:bg-accent hover:text-foreground focus-visible:opacity-100"
          aria-label="Copy password"
          @click="copyPassword"
        >
          <Copy class="h-3.5 w-3.5" />
        </button>
      </div>
    </td>
    <td class="hidden px-3 py-2.5 text-xs text-muted-foreground sm:table-cell">
      {{ credential.savedAt }}
    </td>
    <td class="px-3 py-2.5">
      <DropdownMenu>
        <DropdownMenuTrigger
          class="p-1.5 text-muted-foreground"
          :aria-label="`Actions for ${credential.account}`"
        >
          <MoreHorizontal class="h-4 w-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem @select="emit('edit')">
            <Pencil class="h-4 w-4" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuItem destructive @select="emit('delete')">
            <Trash2 class="h-4 w-4" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </td>
  </tr>
</template>

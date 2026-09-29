import { ref } from "vue";

export type ToastVariant = "success" | "error" | "info";

export interface ToastItem {
  id: number;
  title: string;
  description?: string;
  variant: ToastVariant;
}

export const toasts = ref<ToastItem[]>([]);

let nextId = 1;

export function toast(title: string, opts?: { description?: string; variant?: ToastVariant }) {
  const id = nextId++;
  toasts.value.push({
    id,
    title,
    description: opts?.description,
    variant: opts?.variant ?? "info",
  });
  window.setTimeout(() => dismissToast(id), 3500);
}

export function dismissToast(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id);
}

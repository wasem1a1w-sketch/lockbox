export interface Credential {
  id: string;
  account: string;
  username: string;
  password: string;
  savedAt: string;
  sortOrder: number;
}

export interface SessionInfo {
  vaultExists: boolean;
  unlocked: boolean;
}

export interface ConfigInfo {
  vaultPath: string;
  kdfIterations: number;
  defaultGenLength: number;
  configPath: string;
  vaultExists: boolean;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
  }
}

function injectedToken(): string {
  const t = (window as unknown as { __LOCKBOX_TOKEN__?: string }).__LOCKBOX_TOKEN__;
  return t && !t.includes("__LOCKBOX_TOKEN__") ? t : "";
}

let token = "";

// initToken resolves the per-boot token: injected in prod builds,
// bootstrapped from /api/token in dev (Vite proxy).
export async function initToken(): Promise<void> {
  token = injectedToken();
  if (!token) {
    const res = await fetch("/api/token");
    const data = await res.json();
    token = data.token;
  }
}

export const LOCKED_EVENT = "lockbox:locked";

export async function api<T>(
  path: string,
  opts: { method?: string; body?: unknown } = {},
): Promise<T> {
  const res = await fetch(path, {
    method: opts.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      "X-Lockbox-Token": token,
    },
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
  });

  if (res.status === 401) {
    window.dispatchEvent(new Event(LOCKED_EVENT));
    throw new ApiError("Vault locked", 401);
  }

  const data = (await res.json().catch(() => ({}))) as Record<string, unknown>;
  if (!res.ok) {
    throw new ApiError(
      typeof data.error === "string" ? data.error : res.statusText,
      res.status,
    );
  }
  return data as T;
}

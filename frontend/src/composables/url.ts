const baseURL = import.meta.env.BASE_URL.replace(/\/$/, "");

export function resolveBackendURL(path: string): string {
  return baseURL + path;
}

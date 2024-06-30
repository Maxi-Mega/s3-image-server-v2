const baseURL = import.meta.env.BASE_URL.replace(/\/$/, "");

const wsProto = window.location.protocol === "https:" ? "wss:" : "ws:";
export const wsURL = `${wsProto}//${window.location.host}${baseURL}/api/ws`;

export function resolveBackendURL(path: string): string {
  return baseURL + path;
}

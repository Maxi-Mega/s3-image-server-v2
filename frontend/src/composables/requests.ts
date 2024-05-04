import type { StaticInfo } from "@/models/static_info";
import { resolveBackendURL } from "@/composables/url";

export async function fetchStaticInfo(): Promise<StaticInfo> {
  return fetch(resolveBackendURL("/api/info"))
    .then((resp) => (resp.ok ? resp.json() : Promise.reject(resp)))
    .then((info) => info as StaticInfo);
}

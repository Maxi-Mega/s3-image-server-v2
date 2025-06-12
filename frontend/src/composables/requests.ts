import { resolveBackendURL } from "@/composables/url";
import type { StaticInfo } from "@/models/static_info";

export async function fetchStaticInfo(): Promise<StaticInfo> {
  return fetch(resolveBackendURL("/api/info"))
    .then((resp) => (resp.ok ? resp.json() : Promise.reject(resp)))
    .then((info) => info as StaticInfo);
}

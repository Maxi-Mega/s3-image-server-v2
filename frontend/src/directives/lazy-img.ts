// src/directives/lazy-img.ts
import type { Directive } from "vue";

type Value = {
  src: string;
  srcset?: string;
  sizes?: string;
  rootMargin?: string; // e.g. '400px 0px'
  onLoaded?: (img: HTMLImageElement) => void;
  onError?: (img: HTMLImageElement, ev: Event) => void;
};

export const vLazyImg: Directive<HTMLImageElement, string | Value> = {
  mounted(el, binding) {
    const opts: Value = typeof binding.value === "string" ? { src: binding.value } : binding.value;

    // If no IO support, just load
    if (!("IntersectionObserver" in window)) {
      apply(el, opts);
      return;
    }

    const rootMargin = opts.rootMargin ?? "400px 0px";
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            apply(el, opts);
            io.unobserve(el);
          }
        }
      },
      { rootMargin }
    );
    /* eslint-disable @typescript-eslint/no-explicit-any */
    (el as any).__vLazyImgIO = io;
    io.observe(el);
  },
  unmounted(el) {
    /* eslint-disable @typescript-eslint/no-explicit-any */
    const io: IntersectionObserver | undefined = (el as any).__vLazyImgIO;
    io?.unobserve(el);
    io?.disconnect();
  },
};

function apply(el: HTMLImageElement, v: Value) {
  if (v.srcset) el.setAttribute("srcset", v.srcset);
  if (v.sizes) el.setAttribute("sizes", v.sizes);
  el.addEventListener("load", () => v.onLoaded?.(el), { once: true });
  el.addEventListener("error", (ev) => v.onError?.(el, ev), { once: true });
  el.src = v.src;
}

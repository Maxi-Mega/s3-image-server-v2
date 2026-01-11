<script setup lang="ts">
import ConfigDoc from "@/../doc/config_doc.md";
import ExprDoc from "@/../doc/expr_doc.md";
</script>

<template>
  <main class="min-h-screen bg-gray-800 px-6 py-16 text-slate-100">
    <section class="mx-auto w-full max-w-4xl rounded-2xl border border-white/10 bg-white/5 p-8">
      <header class="mb-10">
        <h1 class="mt-3 text-3xl font-semibold text-white">Documentation</h1>
        <p class="mt-3 text-sm text-slate-300">
          Reference guides for configuring the server and writing expressions.
        </p>
      </header>
      <article class="markdown-content">
        <ConfigDoc />
        <hr class="my-10 border-white/15" />
        <ExprDoc />
      </article>
    </section>
  </main>
</template>

<style scoped>
.markdown-content :deep(h1) {
  margin: 1.75rem 0 0.75rem;
  font-size: 1.75rem;
  font-weight: 600;
  color: #ffffff;
}

.markdown-content :deep(h2) {
  margin: 1.5rem 0 0.5rem;
  font-size: 1.35rem;
  font-weight: 600;
  color: #e2e8f0;
}

.markdown-content :deep(h3) {
  margin: 1.25rem 0 0.4rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #e2e8f0;
}

.markdown-content :deep(p) {
  margin: 0.75rem 0;
  line-height: 1.7;
  color: #cbd5f5;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0.75rem 0 0.75rem 1.5rem;
  line-height: 1.7;
  color: #cbd5f5;
}

.markdown-content :deep(ul) {
  list-style: disc;
}

.markdown-content :deep(ol) {
  list-style: decimal;
}

.markdown-content :deep(li::marker) {
  color: #f8fafc;
}

.markdown-content :deep(code) {
  border-radius: 0.35rem;
  background: rgba(15, 23, 42, 0.8);
  padding: 0.15rem 0.35rem;
  font-size: 0.9rem;
  color: #f8fafc;
}

.markdown-content :deep(pre) {
  margin: 1rem 0;
  overflow: auto;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(2, 6, 23, 0.9);
  padding: 1rem;
  color: #e2e8f0;
}

.markdown-content :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 0.9rem;
  color: inherit;
}

.markdown-content :deep(a) {
  color: #7dd3fc;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 3px;
}

.markdown-content :deep(blockquote) {
  margin: 1rem 0;
  border-left: 3px solid rgba(148, 163, 184, 0.7);
  padding-left: 1rem;
  color: #e2e8f0;
}

.markdown-content :deep(table) {
  margin: 1rem 0;
  width: 100%;
  border-collapse: collapse;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid rgba(148, 163, 184, 0.35);
  padding: 0.6rem 0.75rem;
  text-align: left;
}

.markdown-content :deep(th) {
  background: rgba(15, 23, 42, 0.6);
  color: #f8fafc;
}
</style>

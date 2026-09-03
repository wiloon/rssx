# 前端从 Vue 2 迁移到 Vue 3

> 本文件是 `0002-migrate-frontend-to-vue-3.md` 的中文版本，内容与英文版保持一致。以英文版为准。

## 状态

accepted（已接受）

## 背景

`AGENTS.md` 把前端描述为 "Vue.js 3（Composition API）…… Vuetify 3"，但
`rssx-ui/package.json` 实际是 Vue 2.6.11、Vuetify 2.6.3、`vue-class-component` +
`vue-property-decorator`、Vue Router 3、Vuex 3，跑在 `@vue/cli-service` 4.5 上。
Vue 2 已于 2023 年 12 月停止维护，不再有安全修复。而三栏布局（ADR-0001）本来就会重写
每一个视图组件。

## 决定

把 Vue 3 迁移作为三栏布局工作的一部分来做，而不是做完之后再做。目标：Vue 3 + Vuetify 3、
`<script setup>` Composition API、用 Pinia 取代 Vuex、Vue Router 4，以及用 Vite 取代
`@vue/cli-service`。移除 `vue-class-component` 和 `vue-property-decorator`。迁移落地后
更新 `AGENTS.md`，让技术栈描述与实际一致。

把组件重写两遍——先搬到 Vue 2 上、再为 Vue 3 重写一遍——是白费功夫，而且 class-component
装饰器在 Vue 3 里没有未来。

## 考虑过的方案

- **现在先在 Vue 2 上做布局，之后再上 Vue 3。** 能更早交付可见的改动，但丢掉了组件重写的
  成果，还让一个已 EOL 的框架继续留在项目里且没有后续排期。
- **先做 Vue 3 迁移构建（compat 兼容模式），布局随后。** 用 `@vue/compat` 垫片做一次专门的
  迁移能降低风险，但当前应用很小（六个视图、一个 store 模块）——配合布局直接重写，比"兼容
  过一遍再布局过一遍"总工作量更小。

## 后果

- 单次改动更大：构建工具、路由、状态库一次性全换。缓解方式是应用体量小，且这些都在组件反正
  要被重写的时候一起做。
- 贡献者需要与新配置匹配的 Node/工具链；`.node-version` 和 CI 要在同一次改动里更新。
- 迁移中被移除、需要以后专门恢复的：
  - **PWA / service worker**（`vue-cli-plugin-pwa`、`registerServiceWorker`）——
    如果还需要离线支持，用 `vite-plugin-pwa` 重新加。
  - **Cypress e2e**——旧用例针对的是 ADR-0001 之前的路由；三栏流程的 e2e 要重写。
  - **类型感知的 ESLint**——目前 lint 只跑 `eslint:recommended` +
    `plugin:vue/vue3-essential`；`typescript-eslint` 规则是后续工作。
- `Reader.vue` 目前是一个接了 store 的最小骨架，只为让应用端到端跑起来。它的组件行为和窄屏
  折叠（ADR-0001）是下一个 TDD 阶段的内容。

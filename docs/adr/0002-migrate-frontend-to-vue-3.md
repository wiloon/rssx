# Migrate the frontend from Vue 2 to Vue 3

## Status

accepted

## Context

`AGENTS.md` describes the frontend as "Vue.js 3 (Composition API) ... Vuetify 3",
but `rssx-ui/package.json` is Vue 2.6.11, Vuetify 2.6.3, `vue-class-component` +
`vue-property-decorator`, Vue Router 3, Vuex 3, on `@vue/cli-service` 4.5. Vue 2
reached end of life in December 2023 and receives no further security fixes.
The three-pane layout (ADR-0001) rewrites every view component anyway.

## Decision

Do the Vue 3 migration as part of the three-pane layout work rather than after
it. Target: Vue 3 + Vuetify 3, `<script setup>` Composition API, Pinia in place
of Vuex, Vue Router 4, and Vite in place of `@vue/cli-service`. Drop
`vue-class-component` and `vue-property-decorator`. Update `AGENTS.md` so its
stack description matches reality once the migration lands.

Rewriting the components twice — once to move the layout onto Vue 2, then again
for Vue 3 — is wasted work, and the class-component decorators have no Vue 3
future.

## Considered options

- **Layout on Vue 2 now, Vue 3 later.** Ships the visible change sooner but
  throws away the component rewrite and leaves an EOL framework in place with no
  scheduled follow-up.
- **Vue 3 migration build (compat mode) first, layout after.** A dedicated
  migration with the `@vue/compat` shim reduces risk, but the current app is
  small (six views, one store module) — a direct rewrite alongside the layout is
  less total effort than a compat pass plus a layout pass.

## Consequences

- Bigger single change: build tooling, router, and state library all move at
  once. Mitigated by the app's small size and by doing it while the components
  are being rewritten regardless.
- Contributors need Node/tooling matching the new setup; `.node-version` and CI
  must be updated in the same change.
- Dropped in the migration, to be restored deliberately later:
  - **PWA / service worker** (`vue-cli-plugin-pwa`, `registerServiceWorker`) —
    re-add with `vite-plugin-pwa` if offline support is still wanted.
  - **Cypress e2e** — the old specs targeted the pre-ADR-0001 routes; e2e for
    the three-pane flow is a fresh write.
  - **Type-aware ESLint** — lint currently runs `eslint:recommended` +
    `plugin:vue/vue3-essential` only; `typescript-eslint` rules are a follow-up.
- `Reader.vue` landed as a minimal store-wired shell so the app runs end to end.
  Its component behaviour and the narrow-viewport collapse (ADR-0001) are the
  next TDD phase.

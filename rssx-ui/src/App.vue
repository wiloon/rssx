<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getJwtToken, removeJwtToken } from '@/utils/auth'

const route = useRoute()
const router = useRouter()

// Re-evaluated on every navigation, which covers login/logout.
const loggedIn = computed(() => route.fullPath !== '' && Boolean(getJwtToken()))

function logout (): void {
  removeJwtToken()
  router.push({ name: 'Login' })
}
</script>

<template>
  <v-app>
    <v-app-bar density="compact" color="primary" elevation="2">
      <v-app-bar-title class="app-title">
        <v-icon icon="mdi-rss" size="20" class="app-title__icon" />
        RSSX
      </v-app-bar-title>
      <template #append>
        <v-btn
          v-if="loggedIn"
          data-cy="manage-feeds"
          variant="text"
          prepend-icon="mdi-format-list-bulleted"
          :to="{ name: 'FeedManager' }"
        >
          Feeds
        </v-btn>
        <v-btn
          v-if="loggedIn"
          data-cy="logout"
          variant="text"
          prepend-icon="mdi-logout"
          @click="logout"
        >
          Logout
        </v-btn>
      </template>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<style scoped>
.app-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  letter-spacing: 0.3px;
}
.app-title__icon {
  opacity: 0.9;
}
</style>

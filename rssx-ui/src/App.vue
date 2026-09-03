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
    <v-app-bar density="compact" flat>
      <v-app-bar-title>RSSX</v-app-bar-title>
      <template #append>
        <v-btn v-if="loggedIn" data-cy="logout" variant="text" @click="logout">
          退出
        </v-btn>
      </template>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

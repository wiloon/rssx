<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { axiosInstance } from '@/api/http'
import { setJwtToken } from '@/utils/auth'

const router = useRouter()
const name = ref('')
const password = ref('')
const snackbar = ref(false)
const msg = ref('')

async function login (): Promise<void> {
  const { data } = await axiosInstance.post('/login', {
    name: name.value,
    password: password.value
  })
  if (data.code === 20000) {
    setJwtToken(data.data.token)
    router.push({ name: 'Reader' })
  } else {
    msg.value = data.message
    snackbar.value = true
  }
}
</script>

<template>
  <v-container class="mx-auto" style="max-width: 400px">
    <v-text-field v-model="name" data-cy="user-name" label="用户名" />
    <v-text-field
      v-model="password"
      data-cy="password"
      label="密码"
      type="password"
    />
    <v-btn block color="primary" data-cy="login" @click="login">登录</v-btn>
    <v-snackbar v-model="snackbar" :timeout="3000">{{ msg }}</v-snackbar>
  </v-container>
</template>

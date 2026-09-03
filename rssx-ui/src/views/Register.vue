<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { axiosInstance } from '@/api/http'
import { setJwtToken } from '@/utils/auth'

const router = useRouter()
const name = ref('')
const password = ref('')
const passwordCheck = ref('')
const snackbar = ref(false)
const msg = ref('')

async function register (): Promise<void> {
  if (password.value !== passwordCheck.value) {
    msg.value = '两次输入的密码不一致'
    snackbar.value = true
    return
  }
  const { data } = await axiosInstance.post('/register', {
    name: name.value,
    password: password.value
  })
  if (data.code !== 20000) {
    msg.value = data.message
    snackbar.value = true
    return
  }
  setJwtToken(data.data.token)
  router.push({ name: 'Reader' })
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
    <v-text-field
      v-model="passwordCheck"
      data-cy="password-check"
      label="密码确认"
      type="password"
    />
    <v-btn block color="primary" data-cy="register" @click="register">
      注册
    </v-btn>
    <v-snackbar v-model="snackbar" :timeout="3000">{{ msg }}</v-snackbar>
  </v-container>
</template>

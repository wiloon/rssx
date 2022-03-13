<template>
  <div>
    <v-text-field
      data-cy="user-name"
      v-model="name"
      label="用户名">
    </v-text-field>
    <v-text-field
      data-cy="password"
      v-model="password"
      label="密码"
      type="password"
      counter>
    </v-text-field>
    <v-btn
      block
      color="primary"
      @click="login"
      data-cy="login"
      style="margin-right: 10px"
    >登录
    </v-btn>
    <v-snackbar
      v-model="snackbar" timeout="3000"
    >
      {{ msg }}
      <template v-slot:action="{ attrs }">
        <v-btn
          color="pink"
          text
          v-bind="attrs"
          @click="snackbar = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>

</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Axios from 'axios'

@Component({
  components: {}
})
export default class Login extends Vue {
  name = ''
  password = ''
  snackbar = false
  msg = ''

  login (event: any): void {
    console.log('login')
    Axios.post('/login', {
      name: this.name,
      password: this.password
    }).then((response: any) => {
      console.log('response.status: ' + response.status)
      console.log('foo: ' + response.data.code)
      if (response.data.code === 20000) {
        const token = response.data.data.token
        localStorage.setItem('token', token)
        this.$router.push({ name: 'FeedList' })
      } else {
        this.msg = response.data.message
        this.snackbar = true
      }
    })
  }

  mounted () {
    console.log('login mounted')
  }
}
</script>

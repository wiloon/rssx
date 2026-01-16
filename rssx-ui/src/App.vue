<template>
  <v-app id="inspire">
    <v-app-bar
      app
      dense
    >
      <v-toolbar-title>RSSX</v-toolbar-title>
      <v-btn
        data-cy="logout-dialog"
        v-on:click="logout"
        class="navBtn"
        depressed
        v-if="this.$store.state.login">
        退出
      </v-btn>
    </v-app-bar>

    <v-main>
      <router-view/>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({
  components: {}
})
export default class App extends Vue {
  info = null
  drawer = false
  env = ''

  mounted () {
    this.env = process.env.NODE_ENV || 'development'
    console.log('mounted: ' + process.env.NODE_ENV)
  }
}
</script>

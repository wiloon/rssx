<template>
  <v-app id="inspire">
    <v-app-bar
      app
      dense
    >
      <v-toolbar-title>RSSX</v-toolbar-title>
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
    this.env = process.env.NODE_ENV
    console.log('mounted: ' + process.env.NODE_ENV)
  }
}
</script>

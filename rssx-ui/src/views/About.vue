<template>
  <v-card
    max-width="500"
    class="mx-auto"
  >
    <v-list>
      <v-list-item
        v-for="item in items"
        :key="item.Id"
        v-on:click="feedClick"
      >
        <v-list-item-content v-bind:id="item.Id">
          <v-list-item-title v-text="item.Title" v-bind:id="item.Id"></v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-card>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Axios from 'axios'

@Component({
  components: {}
})
export default class About extends Vue {
  feedClick (event: any): void {
    console.log(event.target.id)
    this.$router.push({ path: 'feed-news-list', query: { feedId: event.target.id } })
  }

  items = [
    { icon: 'true', title: 'Jason Oner', avatar: 'https://cdn.vuetifyjs.com/images/lists/1.jpg' },
    { title: 'Travis Howard', avatar: 'https://cdn.vuetifyjs.com/images/lists/2.jpg' },
    { title: 'Ali Connors', avatar: 'https://cdn.vuetifyjs.com/images/lists/3.jpg' },
    { title: 'Cindy Baker', avatar: 'https://cdn.vuetifyjs.com/images/lists/4.jpg' }
  ]

  mounted () {
    Axios
      .get('/feeds')
      .then(
        response => {
          console.log(response.data)
          this.items = response.data
        }
      )
  }
}
</script>

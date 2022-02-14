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
export default class FeedList extends Vue {
  item = 1
  items = new Map()

  feedClick (event: any): void {
    console.log(event.target.id)
    this.$router.push({ path: 'feed-news-list', query: { feedId: event.target.id } })
  }

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

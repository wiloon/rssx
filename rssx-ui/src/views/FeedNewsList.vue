<template>
  <v-container>
    <v-btn color="primary" v-on:click="markRead">Mark Read</v-btn>
    <v-card
      max-width="500"
      class="mx-auto"
    >
      <v-list>
        <v-list-item
          v-for="item in items"
          :key="item.Id"
        >
          <v-list-item-content>
            <router-link :to="{ path: 'news', query: { newsid: item.Id, feedid: item.FeedId }}">
              <v-list-item-title v-text="item.Title" v-bind:class="{ read: item.ReadFlag }"></v-list-item-title>
            </router-link>

          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>
  </v-container>

</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Axios from 'axios'

@Component({
  components: {}
})
export default class FeedNewsList extends Vue {
  item = 1
  items = new Map()
  feedId = ''

  markRead (): void {
    Axios
      .get('/mark-read',
        {
          params: { feedId: this.feedId }
        })
      .then(
        response => {
          console.log(response.data)
          this.items = response.data
        }
      )
  }

  newsClick (event: any): void {
    console.log(event.target.id)
    console.log(event.target.feedid)
    this.$router.push({ path: 'news', query: { newsid: event.target.id, feedId: event.target.feedid } })
  }

  mounted () {
    this.feedId = this.$route.query.feedId as string

    Axios
      .get('/news-list',
        {
          params: { id: this.$route.query.feedId }
        })
      .then(
        response => {
          console.log(response.data)
          this.items = response.data
        }
      )
  }
}
</script>
<style scoped lang="stylus">
.read
  color gray

.unread
  color black

a
  text-decoration none
</style>

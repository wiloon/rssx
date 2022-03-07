<template>
  <v-container>
    <v-btn v-on:click="back" style="margin-right: 10px" class="reload">Back</v-btn>
    <v-btn v-on:click="markRead" class="reload">Mark Read</v-btn>
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

  back (): void {
    this.$router.push({ name: 'FeedList' })
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
.reload
  margin-bottom 5px;
.read
  color gray

.unread
  color black

a
  text-decoration none
</style>

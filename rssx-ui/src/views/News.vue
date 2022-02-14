<template>
  <v-container
    fluid
  >
    <v-row no-gutters>
      <v-col cols="9" sm="9">
        <v-btn color="primary" v-on:click="previousNews">Previous</v-btn>
      </v-col>
      <v-col cols="3" sm="3">
        <v-btn color="primary" v-on:click="nextNews">Next</v-btn>
      </v-col>
    </v-row>

    <v-card
      class="mx-auto"
      outlined
    >
      <v-list-item three-line>
        <v-list-item-content>
          <div id="progress-div"><v-progress-linear indeterminate :active="progressActive" height="1"></v-progress-linear></div>

          <v-list-item-title class="headline mb-1" v-on:click="newsClick(items.Url)">
            {{ items.Title }}
          </v-list-item-title>
          <v-list-item-subtitle> {{ items.PubDate }}</v-list-item-subtitle>
          {{ items.Description }}
        </v-list-item-content>
      </v-list-item>
    </v-card>
    <v-row no-gutters>
      <v-col cols="9" sm="9">
        <v-btn color="primary" v-on:click="previousNews">Previous</v-btn>
      </v-col>
      <v-col cols="3" sm="3">
        <v-btn color="primary" v-on:click="nextNews">Next</v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Axios from 'axios'

@Component({
  components: {}
})
export default class News extends Vue {
  feedId = ''
  newsId = ''
  nextNewsId = ''
  previousNewsId = ''
  items = new Map()
  progressActive = false

  previousNews (): void {
    console.log('previous')
    Axios
      .get('/previous-news', {
        params: { newsId: this.newsId, feedId: this.feedId }
      })
      .then(
        response => {
          console.log(response.data)
          this.$router.replace({ path: 'news', query: { feedid: response.data.FeedId, newsid: response.data.Id } })
          this.items = response.data
          this.newsId = response.data.Id
          this.feedId = response.data.FeedId
          this.nextNewsId = response.data.NextId
        }
      )
  }

  nextNews (): void {
    console.log('next news: id: ' + this.nextNewsId)
    this.$router.replace({ path: 'news', query: { feedid: this.feedId, newsid: this.nextNewsId } })
    this.loadOneNews(this.feedId, this.nextNewsId)
  }

  mounted () {
    console.log(this.$route.query)
    this.loadOneNews(this.$route.query.feedid as string, this.$route.query.newsid as string)
  }

  loadOneNews (feedId: string, newsId: string) {
    Axios
      .get('/news', {
        params: { id: newsId, feedId: feedId }
      })
      .then(
        response => {
          console.log(response.data)
          this.items = response.data
          this.feedId = response.data.FeedId
          this.nextNewsId = response.data.NextId
          this.newsId = response.data.Id
        }
      )
  }

  newsClick (url: string): void {
    console.log('click' + url)
    this.progressActive = true
    window.location.href = url
  }
}
</script>
<style scoped lang="stylus">
#progress-div
  height 1px
</style>

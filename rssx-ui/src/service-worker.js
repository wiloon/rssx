workbox.core.setCacheNameDetails({
  prefix: 'rssx',
  suffix: 'v1'
})

workbox.core.skipWaiting()
workbox.core.clientsClaim()
workbox.precaching.precacheAndRoute(self.__precacheManifest || [])

workbox.routing.registerRoute(
  /\/api/, new workbox.strategies.NetworkOnly()
)

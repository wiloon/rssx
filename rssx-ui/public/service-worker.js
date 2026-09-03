// Self-destroying service worker.
//
// The pre-Vue-3 build shipped a Workbox PWA service worker that precached the
// whole app and served it offline-first. Removing the PWA does not unregister
// that worker from browsers that already have it — they keep serving the stale
// cached app. This file replaces it: when a stuck client checks for an update
// it gets this, which unregisters itself, clears all caches, and reloads every
// open tab so the real (network) app loads.
self.addEventListener('install', () => self.skipWaiting())

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      await self.registration.unregister()
      const keys = await caches.keys()
      await Promise.all(keys.map((k) => caches.delete(k)))
      const clients = await self.clients.matchAll({ type: 'window' })
      for (const client of clients) {
        client.navigate(client.url)
      }
    })()
  )
})

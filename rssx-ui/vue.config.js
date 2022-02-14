module.exports = {
  pwa: {
    // 一些基础配置
    name: 'RSSX',
    assetsVersion: '1.0.0',
    themeColor: '#f5f5f5',
    msTileColor: '#f5f5f5',
    appleMobileWebAppCapable: 'yes',
    appleMobileWebAppStatusBarStyle: 'debault',
    workboxPluginMode: 'InjectManifest',
    workboxOptions: {
      swSrc: 'src/service-worker.js'
    }
  },
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: true,
        pathRewrite: {
          '^/api': ''
        }
      }
    }
  },
  transpileDependencies: [
    'vuetify'
  ]
}

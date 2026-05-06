const apiProxyTarget = process.env.VUE_APP_API_PROXY_TARGET || 'http://localhost:9090'

module.exports = {
  devServer: {
    port: 8080,
    proxy: {
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        pathRewrite: {
          '^/api': '/api/v1'
        }
      }
    }
  }
}
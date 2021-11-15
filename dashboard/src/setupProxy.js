// eslint-disable-next-line @typescript-eslint/no-var-requires
const proxy = require('http-proxy-middleware')

// eslint-disable-next-line no-undef
module.exports = (app) => {
    app.use(
        proxy.createProxyMiddleware(['/api/', '/oauth', '/callback', '/logout', '/swagger', '/openapi.json'], {
            target: process.env.PROXY || 'http://127.0.0.1:7777',
        })
    )
    app.use(
        proxy.createProxyMiddleware('/ws', {
            target: process.env.PROXY || 'ws://127.0.0.1:7777',
            ws: true,
        })
    )
}

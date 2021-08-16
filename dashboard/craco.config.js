const path = require('path')
const SimpleProgressWebpackPlugin = require('simple-progress-webpack-plugin')
const CracoEsbuildPlugin = require('craco-esbuild');

module.exports = {
    plugins: [{ plugin: CracoEsbuildPlugin }],
    webpack: {
        alias: {
            '@': path.resolve(__dirname, 'src'),
        },
        plugins: [new SimpleProgressWebpackPlugin()],
        configure: (webpackConfig, { env, paths }) => {
            // https://github.com/pmndrs/react-spring/issues/1078#issuecomment-752143468
            webpackConfig.module.rules.push({
                test: /react-spring/,
                sideEffects: true,
            })
            return webpackConfig
        }
    },
}

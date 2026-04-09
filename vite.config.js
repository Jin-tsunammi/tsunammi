import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import mkcert from 'vite-plugin-mkcert'
import {nodePolyfills} from 'vite-plugin-node-polyfills'
import path from "path";

// https://vite.dev/config/
export default defineConfig({
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
            "@assets": path.resolve(__dirname, "./src/assets"),
            "@helpers": path.resolve(__dirname, "./src/helpers"),
            "@components": path.resolve(__dirname, "./src/components"),
        },
    },
    css: {
        preprocessorOptions: {
            scss: {
                // Make SCSS variables/mixins available in every <style lang="scss"> block.
                additionalData: `@use "@/assets/styles/variables" as *;`,
            },
        },
    },
    plugins: [
        vue(),
        mkcert(),
        nodePolyfills()
    ],
    optimizeDeps: {
        exclude: ['react', 'react-dom']
    },
    build: {
        rollupOptions: {
            external: ['react', 'react-dom']
        }
    }
})

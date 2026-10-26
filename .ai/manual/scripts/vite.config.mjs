import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '../../../cm_collectors_html');
const require = createRequire(path.join(root, 'package.json'));
const vue = (await import(pathToFileURL(require.resolve('@vitejs/plugin-vue')))).default;
const vueJsx = (await import(pathToFileURL(require.resolve('@vitejs/plugin-vue-jsx')))).default;
export default {
 root, plugins: [vue(), vueJsx()],
 cacheDir: path.resolve(here, '../.local/vite-cache'),
 resolve: { alias: { '@': path.join(root, 'src') } },
 optimizeDeps: { entries: ['index.html'] },
 server: { host: '127.0.0.1', port: 18346, strictPort: true,
   proxy: { '/api': { target: 'http://127.0.0.1:18345', changeOrigin: true } } },
};

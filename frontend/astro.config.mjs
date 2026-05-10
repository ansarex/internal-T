import { defineConfig } from 'astro/config';
import node from '@astrojs/node';
import vue from '@astrojs/vue';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  output: 'server',
  adapter: node({
    mode: 'standalone',
  }),
  integrations: [
    vue({
      appEntrypoint: '/src/app.ts',
    }),
    tailwind(),
  ],
  vite: {
    optimizeDeps: {
      include: ['pinia', 'axios'],
    },
  },
});

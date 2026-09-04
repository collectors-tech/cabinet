import path from 'path'
import { fileURLToPath } from 'url'
import { defineConfig } from 'cypress'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  component: {
    specPattern: 'cypress/component/**/*.cy.{ts,tsx}',
    supportFile: 'cypress/support/component.ts',
    devServer: {
      framework: 'react',
      bundler: 'vite',
      viteConfig: {
        plugins: [react(), tailwindcss()],
        resolve: {
          alias: {
            '@': path.resolve(__dirname, './src'),
          },
        },
      },
    },
  },
  video: false,
  screenshotOnRunFailure: true,
  screenshotsFolder: 'cypress/artifacts/screenshots',
  videosFolder: 'cypress/artifacts/videos',
  fixturesFolder: 'cypress/fixtures',
})

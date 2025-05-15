// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['@nuxt/eslint', '@nuxt/ui-pro', '@vueuse/nuxt', '@pinia/nuxt'],

  ssr: false,

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiEndpoint:
        import.meta.env.NUXT_API_ENDPOINT ?? 'http://localhost:8080/api'
    }
  },

  routeRules: {
    '/api/**': {
      cors: true
    }
  },

  future: {
    compatibilityVersion: 4
  },

  compatibilityDate: '2024-07-11',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})

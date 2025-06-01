<script setup lang="ts">
const route = useRoute()
</script>

<template>
  <div class="h-(--ui-header-height) shrink-0 flex lg:hidden items-center justify-between border-b border-default px-4 sm:px-6 gap-1.5">
    <div class="flex items-center gap-1.5 min-w-0">
      <USlideover side="left">
        <UButton
          color="neutral"
          variant="soft"
          icon="i-lucide:menu"
          class="mr-3"
        />

        <template #content>
          <NavigationMain class="!flex !w-full" />
        </template>
      </USlideover>

      <h1 class="flex items-center gap-1.5 font-semibold text-highlighted truncate">
        {{ route.meta.title }}
      </h1>
    </div>
  </div>
</template>

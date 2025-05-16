<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <div class="relative flex flex-col flex-1 lg:grid gap-x-8 gap-y-4 lg:grid-cols-2 lg:items-center p-0 sm:p-0 mb-4">
    <div class="flex flex-col flex-1 items-start">
      <div class="flex-1">
        <div class="text-base text-pretty font-semibold text-highlighted">
          {{ title }}
        </div>

        <div class="text-sm text-pretty text-muted mt-1">
          {{ description }}
        </div>
      </div>
    </div>

    <slot />
  </div>
</template>

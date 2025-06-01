<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <UCard
    variant="subtle"
  >
    <div class="mb-4">
      <div class="text-base text-pretty font-semibold text-highlighted">
        {{ title }}
      </div>
      <div class="text-sm text-pretty text-muted mt-1">
        {{ description }}
      </div>
    </div>

    <slot />
  </UCard>
</template>

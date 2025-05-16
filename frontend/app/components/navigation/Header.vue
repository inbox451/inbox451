<script setup lang="ts">
defineProps<{
  title?: string
}>()

const route = useRoute()
</script>

<template>
  <div class="h-(--ui-header-height) shrink-0 hidden lg:flex items-center justify-between border-b border-default px-4 sm:px-6 gap-1.5">
    <div class="flex items-center gap-1.5 min-w-0">
      <slot name="leading" />

      <h1 class="font-semibold text-highlighted truncate">
        {{ title ?? route.meta.title }}
      </h1>

      <slot name="trailing" />
    </div>

    <div class="flex items-center shrink-0 gap-1.5">
      <slot name="right" />
    </div>
  </div>
</template>

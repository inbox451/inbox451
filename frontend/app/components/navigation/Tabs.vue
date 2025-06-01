<script setup lang="ts">
import type { NavigationMenuProps } from '#ui/types'

defineProps<{
  links: NavigationMenuProps['items']
}>()
</script>

<template>
  <div class="shrink-0 flex items-center justify-between border-b border-default px-4 sm:px-6 gap-1.5 overflow-x-auto">
    <UNavigationMenu :items="links" highlight class="-mx-1 flex-1 " />
  </div>
</template>

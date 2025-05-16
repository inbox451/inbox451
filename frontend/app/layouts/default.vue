<script setup lang="ts">
</script>

<template>
  <div class="fixed inset-0 lg:flex-row flex-col flex overflow-hidden">
    <NavigationMain />
    <NavigationMobile />
    <slot />
  </div>
</template>

<script setup>
defineProps({
  msg: String,
})
</script>

<template>
  <h1>{{ msg }}</h1>
  <p>Where test emails go up in flames 🔥</p>
</template>

<style scoped>
.card {
  padding: 2em;
}
</style>

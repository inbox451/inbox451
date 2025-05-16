<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { breakpointsTailwind } from '@vueuse/core'
import type { Messages } from '~/types'

definePageMeta({
  title: 'Inbox'
})

const route = useRoute()
const nuxtApp = useNuxtApp()
const userStore = useUserStore()

const tabItems = [{
  label: 'All',
  value: 'all'
}, {
  label: 'Unread',
  value: 'unread'
}]

const selectedTab = ref('all')
const selectedMail = ref<Messages | null>()

const { data: mails, refresh } = await useLazyAsyncData(() => {
  if (!userStore.selectedProject?.id) return Promise.resolve([])

  return nuxtApp.$api.getMessages(userStore.selectedProject.id, String(route.params.id)).then((res) => {
    return res.data
  })
})

// Filter mails based on the selected tab
const filteredMails = computed(() => {
  if (selectedTab.value === 'unread') {
    return mails.value?.filter(mail => !mail.is_read) || []
  }

  return mails.value || []
})

const isMailPanelOpen = computed({
  get() {
    return !!selectedMail.value
  },
  set(value: boolean) {
    if (!value) {
      selectedMail.value = null
    }
  }
})

// Reset selected mail if it's not in the filtered mails
watch(filteredMails, () => {
  if (!filteredMails.value || !selectedMail.value) return

  if (!filteredMails.value.find(mail => mail.id === selectedMail.value?.id)) {
    selectedMail.value = null
  }
})

// Mark mail as read when selected
watch(selectedMail, async (newMail) => {
  if (!userStore.selectedProject?.id) return

  if (newMail && !newMail?.is_read) {
    await nuxtApp.$api.markAsRead(userStore.selectedProject?.id, String(route.params.id), newMail.id)
    await refresh()
  }
})

// When mails are refreshed, reselect the selected mail as they might have changed
watch(mails, () => {
  if (!selectedMail.value) return

  const mail = mails.value?.find(mail => mail.id === selectedMail.value?.id)
  if (mail) {
    selectedMail.value = mail
  } else {
    selectedMail.value = null
  }
})

// Refresh mails every 5 seconds
const interval = setInterval(async () => {
  await refresh()
}, 5000)

onBeforeUnmount(() => {
  clearInterval(interval)
})

const breakpoints = useBreakpoints(breakpointsTailwind)
const isMobile = breakpoints.smaller('lg')
</script>

<template>
  <div class="dashboard-panel shrink-0 w-full lg:w-[25%]">
    <NavigationHeader>
      <template #trailing>
        <UBadge :label="filteredMails?.length" variant="subtle" />
      </template>
      <template #right>
        <UTabs
          v-model="selectedTab"
          :items="tabItems"
          class="w-32"
          :content="false"
          size="xs"
        />
      </template>
    </NavigationHeader>

    <InboxList v-model="selectedMail" :mails="filteredMails" />
  </div>

  <InboxMail
    v-if="selectedMail"
    :mail="selectedMail"
    @close="selectedMail = null"
    @refresh="refresh"
  />
  <div v-else class="hidden lg:flex flex-1 items-center justify-center">
    <UIcon name="i-lucide-inbox" class="size-32 text-dimmed" />
  </div>

  <USlideover v-if="isMobile" v-model:open="isMailPanelOpen">
    <template #content>
      <InboxMail v-if="selectedMail" :mail="selectedMail" @close="selectedMail = null" />
    </template>
  </USlideover>
</template>

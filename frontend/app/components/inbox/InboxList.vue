<script setup lang="ts">
import { format, isToday } from 'date-fns'
import type { Messages } from '~/types'

const props = defineProps<{
  mails: Messages[]
}>()

const mailsRefs = ref<Element[]>([])

const selectedMail = defineModel<Messages | null>()

watch(selectedMail, () => {
  if (!selectedMail.value) {
    return
  }
  const ref = mailsRefs.value[selectedMail.value.id]
  if (ref) {
    ref.scrollIntoView({ block: 'nearest' })
  }
})

defineShortcuts({
  arrowdown: () => {
    const index = props.mails.findIndex(mail => mail.id === selectedMail.value?.id)

    if (index === -1) {
      selectedMail.value = props.mails[0]
    } else if (index < props.mails.length - 1) {
      selectedMail.value = props.mails[index + 1]
    }
  },
  arrowup: () => {
    const index = props.mails.findIndex(mail => mail.id === selectedMail.value?.id)

    if (index === -1) {
      selectedMail.value = props.mails[props.mails.length - 1]
    } else if (index > 0) {
      selectedMail.value = props.mails[index - 1]
    }
  }
})
</script>

<template>
  <div class="overflow-y-auto divide-y divide-default">
    <div
      v-for="(mail, index) in mails"
      :key="index"
      :ref="el => { mailsRefs[mail.id] = el as Element }"
    >
      <div
        class="p-4 sm:px-6 text-sm cursor-pointer border-l-2 transition-colors"
        :class="[
          mail.is_read ? 'text-toned' : 'text-highlighted',
          selectedMail && selectedMail.id === mail.id ? 'border-primary bg-primary/10' : 'border-(--ui-bg) hover:border-primary hover:bg-primary/5'
        ]"
        @click="selectedMail = mail"
      >
        <div class="flex items-center justify-between" :class="[!mail.is_read && 'font-semibold']">
          <div class="flex items-center gap-3">
            {{ mail.sender }}

            <UChip v-if="!mail.is_read" />
          </div>

          <span>{{ isToday(new Date(mail.created_at)) ? format(new Date(mail.created_at), 'HH:mm') : format(new Date(mail.created_at), 'dd MMM') }}</span>
        </div>
        <p class="truncate" :class="[!mail.is_read && 'font-semibold']">
          {{ mail.subject }}
        </p>
        <p class="text-dimmed line-clamp-1">
          {{ mail.body }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  title: 'Users'
})

const nuxtApp = useNuxtApp()
// const toast = useToast()

const { data: users } = await useLazyAsyncData('users', () => nuxtApp.$api.getUsers().then((res) => {
  return res.data
}))

const newEmail = ref('')
const isLoading = ref(false)
const isModalOpen = ref(false)

const q = ref('')

const filteredUsers = computed(() => {
  if (!users.value) return []

  return users.value.filter((user) => {
    return user.name.search(new RegExp(q.value, 'i')) !== -1 || user.email.search(new RegExp(q.value, 'i')) !== -1
  })
})

async function inviteUser() {
  console.log('inviteUser', newEmail.value)
  // await nuxtApp.$api.createProject(newEmail.value).then(async () => {
  //   toast.add({ title: 'Success', description: 'New project has been created.', color: 'success' })

  //   await refreshProjects()
  // })
  //   .catch((error) => {
  //     toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
  //   })
  //   .finally(() => {
  //     isModalOpen.value = false
  //   })

  // newEmail.value = ''
  // isLoading.value = false
}
</script>

<template>
  <div>
    <UiCardHeader
      title="Users"
      description="Invite new users by email address."
    >
      <UModal v-model:open="isModalOpen" title="Invite User">
        <UButton
          label="Invite User"
          color="neutral"
          class="w-fit lg:ms-auto"
        />

        <template #body>
          <UInput
            v-model="newEmail"
            placeholder="Enter new email address"
            class="w-full "
          />
          <UButton
            label="Invite"
            color="neutral"
            :disabled="!newEmail"
            class="inline-flex items-center justify-center w-full mt-4"
            :loading="isLoading"
            @click="inviteUser"
          />
        </template>
      </UModal>
    </UiCardHeader>

    <UCard
      variant="subtle"
      :ui="{
        body: 'p-0 sm:p-0'
      }"
    >
      <SettingsUsersList :data="filteredUsers" />
    </UCard>
  </div>
</template>

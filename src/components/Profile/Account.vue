<template>
  <div :class="['account', {edit: isEditMode}]">
    <UIBaseInput v-model="email" :is_readonly="!isEditMode" size="large" label="Email" :error-message="emailError"/>
    <div class="account__btns">
      <UIButton
        v-if="!isEditMode"
        color_type="ghost"
        size="large"
        @cta="openEditMode"
      >
        <template #left-icon>
          <SVGEdit />
        </template>
        Edit
      </UIButton>
      <div v-else class="account__edit">
        <UIButton
          color_type="primary"
          size="large"
          @cta="handleEmailChange('save')"
        >
          Save
        </UIButton>
        <UIButton
          color_type="ghost"
          size="large"
          @cta="handleEmailChange('cancel')"
        >
          Cancel
        </UIButton>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIBaseInput from "../UI/UIBaseInput.vue";
import UIButton from "../UI/UIButton.vue";
import SVGEdit from "../SVG/SVGEdit.vue";
import {ref, watch} from "vue";
import {useUserStore} from "../../store/userStore.js";
const props = defineProps({isEmailExists: {type: Boolean, default: false}});
const emits = defineEmits(['openModal', 'update:email'])
const userStore = useUserStore();
const isEditMode = ref(false);
const email = ref('');
const emailError = ref('');

const openEditMode = () => isEditMode.value = true;

const handleEmailChange = (action) => {
  if (action === "cancel") {
    email.value = userStore.userData?.email || '';
    isEditMode.value = false;
  } else if (action === 'save') {
    const emailPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    if (email.value === userStore.userData?.email) {
      emailError.value = 'Enter new email'
    } else if (!emailPattern.test(email.value)) {
      emailError.value = 'Invalid email';
    } else {
      emailError.value = '';
      emits('openModal')
    }
  }
}

watch(() => userStore.userData, (newVal) => {
  if (newVal) {
    email.value = newVal.email;

    if (isEditMode.value) {
      isEditMode.value = false;
    }
  }
}, {deep: true});

watch(() => email.value, (newVal) => {
  emits('update:email', newVal);
})

watch(() => props.isEmailExists, (newVal) => {
  if (newVal) {
    emailError.value = 'This email already exists.';
  }
})
</script>
<style scoped lang="scss">
.account {
  display: flex;
  align-items: flex-end;
  gap: 20px;

  & .base-input {
    max-width: 288px;
  }

  &__btns {
    margin-top: 27px;
    display: flex;
    align-items: center;
    gap: 10px;
    align-self: flex-start;
  }

  &__edit {
    display: flex;
    align-items: center;
    gap: 10px;
  }
}

@media (max-width: 1200px) {
  .account {
    gap: 0;

    & .base-input {
      max-width: none;
    }

    &.edit {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;

      & .account {
        &__btns {
          width: 100%;
          display: flex;
          align-items: center;
          gap: 10px;
        }

        &__edit {
          width: 100%;

          & .ui-button {
            width: 100%;
          }
        }
      }
    }
  }
}
</style>
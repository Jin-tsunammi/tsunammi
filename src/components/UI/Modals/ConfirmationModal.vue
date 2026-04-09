<template>
  <div class="confirmation-modal">
    <slot v-if="isCustomContent" name="confirmation-custom-content"/>
    <template v-else>
      <div :class="['heading-4', headerColor]">{{modalStore.modalData.title}}</div>
      <span class="paragraph-small regular grey">{{mainText}}</span>
      <span class="paragraph-small regular grey addition">{{additionalText}}</span>
      <div :class="['confirmation-modal__btns', {disabled: isLoading}]">
        <UIButton
          v-if="isCancel"
          class="cancel"
          color_type="outline"
          @cta="modalStore.closeModal"
        >
          {{cancellationBtnText}}
        </UIButton>
        <UIButton
          :color_type="confirmationBtnStyle"
          @cta="emits('handleConfirmation')"
        >
          <template #left-icon>
            <UISpinner
              v-if="isLoading"
            />
          </template>
          {{confirmationBtnText}}
        </UIButton>
      </div>
    </template>
  </div>
</template>
<script setup>
import UIButton from "../UIButton.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import UISpinner from "../UISpinner.vue";

defineProps({
  headerColor: {type: String, default: "default"},
  additionalText: {type: String, default: ""},
  mainText: {type: String, default: ""},
  confirmationBtnText: {type: String, default: ""},
  cancellationBtnText: {type: String, default: "Cancel"},
  confirmationBtnStyle: {type: String, default: "primary"},
  isLoading: {type: Boolean, default: false},
  isCancel: {type: Boolean, default: true},
  isCustomContent: {type: Boolean, default: false},
})
const emits = defineEmits(["handleConfirmation"]);
const modalStore = useModalsStore();
</script>
<style scoped lang="scss">
.confirmation-modal {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  border-radius: 10px;
  border: 1px solid #E5E7EB;
  background: #FFF;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -4px rgba(0, 0, 0, 0.10);
  padding: 32px;
  max-width: 320px;

  & .heading-4 {
    margin-bottom: 16px;
    text-align: center;

    &.success {
      color: #16A34A;
    }
  }

  & span {
    text-align: center;
  }

  & .addition {
    display: block;
    margin-top: 8px;
  }

  &__btns {
    width: 100%;
    margin-top: 16px;
    display: flex;
    align-items: center;
    gap: 8px;

    & .ui-button {
      width: 100%;
    }

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }
  }
}
</style>
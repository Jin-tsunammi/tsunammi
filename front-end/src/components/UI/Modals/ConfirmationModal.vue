<template>
  <div :class="['confirmation-modal', headerColor]">
    <slot v-if="isCustomContent" name="confirmation-custom-content"/>
    <template v-else>
      <div v-if="headerColor && headerColor !== 'default'" class="status">
        <SVGSuccessCircle v-if="headerColor === 'success'" />
        <SVGTriangleWarning v-if="headerColor === 'error'" />
      </div>
      <div :class="['heading-4']">{{modalStore.modalData.title}}</div>
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
          v-if="isConfirmationBtn"
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
import SVGTriangleWarning from "../../SVG/SVGTriangleWarning.vue";
import SVGSuccessCircle from "../../SVG/SVGSuccessCircle.vue";

defineProps({
  headerColor: {type: String, default: "default"},
  additionalText: {type: String, default: ""},
  mainText: {type: String, default: ""},
  confirmationBtnText: {type: String, default: ""},
  cancellationBtnText: {type: String, default: "Cancel"},
  confirmationBtnStyle: {type: String, default: "primary"},
  isLoading: {type: Boolean, default: false},
  isCancel: {type: Boolean, default: true},
  isConfirmationBtn: {type: Boolean, default: true},
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

  &.success {
    & .heading-4 {
      color: #16A34A;
    }

    & .status {
      background: rgba(22, 163, 74, 0.13);
    }
  }

  &.error {
    & .heading-4 {
      color: #DC2626;
    }

    & .status {
      background: rgba(220, 38, 38, 0.13);
    }
  }

  & .status {
    width: 38px;
    height: 38px;
    border-radius: 6.154px;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    justify-content: center;

    & svg {
      width: 50%;
      height: 50%;
    }
  }

  & .heading-4 {
    margin-bottom: 16px;
    text-align: center;
  }

  & span {
    text-align: center;
  }

  & .addition {
    display: block;
    margin-top: 8px;
    white-space: pre-line;
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
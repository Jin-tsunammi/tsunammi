<template>
  <div class="modal-profile-code">
    <div class="modal-profile-code__icon">
      <SVGEmail />
    </div>
    <span class="modal-profile-code__title heading-4">{{modalStore.modalData.title}}</span>
    <p class="paragraph-small">We sent a 6-digit verification code to <br> <span>{{email}}</span></p>
    <CodeStage
      ref="codeStageRef"
      :email="email"
      @handle-stage-change="modalStore.closeModal"
      @handle-code-request="handleCodeRequest"
      @update:code="code = $event"
    />
    <div class="modal-profile-code__btns">
      <UIButton :color_type="'primary'" :is_disabled="code?.length !== 6" size="large" @cta="handleCodeSubmit">
        Confirm
      </UIButton>
      <UIButton :color_type="'outline'" size="large" @cta="modalStore.closeModal">
        Cancel
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import CodeStage from "../../Login/CodeStage.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import {GetCodeByEmail} from "../../../api/api.js";
import {useToastStore} from "../../../store/toastStore.js";
import UIButton from "../../UI/UIButton.vue";
import {ref} from "vue";
import SVGEmail from "../../SVG/SVGEmail.vue";

const props = defineProps({
  email: {type: String, default: ''},
})
const modalStore = useModalsStore();
const toastStore = useToastStore();
const code = ref('');
const codeStageRef = ref(null);

const handleCodeRequest = async() => {
  try {
    await GetCodeByEmail({email: props.email});

  } catch (error) {
    console.error(error);
    toastStore.error({text: "Something went wrong"});
  }
}

const handleCodeSubmit = () => {
  codeStageRef.value.handleCodeSubmit();
}
</script>
<style scoped lang="scss">
.modal-profile-code {
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  background: #FFF;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
  max-width: 400px;
  padding: 40px;
  display: flex;
  flex-direction: column;

  &__icon {
    margin: 0 auto;
    aspect-ratio: 1/1;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 40px;
    max-width: 40px;
    border-radius: 6.154px;
    background: rgba(234, 88, 12, 0.13);
  }

  &__title {
    margin: 12px auto;
  }

  & p {
    color: #6B7280;
    font-weight: 400;
    text-align: center;

    ::v-deep(span) {
      color: #111827;
    }
  }

  &__btns {
    margin-top: 28px;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
}
</style>
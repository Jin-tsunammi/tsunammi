<template>
  <div class="private-key-modal">
    <p class="paragraph-small regular grey">Your private key gives full access to your wallet. Store it securely and never share it.</p>
    <UIBaseInput :is_readonly="true" v-model="inputVal" size="large"/>
    <div class="private-key-modal__btns">
      <UIButton color_type="outline" size="large" @cta="modalStore.closeModal">Cancel</UIButton>
      <UIButton color_type="primary" size="large" @cta="handleTextCopy">{{copied ? 'Copied' : 'Copy'}}</UIButton>
    </div>
  </div>
</template>
<script setup>
import UIBaseInput from "../../UI/UIBaseInput.vue";
import UIButton from "../../UI/UIButton.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import {useClipboard} from "@vueuse/core";

const modalStore = useModalsStore();
const inputVal = defineModel({type: String, default: ''});

const {copy, copied, isSupported} = useClipboard();

const handleTextCopy = async () => {
  if (!isSupported) {
    alert('Copy does not supported');

    return;
  }

  if (!inputVal.value) return;

  await copy(inputVal.value);
}
</script>
<style scoped lang="scss">
.private-key-modal {
  width: 350px;

  padding: 0 24px 16px;

  & p {
    margin-bottom: 16px;
  }

  &__btns {
    margin-top: 20px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
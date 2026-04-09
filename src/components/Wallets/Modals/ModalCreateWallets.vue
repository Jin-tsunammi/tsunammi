<template>
  <div class="create-wallets">

    <span class="create-wallets__top paragraph-small regular grey">{{`Generate new wallets for ${selectedProject.name || ''} project.`}}</span>

    <div class="create-wallets__qnty">
      <UIBaseInput
        v-model="walletsQnty"
        :error-message="walletsQntyError"
        label="Number of wallets"
        placeholder="1 - 1000"
        type="number"
        :max-length="1000"
        size="large"
        @handle-input="handleWalletsInput($event.target.value)"
      >
        <template #bottom-left >
          <span class="paragraph-mini regular grey">min: 1  |  max: 1000</span>
        </template>
      </UIBaseInput>
    </div>

    <div :class="['create-wallets__btns', {disabled: isGenerating}]">
      <UIButton
        color_type="outline"
        @cta="modalsStore.closeModal"
      >
        Cancel
      </UIButton>
      <UIButton color_type="primary" @cta="handleImport">
        {{isGenerating ? 'Generating...' : 'Generate'}}
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../../UI/UIButton.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import UIBaseInput from "../../UI/UIBaseInput.vue";
import {computed, ref, watch} from "vue";
import {useProjectsStore} from "../../../store/projectsStore.js";
import {useToastStore} from "../../../store/toastStore.js";
import {GenerateSolWallets} from "../../../api/api.js";
import {useRoute} from "vue-router";
import {errorToast} from "../../../helpers/index.js";

const route = useRoute();
const modalsStore = useModalsStore();
const projectsStore = useProjectsStore();
const toastStore = useToastStore();
const walletsQnty = ref(null);
const walletsQntyError = ref('');
const isGenerating = ref(false);
const selectedProject = computed(() => {
  if (modalsStore.modalData.item) {
    return modalsStore.modalData.item;
  } else if (projectsStore.selectedProject) {
    return projectsStore.selectedProject;
  }
});

const handleWalletsInput = (value) => {
  const digitsOnly = value.replace(/\D+/g, '')

  let num = parseInt(digitsOnly, 10)
  if (!isNaN(num)) {
    if (num < 1) num = 1
    if (num > 1000) num = 1000
    walletsQnty.value = num
  } else {
    walletsQnty.value = null
  }
}

const checkFields = () => {
  let status = true;
  if (!walletsQnty.value || !selectedProject.value) status = false;

  if (!walletsQnty.value) {
    walletsQntyError.value = 'Enter a number from 1 to 1000'
  }

  return status;
}

const handleImport = async() => {
  if (!checkFields()) return;

  try {
    isGenerating.value = true;
    await GenerateSolWallets({count: walletsQnty.value, project_ids: [selectedProject.value?.id]});
    await projectsStore.updateProjectData(selectedProject.value?.id, false);
    toastStore.success({text: "Wallets have been generated"});
    modalsStore.closeModal();
  } catch (error) {
    console.error(error);
    errorToast(error.response.data)
  } finally {
    isGenerating.value = false;
  }
}
</script>
<style scoped lang="scss">

.create-wallets {
  width: 360px;
  padding: 0 20px 16px 20px;
  
  &__top {
    display: block;
    margin-bottom: 16px;
  }

  &__btns {
    margin-top: 16px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }
  }
}

@media (max-width: 1200px) {
  .create-wallets {
    width: 100%;
  }
}
</style>
<template>
  <div class="project-create-modal">
    <div class="top">
      <div class="project-create-modal__image">
        <div class="avatar">
          <DefaultAvatar />
        </div>
      </div>
      <UIBaseInput label="Wallet Pool's name" size="large" v-model="projectName" placeholder="Enter Pool's name..." />
    </div>
    <div :class="['project-create-modal__btns', {disabled: isInProgress}]">
      <UIButton color_type="outline" size="large" @cta="modalsStore.closeModal">Cancel</UIButton>
      <UIButton
        color_type="primary"
        size="large"
        @cta="handleProjectCreate"
        :is_disabled="isContinueBtnDisabled"
      >
        <template #left-icon>
          <UISpinner v-if="isInProgress" />
          {{createText}}
        </template>
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import UIBaseInput from "../../UI/UIBaseInput.vue";
import UIButton from "../../UI/UIButton.vue";
import {computed, ref, watch} from "vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import {useProjectsStore} from "../../../store/projectsStore.js";
import SVGImage from "../../SVG/SVGImage.vue";
import SVGPlus from "../../SVG/SVGPlus.vue";
import SVGEdit from "../../SVG/SVGEdit.vue";
import UISpinner from "../../UI/UISpinner.vue";
import DefaultAvatar from "../../UI/DefaultAvatar.vue";

const modalsStore = useModalsStore();
const projectsStore = useProjectsStore();
const DEFAULT_TEXT = 'JPG, JPEG or PNG. Max size of 800K';
const projectName = ref('');
const projectNameError = ref('');
const isInProgress = ref(false);
const imageRef = ref(null);
const imageText = ref(DEFAULT_TEXT);
const imageFile = ref({
  file: null,
  address: '',
});
const isContinueBtnDisabled = computed(() => {
  const project = modalsStore.modalData.item;
  const enteredName = projectName.value.trim();

  if (project) {
    if (!enteredName.length) return true;

    else return project.name === enteredName;
  } else {
    return !enteredName.length;
  }
})
const createText = computed(() => {
  if (isInProgress.value) {
    if (modalsStore.modalData.item) return 'Saving...';
    else return 'Creating...'
  } else {
    if (modalsStore.modalData.item) return 'Save';
    else return 'Create'
  }
})

const handleProjectCreate = async () => {
  isInProgress.value = true;
  try {
    if (modalsStore.modalData.item) {
      if (modalsStore.modalData.item.name === projectName.value) {
        projectNameError.value = 'Enter new name';
        isInProgress.value = false;
      } else {
        await projectsStore.updateProjectName(modalsStore.modalData.item?.id, projectName.value);
        modalsStore.closeModal()
      }
    } else {
      const project = await projectsStore.handleProjectCreate({name: projectName.value});

      if (project) {
        modalsStore.modalData.title = `Wallet Pool ${project.name || ''} created successfully`
        modalsStore.modalData.type = 'create-confirmation'
        modalsStore.modalData.action = 'confirmation'
        modalsStore.modalData.item = project;
        modalsStore.modalData.is_poen = true;
      }
    }
  } finally {
    isInProgress.value = false;
  }
}

const openImageInput = () => {
  if (!imageRef.value) return;

  imageRef.value.click();
}

const uploadImage = (event) => {
  const file = event.target.files[0];
  if (file) {
    if (file.size > 800 * 1024) {
      imageText.value = 'Image size it too large. Max 800K';
      return;
    }

    if (!['image/png', 'image/jpeg', 'image/jpg'].includes(file.type)) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      imageFile.value.file = file;
      imageFile.value.address = e.target.result;

      imageText.value = DEFAULT_TEXT;
      if (imageFile.value) {
        imageFile.value.value = ''
      }
    };
    reader.readAsDataURL(file);
  }
}

watch(() => modalsStore.modalData.item, (val) => {
  if (val) {
    projectName.value = val.name;
  }
}, {immediate: true})
</script>
<style scoped lang="scss">
.project-create-modal {
  width: 350px;

  & .top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__image {
    display: flex;
    align-items: center;
    gap: 24px;
    //margin-bottom: 24px;

    & .avatar {
      display: flex;
      align-items: center;
      justify-content: center;
      aspect-ratio: 1/1;
      min-width: 64px;
      max-width: 64px;
      border-radius: 50%;
    }

    & .image {
      aspect-ratio: 1/1;
      min-width: 64px;
      max-width: 64px;
      border-radius: 50%;
      position: relative;
      display: flex;
      background: #F3F4F6;
      align-items: center;
      justify-content: center;

      &-inner {
        display: flex;
        background: #F3F4F6;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        border-radius: 50%;
      }

      & button {
        position: absolute;
        z-index: 2;
        background: #FFF;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        bottom: 0;
        right: 0;
        border: 1px solid #D1D5DB;
        aspect-ratio: 1/1;
        min-width: 20px;
        max-width: 20px;
        padding-left: 1px;

        & svg {
          width: 60%;
          height: 60%;
        }
      }

      & input {
        position: absolute;
        pointer-events: none;
        visibility: hidden;
      }
    }
  }

  &__btns {
    margin-top: 24px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    width: 100%;

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }
  }
}
</style>
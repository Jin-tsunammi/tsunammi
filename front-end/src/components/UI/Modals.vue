<template>
  <teleport to="#modals">
    <div v-if="modalsStore.modalData.is_open" class="modal">
      <Transition name="show-up">
        <div v-if="isTransitionVisible" class="modal__content">
          <div v-if="modalsStore.modalData.action === 'configure'" class="modal__default">
            <div class="modal__default_top">
              <slot v-if="modalsStore.modalData.icon" name="title-icon-left" />
              <div class=" modal__title heading-4">{{ modalsStore.modalData.title }}</div>
              <button v-if="modalsStore.modalData.is_close_icon" @click="modalsStore.closeModal()">
                <SVGClose/>
              </button>
            </div>
            <slot v-if="modalsStore.modalData.is_custom" name="custom-content" />
            <div v-else class="modal__default_content">
              <slot/>
            </div>
          </div>

          <div v-if="modalsStore.modalData.action === 'confirmation'" class="modal__confirmation">
            <slot/>
          </div>

          <div v-if="modalsStore.modalData.action === 'login'" class="modal__login">
            <LoginModal />
          </div>
        </div>
      </Transition>
    </div>
  </teleport>
</template>
<script setup>
import SVGClose from "../SVG/SVGClose.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import {ref, watch} from "vue";
import LoginModal from "../Login/LoginModal.vue";

const modalsStore = useModalsStore();
const isTransitionVisible = ref(false);

watch(() => modalsStore.modalData.is_open, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      isTransitionVisible.value = true;
    }, 50);
  } else {
    setTimeout(() => {
      isTransitionVisible.value = false;
    }, 50);
  }
})

</script>
<style scoped lang="scss">
@import "../../assets/styles/main.scss";

.modal {
  position: fixed;
  top: 0;
  right: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100vw;
  min-height: 100vh;
  height: 100%;
  background: rgba(0, 0, 0, 0.50);
  backdrop-filter: blur(5px);
  z-index: 10000;

  &__content {
    display: flex;
    width: fit-content;

    &:has(.modal__login) {
      height: 100%;
    }
  }

  &__default {
    width: 100%;
    border-radius: 10px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -4px rgba(0, 0, 0, 0.10);

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 16px 24px;

      & .heading-4 {
        margin-right: auto;
        color: #030712;
      }

      & button {
        margin-left: auto;
        background: transparent;
        min-width: 16px;
        min-height: 16px;
        display: flex;
        align-items: center;
        justify-content: center;

        &:hover {
        }

        &:active {
        }
      }
    }

    &_content {
      padding: 24px 24px 16px;
    }
  }

  &__login {
    width: 687px;
    height: 100%;
  }
}

@media (max-width: 1200px) {
  .modal {
    width: 100vw;

    &__content {
      width: 100%;
    }
  }
}
</style>
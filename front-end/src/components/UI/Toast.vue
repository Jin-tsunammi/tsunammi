<template>
  <Transition name="toast">
    <div class="toaster__wrapper">
      <TransitionGroup name="toast" tag="ul">
        <div
          v-for="toast in toastStore.toasts"
          :class="['toaster__inner', toast.status]"
          :key="toast.text"
        >
          <div :class="['toaster__icon', {top: toast.second_text}]">
            <component v-if="toastIconMap[toast.status]" class="toaster__inner-icon" :is="toastIconMap[toast.status]"/>
          </div>

          <div class="toaster__inner-info">
            <span class="toaster__inner-text paragraph-small" v-html="toast.text"></span>
            <span v-if="toast.second_text" class="toaster__inner-second-text paragraph-small" v-html="toast.second_text"></span>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Transition>
</template>
<script setup>
import {useToastStore} from "../../store/toastStore.js";
import SVGChecked from "../SVG/SVGChecked.vue";
import SVGWarning from "../SVG/SVGWarning.vue";

const toastStore = useToastStore();

const toastIconMap = {
  error: SVGWarning,
  success: SVGChecked,
  info: null,
};
</script>

<style scoped lang="scss">
.toast-enter-from,
.toast-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.toast-enter-active,
.toast-leave-active {
  transition: 0.25s ease all;
}

.toaster {
  &__wrapper {
    position: fixed;
    top: 90px;
    right: 32px;
    width: fit-content;
    z-index: 15000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 0;

    ::v-deep(ul) {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
  }

  &__icon {
    min-width: 16px;
    min-height: 16px;
    max-width: 16px;
    max-height: 16px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    &.top {
      margin-top: 4px;
      align-self: flex-start;
    }
  }

  &__inner {
    border-radius: 8px;
    background: #FFF;
    position: relative;
    overflow: hidden;
    width: 400px;

    border: 1px solid #E5E7EB;

    color: #FFF;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;

    &.success {
      & .toaster {
        &__inner {
          &-text {
            color: #16A34A;
          }
          &-icon {
            ::v-deep(path) {
              stroke: #16A34A;
            }
          }
        }
      }
    }

    &.error {
      .toaster {
        &__inner {
          &-text, &-second-text {
            color: #DC2626;
          }
        }
      }
    }

    &-icon {
      min-width: 16px;
      max-width: 16px;
      aspect-ratio: 1/1;
    }

    &-text {
      color: #030712;
      font-weight: 500;
    }

    &-info {
      display: flex;
      flex-direction: column;
    }

    &-second-text {
      color: #6B7280;
      font-weight: 400;
    }
  }
}
</style>

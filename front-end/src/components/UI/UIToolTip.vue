<template>
  <div v-if="isShown" class="ui-tooltip__wrapper">
    <Transition name="show-up">
      <div v-if="isTextVisible" :class="['ui-tooltip', position]">
        <slot name="svg"/>
        <span>{{text}}</span>
      </div>
    </Transition>
  </div>
</template>
<script setup>
import {ref, watch} from "vue";

const props = defineProps({
  isShown: {type: Boolean, default: false},
  text: {type: String, required: true},
  position: {type: String, default: 'bottom'},
})

const isTextVisible = ref(false);


watch(() => props.isShown, (newVal) => {
  if (newVal) {
    setTimeout(() => {
      isTextVisible.value = true;
    }, 100)
  } else {
    isTextVisible.value = false;
  }
}, {immediate: true})
</script>
<style scoped lang="scss">
.ui-tooltip {
  background: #000;
  border: 1px solid #EBEBEB;
  border-radius: 8px;
  padding: 6px 8px;
  color: #FFF;
  overflow: hidden;
  text-align: left;
  display: flex;
  align-items: center;

  font-size: inherit;
  font-style: normal;
  font-weight: inherit;
  line-height: inherit;

  &::after {
    content: "";
    position: absolute;
    left: 50%;
    margin-left: -7px;
    border-width: 7px;
    border-style: solid;

  }

  &.top {
    &::after {
      bottom: calc(100% - 1px);
      border-color: transparent transparent black transparent;
    }
  }

  &.bottom {
    &::after {
      top: calc(100% - 1px);
      border-color: black transparent transparent transparent;
    }
  }

  &.hidden {
    &::after {
      display: none;
    }
  }
}
.show-up-enter-active,
.show-up-leave-active {
  transition: opacity 0.3s ease;
}

.show-up-enter-from, .show-up-leave-to {
  opacity: 0;
}

.show-up-enter-to, .show-up-leave-from {
  opacity: 1;
}
</style>
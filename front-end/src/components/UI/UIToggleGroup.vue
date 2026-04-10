<template>
  <div :class="['ui-toggle-group', size, skin, roundness]">
    <button
      v-for="elem in options"
      :key="elem.val"
      :class="['paragraph-small medium', {active: selectedOption === String(elem.val)}]"
      @click="emits('handleOptionSelect', elem)"
    >
      {{elem[label] || ''}}
    </button>
  </div>
</template>
<script setup>
defineProps({
  options: {type: Array, default: []},
  label: {type: String, default: ''},
  selectedOption: {type: String, default: ''},
  size: {type: String, default: 'regular'}, //regular | large | small | mini
  skin: {type: String, default: 'outlined'}, //outlined | ghost
  roundness: {type: String, default: 'default'}, //default | round
})
const emits = defineEmits(['handleOptionSelect']);
</script>
<style scoped lang="scss">
.ui-toggle-group {
  display: flex;
  align-items: center;
  overflow: hidden;

  & button {
    height: 100%;
    width: 100%;
    background: transparent;
    display: flex;
    align-items: center;
    justify-content: center;
    border-right: 1px solid #E5E7EB;

    &.active {
      //background: rgba(0, 0, 0, 0.05);
      background: #374151;
      color: #FFF;
    }

    &:last-child {
      border-right: none;
    }

    &:disabled {
      opacity: .5;
    }
  }

  &.regular {
    height: 36px;
  }

  &.large {
    height: 40px;
  }

  &.small {
    height: 32px;
  }

  &.mini {
    height: 24px;
  }

  &.default {
    border-radius: 8px;
    border: 1px solid #E5E7EB;
  }

  &.round {
    border-radius: 9999px;
    border: 1px solid #E5E7EB;
  }

  &.outlined {
    border: 1px solid #E5E7EB;
  }

  &.ghost {
    border: 1px solid transparent;
  }

  &.default {
    border-radius: 8px;
  }
}
</style>
<template>
  <div class="ui-ghost-btns-group">
    <button
      :class="['paragraph-mini medium', {active: String(elem.val) === selectedOption}]"
      v-for="elem in options"
      :key="elem.val"
      @click="emits('handleOptionSelect', elem.val)"
    >
      {{elem.label}}
    </button>
  </div>
</template>
<script setup>
const props = defineProps({
  options: {type: Array, default:[]},
  selectedOption: {type: String, default:''},
})

const emits = defineEmits(['handleOptionSelect']);
</script>
<style scoped lang="scss">
.ui-ghost-btns-group {
  display: flex;
  align-items: center;

  & button {
    height: 24px;
    padding: 0 8px;
    background: transparent;
    color: #4B5563;

    &:hover, &.active {
      color: #030712;
    }
  }
}
</style>
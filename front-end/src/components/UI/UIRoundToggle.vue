<template>
  <div @click="handleChange" :class="['ui-round-toggle', {reverse: isReverse}]">
    <span v-if="label" class="paragraph-small regular" >{{label}}</span>
    <button :class="{active: isActive}"></button>
  </div>
</template>
<script setup>
const props = defineProps({
  label: {type: String, default: ''},
  isReverse: {type: Boolean, default: false},
})

const isActive = defineModel('isActive', { type: Boolean, default: false });

const handleChange = () => {
  isActive.value = !isActive.value;
}
</script>
<style scoped lang="scss">
.ui-round-toggle {
  display: flex;
  align-items: center;
  gap: 8px;

  &.reverse {
    & span {
      order: 2;
    }
  }

  & button {
    transition: .3s ease;
    position: relative;
    border-radius: 12px;
    height: 18px;
    width: 33px;
    background: #E5E7EB;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

    &::before {
      content: "";
      position: absolute;
      width: 16px;
      height: 16px;
      background: #FFF;
      z-index: 1;
      border-radius: 50%;
      top: 50%;
      left: 1px;
      transform: translateY(-50%);
      transition: .3s ease;
    }

    &.active {
      background: #000;
      &::before {
        left: calc(50% - 1px);
      }
    }
  }
}
</style>
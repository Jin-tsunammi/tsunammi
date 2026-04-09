<template>
  <div class="ui-increase-decrease">
    <button class="increase" @click="emits('handleIncrease')">
      <SVGSmallArrowDown color="#030712"/>
    </button>
    <button class="decrease" @click="emits('handleDecrease')">
      <SVGSmallArrowDown color="#030712"/>
    </button>
  </div>
</template>
<script setup>
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";

const emits = defineEmits(['handleIncrease', 'handleDecrease']);
</script>
<style scoped lang="scss">
.ui-increase-decrease {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  & button {
    background: transparent;
    display: flex;
    align-items: center;
    justify-content: center;
    aspect-ratio: 1/1;
    min-width: 10px;
    max-width: 10px;

    &.increase {
      & svg {
        transform: rotate(180deg);
      }
    }

    &.decrease {
      & svg {
        position: relative;
        transform: translateX(0.5px);
      }
    }
  }
}
</style>
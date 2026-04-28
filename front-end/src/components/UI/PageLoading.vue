<template>
  <div class="loader-wrapper heading-5 grey">
    <SVGLogoTsunammi />
    <span>Page is loading...</span>
  </div>
</template>
<script setup>
import SVGLogoTsunammi from "../SVG/SVGLogoTsunammi.vue";
</script>
<style scoped lang="scss">
.loader-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  position: absolute;
  top: calc(50% - 63px);
  left: 50%;
  transform: translate(-50%, -50%);

  & svg {
    width: 36px;
    height: 36px;
    animation: rotate-center 4s linear infinite both;
  }
}

@keyframes rotate-center {
  0% {
    transform: rotate(0);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
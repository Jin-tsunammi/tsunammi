<template>
  <div class="default-avatar">
    <SVGLogoTsunammi color="#FFF" />
  </div>
</template>
<script setup>
import SVGLogoTsunammi from "../SVG/SVGLogoTsunammi.vue";
</script>
<style scoped lang="scss">
.default-avatar {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #000;

  & svg {
    width: 50%;
    height: 50%;
  }
}
</style>
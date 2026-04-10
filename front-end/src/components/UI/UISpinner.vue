<template>
  <span class="loader"></span>
</template>
<script setup>
import {computed} from "vue";

const props = defineProps({
  color: {type: String, default: "#FFF"},
})

const backgroundColor = computed(() => {
  return props.color;
})
</script>
<style scoped lang="scss">
.loader {
  width: 13px;
  height: 13px;
  border: 1.5px solid v-bind(backgroundColor);
  border-bottom-color: transparent;
  border-radius: 50%;
  display: inline-block;
  box-sizing: border-box;
  animation: rotation 1s linear infinite;
}

@keyframes rotation {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
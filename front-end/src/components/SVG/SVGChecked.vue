<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="11" height="8" viewBox="0 0 11 8" fill="none">
    <path d="M9.83333 0.5L3.41667 6.91667L0.5 4" :stroke="color" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
</template>
<script setup>
defineProps({
  color: {type: String, default: "white"},
})
</script>
<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" fill="none">
    <path d="M5.83301 14.1667L14.1663 5.83337M14.1663 5.83337H6.66634M14.1663 5.83337V13.3334" :stroke="color" stroke-width="1.66667" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
</template>
<script setup>
defineProps({
  color: {type: String, default: '#09090B'}
})
</script>

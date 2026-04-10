<template>
  <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 28 28" fill="none">
    <path d="M9.33301 16.3334C13.199 16.3334 16.333 13.1994 16.333 9.33337C16.333 5.46738 13.199 2.33337 9.33301 2.33337C5.46701 2.33337 2.33301 5.46738 2.33301 9.33337C2.33301 13.1994 5.46701 16.3334 9.33301 16.3334Z" :stroke="color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M21.1051 12.0984C22.208 12.5096 23.1894 13.1922 23.9585 14.0831C24.7277 14.9741 25.2597 16.0446 25.5055 17.1956C25.7513 18.3467 25.7029 19.5412 25.3647 20.6685C25.0265 21.7959 24.4096 22.8198 23.5708 23.6456C22.7321 24.4714 21.6987 25.0723 20.5661 25.3928C19.4336 25.7134 18.2385 25.7431 17.0914 25.4794C15.9444 25.2157 14.8823 24.667 14.0034 23.8841C13.1246 23.1011 12.4574 22.1092 12.0635 21.0001" :stroke="color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M8.16699 7H9.33366V11.6667" :stroke="color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M19.4948 16.1934L20.3115 17.0217L17.0215 20.3117" :stroke="color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
</template>
<script setup>
defineProps({
  color: {type: String, default: "#EA580C"},
})
</script>

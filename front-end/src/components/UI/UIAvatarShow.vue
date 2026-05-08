<template>
  <div class="ui-avatar-show">
    <img
      v-if="isPictureValid"
      v-show="imageLoaded && !imageError"
      :src="src"
      :alt="alt"
      @load="imageLoaded = true"
      @error="imageError = true"
    >
    <DefaultAvatar v-if="!isPictureValid || !imageLoaded || imageError" />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import DefaultAvatar from './DefaultAvatar.vue'
import { isTokenPictureValid } from '../../helpers/index.js'

const props = defineProps({
  mint: { type: String, default: '' },
  src: { type: String, default: '' },
  alt: { type: String, default: 'image' },
  isToken: { type: Boolean, default: true },
})

const imageLoaded = ref(false)
const imageError = ref(false)

const isPictureValid = computed(() => {
  if (!props.mint) return false
  if (!props.src) return false
  return isTokenPictureValid(props.mint, props.isToken)
})

watch(
  () => [props.mint, props.src],
  () => {
    imageLoaded.value = false
    imageError.value = false
  }
)
</script>

<style scoped lang="scss">
.ui-avatar-show {
  width: 100%;
  height: 100%;
  display: flex;
  border-radius: 50%;
  overflow: hidden;
}

.ui-avatar-show > img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>

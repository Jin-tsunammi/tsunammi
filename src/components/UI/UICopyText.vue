<template>
  <button class="ui-copy-text" @click="handleTextCopy">
    <SVGCopy v-if="!copied"/>
    <SVGCopied v-else/>
  </button>
</template>
<script setup>
import {useClipboard} from '@vueuse/core'
import SVGCopy from "../SVG/SVGCopy.vue";
import SVGCopied from "../SVG/SVGCopied.vue";

const props = defineProps({
  copyText: {type: String, default: ''},
})

const {copy, copied, isSupported} = useClipboard(props.copyText);

const handleTextCopy = async () => {
  if (!isSupported) {
    alert('Copy does not supported');

    return;
  }

  await copy(props.copyText);
}
</script>
<style scoped lang="scss">
.ui-copy-text {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  margin-left: 10px;

  & svg {
    width: 100%;
    height: 100%;
  }
}
</style>
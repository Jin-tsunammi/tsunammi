<template>
  <button
    :disabled="is_disabled"
    :class="['ui-button', textSize, type, color_type, size]"
    @click.stop="emits('cta')"
  >
    <slot name="left-icon"/>
    <slot />
    <slot name="right-icon"/>
  </button>
</template>
<script setup>
import {computed} from "vue";

 const props = defineProps({
  type: {type: String, default: 'default'}, //default (rectangle) | round
  color_type: {type: String, default: 'accent'}, //accent | primary | secondary | outline | ghost | destructive | ghost-muted
  size: {type: String, default: 'regular'}, //regular | large | small | mini+
  is_disabled: {type: Boolean, default: false},
})

const emits = defineEmits(['cta']);

const textSize = computed(() => {
  if (props.size === 'mini') return 'paragraph-mini';
  else return 'paragraph-small';
})
</script>
<style scoped lang="scss">
.ui-button {
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  transition: .3s;
  border: 1px solid transparent;
  gap: 6px;
  min-width: fit-content;

  &:disabled {
    opacity: .3;
    cursor: not-allowed;
  }

  &.default {
    border-radius: 8px;
  }

  &.rounded {
    border-radius: 9999px;
  }

  &.regular {
    height: 36px;
    padding: 0 16px;
  }

  &.large {
    height: 40px;
    padding: 0 24px;
  }

  &.small {
    height: 32px;
    padding: 0 12px;
  }

  &.mini {
    height: 24px;
    padding: 0 8px;
    font-weight: 500;
  }

  &.accent {
    background: #EA580C;
    color: #F9FAFB;

    &:hover, &:active {
      background: #F97316;
    }

    &:focus {
      background: #EA580C;
      box-shadow: 0 0 0 3px #FED7AA;
    }
  }

  &.primary {
    background: #111827;
    color: #F9FAFB;

    &:hover, &:active {
      background: #374151;
    }

    &:focus {
      background: #111827;
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }

  &.secondary {
    background: #F3F4F6;
    color: #111827;

    &:hover, &:active {
      background: #F9FAFB;
    }

    &:focus {
      background: #F9FAFB;
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }

  &.outline {
    border-color: #D1D5DB;
    color: #111827;
    background: rgba(255, 255, 255, 0.10);
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);

    &:hover, &:active {
      border-color: #D1D5DB;
      background: rgba(0, 0, 0, 0.03);
      box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
    }

    &:focus {
      border-color: #9CA3AF;
      background: rgba(255, 255, 255, 0.10);
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }

  &.ghost {
    background: transparent;
    color: #374151;

    &:hover, &:active {
      background: rgba(0, 0, 0, 0.05);
    }

    &:focus {
      background: transparent;
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }

  &.destructive {
    background: #DC2626;
    color: #FFF;

    &:hover, &:active {
      background: #DC2626;
    }

    &:focus {
      background: #DC2626;
      box-shadow: 0 0 0 3px #FCA5A5;
    }
  }

  &.ghost-muted {
    background: transparent;
    color: rgba(55, 65, 81, .5);

    &:hover, &:active {
      background: rgba(0, 0, 0, 0.05);
    }

    &:focus {
      background: transparent;
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }
}
</style>
<template>
  <div
    :class="['ui-checkbox', type]"
    @click="toggle"
    role="checkbox"
    :aria-checked="props.modelValue"
    :tabindex="disabled ? -1 : 0"
    @keydown.space.prevent="toggle"
  >
    <div
      :class="['ui-checkbox__box', containerClasses]"
    >
      <SVGChecked v-if="modelValue && type === 'default'" />
      <div v-if="modelValue && type === 'round'" class="dot"></div>
    </div>

    <span v-if="label" class="ui-checkbox__label paragraph-small">
      {{ label }}
    </span>
  </div>
</template>
<script setup>
import { computed } from 'vue';
import SVGChecked from "../SVG/SVGChecked.vue";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  type: {
    type: String,
    default: 'default' // default | round
  },
});

const emit = defineEmits(['update:modelValue', 'change']);

const toggle = () => {
  if (props.disabled) return;

  const newValue = !props.modelValue;
  emit('update:modelValue', newValue);
  emit('change', newValue);
};

const containerClasses = computed(() => ({
  'is-checked': props.modelValue,
  'is-disabled': props.disabled
}));
</script>
<style scoped lang="scss">
.ui-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;

  &.round {
    & .ui-checkbox {
      &__box {
        border-radius: 50%;
        background: #FFF;
        border-color: #D1D5DB;

        &.is-checked {
          & .dot {
            display: flex;
          }
        }
      }
    }
  }

  &__box {
    display: flex;
    flex-direction: column;
    min-width: 16px;
    max-width: 16px;
    min-height: 16px;
    max-height: 16px;
    border-radius: 4px;
    cursor: pointer;
    border: 1px solid #D1D5DB;
    align-items: center;
    justify-content: center;

    &.is-checked {
      border-color: #000;
      background: #000;
    }

    & svg {
      width: 70%;
      height: 70%;
    }
  }
  &__label {
    color: #374151;
    font-weight: 400;
  }

  & .dot {
    display: none;
    border-radius: 50%;
    width: 8px;
    height: 8px;
    background: #030712;
  }
}
</style>
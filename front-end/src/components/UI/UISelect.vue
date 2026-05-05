<template>
  <div class="ui-select">
    <span v-if="label" class="ui-select__label paragraph-small">{{ label }}</span>
    <OnClickOutside
      @trigger="handleClose"
      :class="['ui-select__input_wrapper', {disabled: is_disabled}]"
    >
      <div
        :class="[
        'ui-select__input paragraph-small',
        size,
        {
          error: errorMessage,
          placeholder: !selected,
          open: modelValue
        }
      ]"
        @click.stop="toggle"
      >
        <slot name="left-icon"/>
        <span>
        {{ selected ? selected : placeholder }}
      </span>

        <div class="ui-select__input_right">
          <slot name="right-icon"/>
          <button>
            <SVGSmallArrowDown color="#6B7280"/>
          </button>
        </div>
      </div>

      <div class="ui-select__dropdown">
        <slot name="dropdown"/>
      </div>
    </OnClickOutside>
    <span v-if="errorMessage" class="ui-select__error paragraph-mini">{{ errorMessage }}</span>
  </div>
</template>
<script setup>
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";
import {OnClickOutside} from "@vueuse/components";

const props = defineProps({
  label: {type: String, default: ''},
  placeholder: {type: String, default: ''},
  errorMessage: {type: String, default: ''},
  selected: {type: String, default: '', required: true},
  modelValue: {type: Boolean, default: false, required: true},
  is_disabled: {type: Boolean, default: false},
  size: {type: String, default: 'regular'}, // regular | large | small | mini
})

const emit = defineEmits(['update:modelValue', 'change']);

const toggle = () => {
  const newValue = !props.modelValue;
  emit('update:modelValue', newValue);
  emit('change', newValue);
};

const handleClose = () => {
  if (!props.modelValue) return;

  emit('update:modelValue', false);
  emit('change', false);
}
</script>
<style scoped lang="scss">
.ui-select {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;

  &__label {
    color: #030712;
    font-weight: 500;
  }

  &__dropdown {
    position: absolute;
    z-index: 10;
    top: calc(100% + 8px);
    left: 0;
    width: 100%;
  }

  &__input {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 400;
    overflow: hidden;
    text-overflow: ellipsis;

    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    border-radius: 8px;
    color: #030712;

    &_right {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-left: auto;
    }

    &_wrapper {
      position: relative;

      &.disabled {
        opacity: 0.5;
        pointer-events: none;

        &:hover {
          cursor: not-allowed;
        }
      }
    }

    & button {
      background: transparent;
      width: 16px;
      height: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: .3s ease;
    }

    &.open {
      box-shadow: 0 0 0 3px #D1D5DB;

      & button {
        transform: rotate(180deg);
      }
    }

    &.placeholder {
      color: #6B7280;
    }

    &.error {
      border-color: #EF4444;

      &.open {
        box-shadow: 0 0 0 3px #FCA5A5;
      }
    }

    &:has(input:focus) {

    }

    &.regular {
      height: 36px;
      padding: 0 12px;
    }

    &.large {
      height: 40px;
      padding: 0 16px;
    }

    &.small {
      height: 32px;
      padding: 0 8px;
    }

    &.mini {
      height: 24px;
      padding: 0 6px;
    }
  }

  &__error {
    color: #DC2626;
    font-weight: 400;
  }
}
</style>
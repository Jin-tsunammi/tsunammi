<template>
  <div :class="['base-input']">
    <div v-if="hasTopContent" class="base-input__top">
      <label v-if="label" :for="label" class="base-input__label paragraph-small">{{label}}</label>
      <slot name="top-right" />
    </div>
    <span v-if="topAddText" class="base-input__top_text paragraph-small regular grey">{{topAddText}}</span>
    <div :class="['base-input__input', size, ui_type, {error: errorMessage, disabled: is_disabled}]">
      <slot name="icon-left" />

      <input
        v-model="inputVal"
        :id="label || ''"
        :step="step"
        :maxlength="maxLength"
        :placeholder=placeholder
        :autocomplete="'off'"
        :readonly="is_readonly"
        :inputmode="inputMods"
        @blur="emits('handleBlur', $event)"
        @input="handleInput"
        @keydown="handleKeydown"
        class="paragraph-small"
      >

      <slot name="icon-right" />
    </div>
    <div v-if="hasBottomContent" class="base-input__bottom">
      <div class="base-input__bottom_left">
        <span v-if="errorMessage" class="base-input__error paragraph-mini">{{errorMessage}}</span>
        <span v-else-if="!errorMessage && bottomTextLeft">{{bottomTextLeft}}</span>
        <slot v-else name="bottom-left" />
      </div>
      <div v-if="$slots['bottom-right']" class="base-input__bottom_right">
        <slot name="bottom-right" />
      </div>
    </div>
  </div>
</template>
<script setup>
import {computed, useSlots} from "vue";

const props = defineProps({
  placeholder: {type: String, default: ''},
  type: {type: String, default: 'text'},
  ui_type: {type: String, default: 'default'}, // default | round
  maxLength: {type: Number, default: 60000},
  errorMessage: {type: String, default: ''},
  label: {type: String, default: ''},
  inputMod: {type: String, default: ''},
  topAddText: {type: String, default: ''},
  bottomTextLeft: {type: String, default: ''},
  is_readonly: {type: Boolean, default: false},
  is_disabled: {type: Boolean, default: false},
  is_dot_allowed: {type: Boolean, default: true},
  size: {type: String, default: 'regular'}, // regular | large | small | mini
  step: {type: String, default: 'any'},
})
const inputVal = defineModel({type: [String, Number], default: null});
const emits = defineEmits(['handleInput', 'handleBlur']);
const slots = useSlots()

const hasTopContent = computed(() => {
  return Boolean(
    props.label ||
    slots['top-right']
  )
})

const inputMods = computed(() => {
  if (props.type === 'number') {
    return "decimal"
  } else {
    return props.inputMod;
  }
})
const hasBottomContent = computed(() => {
  return Boolean(
    props.errorMessage ||
    props.bottomTextLeft ||
    slots['bottom-left'] ||
    slots['bottom-right']
  )
})
const topPadding = computed(() => {
  if (props.label) {
    return '0';
  } else if (props.topAddText) {
    return '26px';
  } else return '0';
})

function handleKeydown(event) {
  if (props.type !== 'number') return;
  const target = event?.target;
  const current = String(target?.value ?? '');

  if (event.key === '-') {
    event.preventDefault();
  }

  if ((event.key === ',' || event.key === '.') && target) {
    if (current === '' || current === '-' || !props.is_dot_allowed) {
      event.preventDefault();
      return;
    }

    if (current.includes('.') || current.includes(',')) {
      event.preventDefault();
    }
  }
}

function handleInput(event) {
  if (props.type === 'number') {
    let value = event.target.value;

    value = value.replace(/[^0-9.,]/g, '');

    if (!props.is_dot_allowed) {
      value = value.replace(/[.,]/g, '');
    }

    const parts = value.split(/[.,]/);
    if (parts.length > 2) {
      value = parts[0] + '.' + parts.slice(1).join('');
    }

    event.target.value = value;
    inputVal.value = value;
  }

  emits('handleInput', event);
}
</script>
<style scoped lang="scss">
@import "../../assets/styles/main.scss";

.base-input {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: v-bind(topPadding);

  &__label {
    color: #030712;
    font-weight: 500;
  }

  &__top {
    display: flex;
    align-items: center;
  }

  &__input {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;

    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

    &.error {
      border-color: #EF4444;

      &:has(input:focus) {
        box-shadow: 0 0 0 3px #FCA5A5;
      }
    }

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }

    &:has(input:focus) {
      box-shadow: 0 0 0 3px #D1D5DB;
    }

    &.default {
      border-radius: 8px;
    }

    &.round {
      border-radius: 9999px;
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

    & input {
      width: 100%;
      height: 100%;
      font-weight: 400;

      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  &__bottom {
    display: flex;
    align-items: center;
    gap: 10px;

    &_right {
      margin-left: auto;
    }
  }

  &__error {
    color: #DC2626;
    font-weight: 400;
  }
}
</style>
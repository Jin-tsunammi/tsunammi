<template>
  <div :class="['base-textarea']">
    <label v-if="label" :for="label" class="base-textarea__label paragraph-small">{{label}}</label>
    <div :class="['base-textarea__input', ui_type, {error: errorMessage}]">
      <textarea
        :id="label || ''"
        :maxlength="maxLength"
        v-model="inputVal"
        :placeholder=placeholder
        @blur="emits('handleBlur', $event)"
        @input="handleInput"
        :autocomplete="'off'"
        :readonly="is_readonly"
        class="paragraph-small"
      />

    </div>
    <span v-if="errorMessage" class="base-textarea__error paragraph-mini">{{errorMessage}}</span>
    <span v-if="is_max_length_visible" class="base-textarea__info grey paragraph-mini">{{`Max number of words: ${inputVal.length}/${maxLength}`}}</span>
  </div>
</template>
<script setup>
defineProps({
  placeholder: {type: String, default: ''},
  ui_type: {type: String, default: 'default'}, // default | round
  maxLength: {type: Number, default: 600000},
  errorMessage: {type: String, default: ''},
  label: {type: String, default: ''},
  is_readonly: {type: Boolean, default: false},
  is_max_length_visible: {type: Boolean, default: false},
})
const inputVal = defineModel({type: String, default: ''});
const emits = defineEmits(['handleInput', 'handleBlur']);

function handleInput(event) {
  inputVal.value = event.target.value;
  emits('handleInput', event);
}
</script>
<style scoped lang="scss">
@import "../../assets/styles/main.scss";

.base-textarea {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;

  &__label {
    color: #030712;
    font-weight: 500;
  }

  &__input {
    display: flex;
    align-items: center;
    justify-content: center;

    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    padding: 16px;

    &.error {
      border-color: #EF4444;

      &:has(textarea:focus) {
        box-shadow: 0 0 0 3px #FCA5A5;
      }
    }

    &:has(textarea:focus) {
      box-shadow: 0 0 0 3px #D1D5DB;
    }

    &.default {
      border-radius: 8px;
    }

    & textarea {
      width: 100%;
      height: 100%;
      font-weight: 400;
      border: none;
      outline: none;
      resize: none;

      overflow: auto;
    }
  }

  &__error {
    color: #DC2626;
    font-weight: 400;
  }
}
</style>
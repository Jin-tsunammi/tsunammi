<template>
  <Transition name="dropdown-fade">
    <div v-show="isOpen" ref="uiDropdownRef" class="ui-dropdown">
      <div v-if="isSearch" class="ui-dropdown__search">
        <UIBaseInput v-model="vModelSearch"/>
        <div class="ui-dropdown__divider"></div>
      </div>
      <ul v-show="options.length" class="ui-dropdown__list">
        <li
          class="ui-dropdown__item paragraph-small"
          :class="{active: item[label] === selectedOption, background: isSelectedBackground}"
          v-for="(item, i) in options"
          :key="item.id || i"
          @click.stop="emits('handleOptionSelect', item)"
        >
          <div v-if="item.image || item.svg" class="icon">
            <img v-if="item.image" :src="item.image" alt="icon">
            <component v-if="item.svg" :is="item.svg" />
          </div>
          <slot v-if="$slots['custom-option']" name="custom-option" :item="item"/>
          <span v-else>{{ item[label] || 'Option' }}</span>
        </li>
      </ul>

      <slot v-if="!options.length && $slots['custom-dropdown']" name="custom-dropdown"/>

      <div v-if="!options.length && !$slots['custom-dropdown']" class="ui-dropdown__empty paragraph-small">
        No options available
      </div>
    </div>
  </Transition>
</template>
<script setup>
import UIBaseInput from "./UIBaseInput.vue";
import {ref, watch} from "vue";

const props = defineProps({
  options: {type: Array, default: []},
  selectedOption: {type: String, default: ''},
  label: {type: String, default: ''},
  isOpen: {type: Boolean, default: false},
  isSearch: {type: Boolean, default: false},
  isSelectedBackground: {type: Boolean, default: true},
})
const emits = defineEmits(['handleOptionSelect']);
const vModelSearch = defineModel('search');
const uiDropdownRef = ref(null);

watch( () => props.isOpen, (val) => {
    if (!val && props.isSearch) {
      vModelSearch.value = "";
    }

    if (!val && uiDropdownRef.value) {
      uiDropdownRef.value.scrollTop = 0;
    }
  }
);
</script>
<style scoped lang="scss">
.ui-dropdown {
  border-radius: 8px;
  border: 1px solid #E5E7EB;
  background: #FFF;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.10), 0 2px 4px -2px rgba(0, 0, 0, 0.10);
  max-height: calc(36px * 6 + 6px);
  overflow-y: scroll;

  &__empty {
    height: 36px;
    padding: 0 12px;
    border-radius: 6px;
    font-weight: 400;
    transition: .3s ease;
    cursor: pointer;
    display: flex;
    align-items: center;
  }

  &__search {
    margin-top: 8px;
    padding: 0 12px;
  }

  &__divider {
    height: 1px;
    width: 100%;
    background: #D1D5DB;
    margin-top: 8px;
  }

  &__list {
    display: flex;
    flex-direction: column;
    padding: 2px;
    height: fit-content;
  }

  &__item {
    min-height: 36px;
    padding: 0 12px;
    border-radius: 6px;
    font-weight: 400;
    transition: .3s ease;
    cursor: pointer;
    display: flex;
    align-items: center;

    & .icon {
      min-width: 20px;
      min-height: 20px;
      max-height: 20px;
      max-width: 20px;
      margin-right: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #000;
      border-radius: 50%;

      & img {
        width: 70%;
        height: 70%;
      }
    }

    &:hover {
      background: #F3F4F6;
    }

    &.active.background {
      background: #E5E7EB;
    }
  }
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: opacity 0.2s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
}

.dropdown-fade-enter-to,
.dropdown-fade-leave-from {
  opacity: 1;
}
</style>
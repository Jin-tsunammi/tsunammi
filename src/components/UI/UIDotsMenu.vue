<template>
  <OnClickOutside @trigger="isMenuOpen = false" class="ui-dots-menu" @click.stop>
    <button
      class="ui-dots-menu__dots"
      @click.stop="toggleMenu"
    >
      <SVGDots/>
    </button>

    <Transition name="dropdown-fade">
      <div v-if="isMenuOpen" class="ui-dots-menu__dropdown">
        <div v-if="!$slots['custom-menu']" class="ui-dots-menu__actions">
          <div
            v-for="(group, index) in menu"
            :key="index"
            class="ui-dots-menu__group"
          >
            <div
              v-for="(action) in group"
              :key="action?.label"
              class="ui-dots-menu__action paragraph-small regular"
              :class="action.action"
              @click.stop="handleOptionSelect(action.action)"
            >
              <div class="ui-dots-menu__action_icon">
                <component v-if="action.icon" :is="action.icon" />
              </div>

              <span>{{action.label}}</span>
            </div>

            <div v-if="index !== menu.length - 1" class="ui-dots-menu__group_divider"></div>
          </div>
        </div>

        <slot v-else name="custom-menu" />
      </div>
    </Transition>
  </OnClickOutside>
</template>
<script setup>
import SVGDots from "../SVG/SVGDots.vue";
import {OnClickOutside} from "@vueuse/components";
import {ref} from "vue";

defineProps({
  menu: {type: Array, default: []},
})
const emits = defineEmits(["handleOptionSelect"]);

const isMenuOpen = ref(false);

const handleOptionSelect = (action) => {
  isMenuOpen.value = false;
  emits('handleOptionSelect', action);
}
const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value;
}
</script>
<style scoped lang="scss">
.ui-dots-menu {
  position: relative;
  display: flex;
  flex-direction: column;

  &__dots {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    transition: .3s ease;
    border-radius: 8px;
    min-width: 32px;
    min-height: 32px;

    &:hover{
      background: rgba(0, 0, 0, 0.05);
    }
  }

  &__dropdown {
    position: absolute;
    z-index: 10;
    top: calc(100% + 8px);
    right: 0;

    display: flex;
    min-width: 154px;
    padding: 8px 2px;
    flex-direction: column;
    align-items: center;

    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.10), 0 2px 4px -2px rgba(0, 0, 0, 0.10);
  }

  &__actions {
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  &__group {
    width: 100%;
    display: flex;
    flex-direction: column;

    &_divider {
      height: 1px;
      margin: 8px;
      width: 100%;
      background: #E5E7EB;
    }
  }

  &__action {
    width: 100%;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    background: transparent;
    transition: .3s ease;
    padding: 0 8px;
    border-radius: 6px;
    cursor: pointer;

    &.delete {
      color: #DC2626;

      ::v-deep(path) {
        fill: #EF4444;
      }
    }

    &:hover {
      background: #E5E7EB;
    }

    &_icon {
      display: flex;
      align-items: center;
      justify-content: center;
      aspect-ratio: 1/1;
      min-width: 20px;
      max-width: 20px;

      & svg {
        width: 80%;
        height: 80%;
      }
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
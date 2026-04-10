<template>
  <div
    :class="['ui-alert', status]"
  >
    <div :class="['ui-alert__icon_wrapper', {top: second_text}]">
      <component v-if="icon" class="ui-alert__icon" :is="icon"/>
    </div>

    <div class="ui-alert__info">
      <span class="ui-alert__text paragraph-small" v-html="text"></span>
      <span v-if="second_text" class="ui-alert__second-text paragraph-small" v-html="second_text"></span>
    </div>
  </div>
</template>
<script setup>
defineProps({
  status: {type: String, required: true, default: 'info'},
  text: {type: String, required: true, default: ''},
  second_text: {type: String, required: false, default: ''},
  icon: {type: Object, required:false, default: null,},
})
</script>
<style scoped lang="scss">
.ui-alert {
  border-radius: 8px;
  background: #FFF;
  //box-shadow: 1px 1px 24px 0 rgba(9, 10, 11, 0.50);
  position: relative;
  overflow: hidden;
  width: 100%;

  border: 1px solid #E5E7EB;

  color: #FFF;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;

  &__icon {
    min-width: 16px;
    max-width: 16px;
    aspect-ratio: 1/1;

    &_wrapper {
      min-width: 16px;
      min-height: 16px;
      max-width: 16px;
      max-height: 16px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;

      &.top {
        margin-top: 4px;
        align-self: flex-start;
      }
    }
  }

  &.success {
    & .ui-alert {
      &__icon {
        ::v-deep(path) {
          stroke: #4B5563;
        }
      }
    }
  }

  &.error {
    .ui-alert {
      &__text, &__second-text {
        color: #DC2626;
      }
    }
  }

  &.blue {
    .ui-alert {
      &__text {
        color: #3B82F6;
      }
    }
  }

  &__text {
    color: #030712;
    font-weight: 500;
  }

  &__info {
    display: flex;
    flex-direction: column;
  }

  &__second-text {
    color: #6B7280;
    font-weight: 400;
  }
}
</style>
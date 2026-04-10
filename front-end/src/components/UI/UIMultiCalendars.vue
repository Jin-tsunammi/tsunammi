<template>
  <div class="ui-calendar-range">
    <VueDatePicker
      ref="datepicker"
      v-model="dates"
      range
      multi-calendars
      :preview="false"
      :enable-time-picker="false"
      :formats="{ input: 'MMM d, yyyy', month: 'LLLL' }"
      placeholder="Pick a date"
      @cleared="emits('handleDatesChange', dates)"
    >
      <template #arrow-left>
        <SVGSmallArrowDown class="arrow arrow-left" color="#030712" />
      </template>
      <template #arrow-right>
        <SVGSmallArrowDown class="arrow arrow-right" color="#030712" />
      </template>
      <template #action-buttons>
        <div class="ui-calendar-range__actions">
          <UIButton
            @cta="closeDPMenu"
            color_type="ghost"
            size="large"
          >
            Cancel
          </UIButton>
          <UIButton
            color_type="primary"
            size="large"
            @cta="selectDate"
          >
            Apply
          </UIButton>
        </div>
      </template>
    </VueDatePicker>
  </div>
</template>
<script setup>
import { VueDatePicker } from "@vuepic/vue-datepicker"
import {ref, useTemplateRef} from 'vue';
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";
import UIButton from "./UIButton.vue";

const emits = defineEmits(["handleDatesChange"]);
const dates = ref();

const dp = useTemplateRef('datepicker')

const selectDate = () => {
  dp.value?.selectDate();

  emits("handleDatesChange", dates.value);
}

const closeDPMenu = () => {
  dp.value?.closeMenu();
}
</script>
<style scoped lang="scss">
.ui-calendar-range {
  & .arrow {
    display: flex;
    height: 50%;
    width: 50%;

    &-left {
      transform: rotate(90deg);
    }
    &-right {
      transform: rotate(-90deg);
    }
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  ::v-deep(.dp__main) {
    font-family: Geist, "sans-serif";
    width: min-content;

    & .dp__input {
      height: 40px;
      width: min-content;
      border-radius: 8px;
      border: 1px solid #E5E7EB;
      background: #FFF;
      box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

      color: #030712;
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: 150%;
      letter-spacing: 0.07px;

      &::placeholder {
        color: #6B7280;
        font-family: Geist, "sans-serif";
        font-size: 14px;
        font-style: normal;
        font-weight: 400;
        line-height: 150%;
        letter-spacing: 0.07px;
      }
    }
    & .dp__input_focus {
      border: 1px solid #9CA3AF;
      background: #F3F4F6;
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }
  ::v-deep(.dp__menu) {
    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.10), 0 2px 4px -2px rgba(0, 0, 0, 0.10);
    font-family: Geist, "sans-serif";

    & .dp__calendar_header {
      color: #6B7280;
      font-family: Geist, "sans-serif";
      font-size: 12px;
      font-style: normal;
      font-weight: 400;
      line-height: 150%;
      letter-spacing: 0.18px;
    }

    & .dp__overlay_cell {

    }

    & .dp__inner_nav {
      display: flex;
      align-items: center;
      justify-content: center;
    }

    & .dp__inner_nav:hover {
      background: rgba(255, 255, 255, 0.10);
    }

    & .dp__arrow_top {
      display: none;
    }

    & .dp--header-wrap {
      margin-bottom: 16px;
    }

    & .dp--arrow-btn-nav {
      min-height: 32px;
      min-width: 32px;
      border-radius: 8px;
      border: 1px solid #E5E7EB;
      background: rgba(255, 255, 255, 0.10);
      box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);

      display: flex;
      align-items: center;
      justify-content: center;
    }

    & .dp--year-select, & .dp__month_year_select, & .dp__overlay_cell, & .dp__overlay_cell_active {
      color: #030712;
      font-family: Geist, "sans-serif";
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: 150%;
      letter-spacing: 0.07px;
    }

    & .dp__overlay_cell_active {
      color: #FFF;
    }

    & .dp__selection_preview {
      display: none;
    }
    & .dp--tp-wrap {
      display: none;
    }

    & .dp__calendar_row {
      gap: 1px;
    }

    & .dp__range_start, & .dp__range_end {
      background: #EA580C;
    }

    & .dp__cell_inner {
      height: 48px;
      width: 48px;
      border-radius: 0;

      font-family: Geist, "sans-serif";
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: 150%;
      letter-spacing: 0.07px;
    }
  }
}

</style>
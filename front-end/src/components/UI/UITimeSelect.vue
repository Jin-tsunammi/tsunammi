<template>
  <OnClickOutside :class="['ui-time-picker', {error: errorMessage}]" @trigger="closeDropdown">
    <div
      class="ui-time-picker__selected"
      :class="{active: isDropDown}"
      @click="toggleDropdown"
    >
      <SVGClock color="#6B7280"/>
      <span :class="['paragraph-small regular', {grey: !time}]">{{time ? time : 'Time'}}</span>
      <SVGClose v-if="time" @click.stop="clearTime"/>
    </div>

    <Transition name="dropdown-fade">
      <div v-if="isDropDown" class="ui-time-picker__dropdown">

        <div class="ui-time-picker__dropdown_inputs">
          <div class="input">
            <input
              class="paragraph-small regular"
              type="text"
              v-model="newTime.hh"
              inputmode="numeric"
              @input="validateTime('hh')"
            >
            <UINumberIncreaseDecrease
              @handle-increase="handleTimeIncrease('hh')"
              @handle-decrease="handleTimeDecrease('hh')"
            />
          </div>
          <span class="paragraph-small regular">:</span>
          <div class="input">
            <input
              class="paragraph-small regular"
              type="text"
              v-model="newTime.mm"
              inputmode="numeric"
              @input="validateTime('mm')"
            >
            <UINumberIncreaseDecrease
              @handle-increase="handleTimeIncrease('mm')"
              @handle-decrease="handleTimeDecrease('mm')"
            />
          </div>
        </div>

        <div class="divider"></div>

        <div class="ui-time-picker__dropdown_btns">
          <UIButton
            color_type="ghost"
            size="large"
            @cta="closeDropdown"
          >
            Cancel
          </UIButton>
          <UIButton
            color_type="primary"
            size="large"
            @cta="applyTimeChange"
          >
            Apply
          </UIButton>
        </div>
      </div>
    </Transition>
  </OnClickOutside>
</template>
<script setup>
import {ref, watch} from "vue";
import {OnClickOutside} from "@vueuse/components";
import SVGClock from "../SVG/SVGClock.vue";
import SVGClose from "../SVG/SVGClose.vue";
import UIButton from "./UIButton.vue";
import UINumberIncreaseDecrease from "./UINumberIncreaseDecrease.vue";

const props = defineProps({
  item: {type: Object, default: null},
  errorMessage: {type: String, default: ''},
})
const emits = defineEmits(["handleTimeApply"]);
const time = ref('');
const newTime = ref({
  hh: '00',
  mm: '00'
})
const MAX = {
  hh: 23,
  mm: 59,
}
const isDropDown = ref(false);

const closeDropdown = () => {
  isDropDown.value = false;

  if (!time.value) {
    clearTime()
    return;
  }

  const timeSplit = time.value.split(":");

  if (timeSplit[0] !== newTime.value.hh || timeSplit[1] !== newTime.value.mm) {
    if (!newTime.value.hh) newTime.value.hh = '00';
    if (!newTime.value.mm) newTime.value.mm = '00';

    newTime.value.hh = timeSplit[0];
    newTime.value.mm = timeSplit[1];
  }
}

const toggleDropdown = () => {
  if (isDropDown.value) {
    closeDropdown()
  } else {
    isDropDown.value = true;
  }
}

const clearTime = () => {
  time.value = '';
  newTime.value = {
    hh: '00',
    mm: '00'
  };
}

const handleTimeIncrease = (type) => {
  const max = MAX[type];
  const newVal = (Number(newTime.value[type]) || 0) + 1;

  if (newVal > max) newTime.value[type] = '00'
  else newTime.value[type] = String(newVal).padStart(2, '0');
}

const handleTimeDecrease = (type) => {
  const max = MAX[type];
  const newVal = (Number(newTime.value[type]) || 0) - 1;

  if (newVal < 0) newTime.value[type] = String(max);
  else newTime.value[type] = String(newVal).padStart(2, '0');
}

const applyTimeChange = () => {
  if (!newTime.value.hh) newTime.value.hh = '00';
  if (!newTime.value.mm) newTime.value.mm = '00';

  time.value = newTime.value.hh + ':' + newTime.value.mm;
  closeDropdown();
  emits('handleTimeApply', {time: time.value, item: props.item})
}

const validateTime = (type) => {
  let value = newTime.value[type]

  value = value.replace(/\D/g, '')

  if (value === '') {
    newTime.value[type] = ''
    return
  }

  let num = Number(value)

  if (num > MAX[type]) num = MAX[type]
  if (num < 0) num = 0

  newTime.value[type] = String(num).padStart(2, '0')
}

watch(() => props.item, (newValue) => {
  const nextStartAt = newValue?.start_at;

  if (!nextStartAt) {
    if (time.value) {
      clearTime();
    }
    return;
  }

  const nextDate = new Date(nextStartAt);
  if (Number.isNaN(nextDate.getTime())) return;

  const nextHh = String(nextDate.getHours()).padStart(2, '0');
  const nextMm = String(nextDate.getMinutes()).padStart(2, '0');
  const nextTime = `${nextHh}:${nextMm}`;

  if (time.value !== nextTime) {
    time.value = nextTime;
  }

  if (newTime.value.hh !== nextHh || newTime.value.mm !== nextMm) {
    newTime.value.hh = nextHh;
    newTime.value.mm = nextMm;
  }
}, {immediate: true, deep: true})
</script>
<style scoped lang="scss">
.ui-time-picker {
  height: 40px;
  width: 123px;
  position: relative;

  &.error {
    & .ui-time-picker__selected {
      border-color: #EF4444
    }
  }

  &__selected {
    height: 100%;
    width: 100%;
    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;

    &.active {
      box-shadow: 0 0 0 3px var(--focus-ring, #D1D5DB);
    }
  }

  &__dropdown {
    width: 230px;
    position: absolute;
    z-index: 10;
    top: calc(100% + 8px);
    right: 0;

    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.10), 0 2px 4px -2px rgba(0, 0, 0, 0.10);
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;

    &_inputs {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 48px;
    }

    & .divider {
      height: 1px;
      width: 100%;
      background: #E5E7EB;
    }

    &_btns {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }
}

.input {
  display: flex;
  align-items: center;
  border-radius: 8px;
  background: #F3F4F6;
  height: 100%;
  padding-right: 10px;

  & input {
    width: 48px;
    text-align: center;
    background: transparent;
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
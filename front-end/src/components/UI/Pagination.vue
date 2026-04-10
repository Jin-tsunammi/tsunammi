<template>
  <div class="pagination__wrapper">
    <button
      :disabled="isPrevBtnDisabled"
      @click="handleBtnClick('prev')"
      class="pagination__btn prev paragraph-small"
    >
      <SVGSmallArrowDown color="#030712"/>
      Previous
    </button>
    <ul class="pagination">
      <li
        :class="['pagination__item paragraph-small', {'active': item === currentPage}]"
        v-for="item in pages"
        :key="item"
        @click="handlePageChange(item)"
      >
        {{ item }}
      </li>
    </ul>
    <button
      :disabled="isNextBtnDisabled"
      @click="handleBtnClick('next')"
      class="pagination__btn next paragraph-small"
    >
      Next
      <SVGSmallArrowDown color="#030712"/>
    </button>
  </div>
</template>
<script setup>
import {computed} from "vue";
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";

const props = defineProps({
  total: {type: Number, default: 0},
  currentPage: {type: Number, default: 0},
})
const emits = defineEmits(["cta"]);

const startPage = computed(() => {
  if (props.currentPage === 1) {
    return 1;
  }

  return props.currentPage - 1;
})

const pages = computed(() => {
  let arr = [];
  const distinction = props.total - startPage.value;

  if (props.total <= 6) {
    for (let i = 1; i <= props.total; i++) {
      arr.push(i);
    }
  } else if (distinction < 6) {
    const start = props.total - 5;
    for (let i = start; i <= props.total; i++) {
      arr.push(i);
    }
  } else if (props.currentPage === 1) {
    arr = [
      props.currentPage,
      props.currentPage + 1,
      props.currentPage + 2,
      props.currentPage + 3,
      '...',
      props.total
    ]
  } else {
    arr = [
      startPage.value,
      props.currentPage,
      props.currentPage + 1,
      props.currentPage + 2,
      '...',
      props.total
    ]
  }

  return arr;
});

const handlePageChange = (item) => {
  if (item === '...') return;

  emits('cta', item);
}

const isPrevBtnDisabled = computed(() => {
  return props.currentPage === 1;
})

const isNextBtnDisabled = computed(() => {
  return props.currentPage === props.total;
})

const handleBtnClick = (direction) => {
  if (direction === 'next') {
    const nextPage = props.currentPage + 1;
    if (nextPage > props.total) return;

    emits('cta', props.currentPage + 1);
  } else if (direction === 'prev') {
    const prevPage = props.currentPage - 1;

    if (prevPage < 1) return;

    emits('cta', props.currentPage - 1);
  }
}
</script>
<style scoped lang="scss">

.pagination {
  height: 100%;
  display: flex;
  gap: 8px;

  &__wrapper {
    height: 36px;
    display: flex;
    gap: 8px;
    border-radius: 30px;
    padding: 0 16px;
    width: fit-content;
  }

  &__item {
    height: 100%;
    min-width: 34px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    user-select: none;
    border-radius: 8px;
    border: 1px solid transparent;

    color: #232323;
    font-weight: 500;

    &.active {
      color: #030712;
      border-color: #D1D5DB;
      background: rgba(255, 255, 255, 0.10);
      box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
    }
  }

  &__btn {
    display: flex;
    align-items: center;
    justify-content: center;
    aspect-ratio: 1/1;
    background: transparent;
    gap: 8px;
    padding-inline: 16px;
    font-weight: 500;

    &.next {
      & svg {
        transform: rotate(-90deg);
      }
    }

    &.prev {
      & svg {
        transform: rotate(90deg);
      }
    }

    &:disabled {
      opacity: 0.5;
    }
  }
}
</style>
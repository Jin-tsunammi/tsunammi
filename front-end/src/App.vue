<template>
  <main class="app">
    <Sidebar/>
    <div :class="['app__right', {fullWidth: isPageFullWidth}]">
      <div class="app__right_header">
        <DashboardHeader/>
      </div>
      <div :class="['app__content_wrapper']">
        <div :class="['app__content', {padding: isPadding}]">
          <RouterView/>
        </div>
      </div>
    </div>
    <Toast/>
    <CookiesNotification v-if="isCookiesVisible" :is-cookies="isCookiesVisible" @close-cookies="closeCookiesModal"/>
  </main>
</template>
<script setup>
import Toast from "./components/UI/Toast.vue";
import CookiesNotification from "./components/Login/CookiesNotification.vue";
import {computed, onMounted, ref} from "vue";
import {useRoute} from "vue-router";
import Sidebar from "./components/Base/Sidebar.vue";
import DashboardHeader from "./components/Base/DashboardHeader.vue";

const isCookiesVisible = ref(false);
const route = useRoute();
const isPadding = computed(() => {
  return route.name !== 'Home'
})
const isPageFullWidth = computed(() => {
  const pages = ['Home'];

  return pages.includes(route.name);
})

const closeCookiesModal = () => {
  isCookiesVisible.value = false;
}

onMounted(() => {
  const cookies = JSON.parse(localStorage.getItem('is_cookies'));

  if (!cookies) {
    isCookiesVisible.value = true;
  }
})
</script>
<style lang="scss">
@import './assets/styles/_main.scss';

.app {
  display: flex;
  height: 100dvh;
  width: 100dvw;
  position: relative;

  &__header-mobile {
    display: none;
  }

  &__right {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: #F3F4F6;

    & .app__content {
      max-width: 1211px;
      align-self: center;
    }

    &_header {
      display: flex;
    }

    &.fullWidth {
      & .dashboard-header__inner, .app__content {
        max-width: none;
      }
    }
  }

  &__content {
    flex: 1;
    width: 100%;
    position: relative;
    display: flex;
    flex-direction: column;


    &_wrapper {
      flex: 1;
      min-height: 0;
      width: 100%;
      display: flex;
      flex-direction: column;
      padding: 0 24px;
      overflow: auto;
    }

    &.padding {
      padding: 24px 0;
    }
  }
}

@media (max-width: 1200px) {
  .app {
    &__header-mobile {
      display: flex;
    }

    &__content {
      height: calc(100dvh - 56px);

      &.home {
        height: 100dvh;
      }
    }
  }
}
</style>

<template>
  <div class="header-mobile">
    <div class="header-mobile__inner">
      <div v-show="!sidebarStore.isMobileMenuOpen" class="header-mobile__logo-block">
        <SVGLogo color="white"/>
        <span>MMT</span>
      </div>
      <div class="header-mobile__right">
        <div v-if="route.name !== 'Login' && route.name !== 'SignUp' && !sidebarStore.isMobileMenuOpen" class="header-mobile__dashboard">
          <button>
            <SVGZap />
          </button>
          <router-link class="header-mobile__profile" :to="{name: 'DashboardProfile'}">
            <img src="../../../public/images/default-avatar.webp" alt="Icon">
          </router-link>
        </div>
        <HamburgerMenu />
      </div>
    </div>
  </div>
</template>
<script setup>
import HamburgerMenu from "../UI/HamburgerMenu.vue";
import SVGZap from "../SVG/SVGZap.vue";
import {useRoute} from "vue-router";
import {useSidebarStore} from "../../store/sidebarStore.js";
import SVGLogo from "../SVG/SVGLogo.vue";

const route = useRoute();
const sidebarStore = useSidebarStore()
</script>

<style scoped lang="scss">
.header-mobile {
  width: 100%;
  &__inner {
    height: 56px;
    width: 100%;
    background: #030712;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
  }

  &__right {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-left: auto;
  }

  &__dashboard {
    display: flex;
    align-items: center;
    gap: 12px;

    & button {
      background: transparent;
    }
  }

  &__profile {
    display: flex;
    border-radius: 50%;
    aspect-ratio: 1/1;
    min-width: 28px;
    max-width: 28px;
  }

  &__logo-block {
    display: flex;
    align-items: center;
    gap: 12px;
    color: #FFF;
    font-size: 19.327px;
    font-style: normal;
    font-weight: 700;
    line-height: normal;
    letter-spacing: -0.773px;

    & svg {
      width: 22px;
      height: 26px;
    }
  }
}
</style>
<template>
  <Transition name="mobile-slide">
    <div v-if="sidebarStore.isMobileMenuOpen" class="mobile-sidebar">
      <div class="mobile-sidebar__inner">
        <div v-if="route.name === 'Login' || route.name === 'SignUp'" class="mobile-sidebar__login-menu">
          <router-link
            v-for="link in loginNavLinks"
            :key="link.text"
            :to="link.link"
            class="mobile-sidebar__login-menu_link paragraph-medium"
          >{{link.label}}</router-link>
        </div>
        <div v-else class="mobile-sidebar__dashboard-menu">
          <SidebarMenu />
        </div>
      </div>
    </div>
  </Transition>
</template>
<script setup>
import {useSidebarStore} from "../../store/sidebarStore.js";
import {useRoute} from "vue-router";
import SidebarMenu from "./SidebarMenu.vue";

const route = useRoute()
const sidebarStore = useSidebarStore();

const loginNavLinks = [
  {
    label: "Home",
    link: ""
  },
  {
    label: "Target",
    link: ""
  },
  {
    label: "Smart Buyback",
    link: ""
  },
  {
    label: "Tokens",
    link: ""
  },
]

</script>
<style scoped lang="scss">
.mobile-sidebar {
  position: absolute;
  z-index: 20;
  bottom: 0;
  left: 0;
  height: calc(100dvh - 56px);
  width: 100%;

  &__inner {
    height: 100%;
    width: 100%;
    background: #030712;
    overflow: auto;
  }

  &__login-menu {
    padding-top: 78px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 44px;

    &_link {
      color: #FFF;
      font-weight: 400;
    }
  }

  &__dashboard-menu {
    padding-top: 18px;
    margin-bottom: 20px;
  }
}

.mobile-slide-enter-from,
.mobile-slide-leave-to {
  transform: translateY(100%);
}

.mobile-slide-enter-to,
.mobile-slide-leave-from {
  transform: translateY(0%);
}

.mobile-slide-enter-active,
.mobile-slide-leave-active {
  transition: transform 0.3s ease;
}
</style>
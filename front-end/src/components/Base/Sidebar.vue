<template>
  <div class="sidebar">
    <div class="sidebar__inner">
      <router-link :to="{name: 'Home'}" class="sidebar__logo-block">
        <SVGLogoTsunammi color="#EFF0F2" class="logo"/>
        <SVGLogoTsunammiFull class="name" />
      </router-link>

      <div class="sidebar__menu">
        <SidebarMenu />
      </div>
    </div>
  </div>
</template>
<script setup>
import SidebarMenu from "./SidebarMenu.vue";
import SVGLogoTsunammiFull from "../SVG/SVGLogoTsunammiFull.vue";
import SVGLogoTsunammi from "../SVG/SVGLogoTsunammi.vue";
</script>
<style scoped lang="scss">
.sidebar {
  display: flex;
  flex-direction: column;
  min-width: 293px;
  width: 293px;
  height: 100vh;
  overflow: auto;

  &__inner {
    padding: 24px;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  &__logo-block {
    display: flex;
    align-items: center;
    gap: 12px;
    color: #FFF;
    font-size: 25.253px;
    font-style: normal;
    font-weight: 700;
    line-height: normal;
    letter-spacing: -1.01px;

    & .logo {
      width: 32px;
      height: 32px;
    }
  }

  &__menu {
    margin-top: 62px;
    height: 100%;
  }
}

@media (max-width: 1200px) {
  .sidebar {
    display: none;
  }
}
</style>
<template>
  <button @click="toggleMobMenu" :class="['menu', {open: generalStore.isMobileMenuOpen}]">
    <span></span>
    <span></span>
    <span></span>
  </button>
</template>
<script setup>
import {useSidebarStore} from "../../store/sidebarStore.js";

const generalStore = useSidebarStore();

const toggleMobMenu = () => {
  generalStore.toggleMobileMenu();
}
</script>
<style scoped lang="scss">
.menu {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  position: relative;
  cursor: pointer;
  background: transparent;
  border: none; // Убираем дефолтные границы кнопки
  padding: 0;

  & span {
    display: block;
    position: absolute;
    height: 1.5px;
    width: 100%;
    background: #fff;
    border-radius: 9px;
    transition: .25s ease-in-out;
    left: 0;
    transform-origin: center;

    &:nth-child(1) {
      top: 4px;
    }

    &:nth-child(2) {
      top: 50%;
      transform: translateY(-50%);
    }

    &:nth-child(3) {
      bottom: 4px;
    }
  }

  &.open span {
    &:nth-child(1) {
      top: 50%;
      transform: translateY(-50%) rotate(45deg);
    }

    &:nth-child(2) {
      width: 0;
      opacity: 0;
      left: 50%; // Эффект «схлопывания» к центру
    }

    &:nth-child(3) {
      bottom: 50%;
      transform: translateY(50%) rotate(-45deg);
    }
  }
}
</style>
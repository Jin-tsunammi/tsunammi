<template>
  <div class="login__right">
    <div class="login__right_main">
      <button class="close" @click="modalsStore.closeModal()">
        <SVGClose/>
      </button>
      <img class="abstract main" src="../../../public/images/login-background.webp" alt="Image">
      <img class="abstract inline left" src="../../../public/images/login-left-abstract.webp" alt="Image">
      <img class="abstract inline right" src="../../../public/images/login-right-abstract.webp" alt="Image">
      <div class="login__right_content">
        <div class="login__right_logo">
          <SVGLogoTsunammi />
          <span>tsunammi</span>
        </div>
        <LoginForm />
      </div>
    </div>
  </div>
</template>
<script setup>
import LoginForm from "./LoginForm.vue";
import SVGClose from "../SVG/SVGClose.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import SVGLogoTsunammi from "../SVG/SVGLogoTsunammi.vue";

const modalsStore = useModalsStore();
</script>
<style scoped lang="scss">
.login {
  &__right {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 9px;

    &_main {
      width: 100%;
      height: 100%;
      background: #F3F4F6;
      border-radius: 12px;
      position: relative;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;

      & .close {
        position: absolute;
        z-index: 3;
        top: 24px;
        right: 24px;

        & svg {
          width: 20px;
          height: 20px;
        }
      }
    }

    & .login__bottom {
      display: none;
    }

    & .abstract {
      position: absolute;
      z-index: 2;

      &.inline {
        top: 50%;
        transform: translateY(-50%);
        height: 396px;
        width: 128px;
      }

      &.main {
        top: 0;
      }

      &.left {
        left: 0;
      }

      &.right {
        right: 0;
      }
    }

    &_ellipse {
      position: absolute;
      width: 74.3%;
      aspect-ratio: 511 / 214;
      left: 50%;
      bottom: -15%;
      z-index: 2;
      transform: translateX(-50%);
      border-radius: 511px;
      opacity: 0.5;
      background: #9CA3AF;
      filter: blur(50px);
    }

    &_content {
      position: relative;
      z-index: 3;
      width: 100%;
      max-width: 400px;
      display: flex;
      flex-direction: column;
      gap: 40px;
    }

    &_logo {
      align-self: center;
      display: flex;
      align-items: center;
      gap: 12px;

      & svg {
        width: 40px;
        height: 40px;
      }

      & span {
        color: #030712;
        font-family: Geist, sans-serif;
        font-size: 29.033px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }
    }
  }
}
</style>
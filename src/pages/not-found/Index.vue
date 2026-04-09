<template>
  <div class="error-page">
    <div class="error-page__inner">
      <div class="error-page__content">
        <svg xmlns="http://www.w3.org/2000/svg" width="211" height="90" viewBox="0 0 211 90" fill="none">
          <path d="M185.241 88.059V70.3501H145.456V57.7356L183.3 1.94043H200.16V57.0078H210.47V70.3501H200.16V88.059H185.241ZM159.769 57.0078H185.241V21.2262L159.769 57.0078Z" fill="#D1D5DB"/>
          <path d="M84.2509 74.5957L116.515 14.5553L126.219 15.4043L93.9545 75.4447L84.2509 74.5957ZM105.235 90C98.8466 90 93.348 88.1806 88.7388 84.5418C84.1297 80.903 80.5717 75.7278 78.065 69.0162C75.6391 62.3046 74.4261 54.3396 74.4261 45.1213C74.4261 35.8221 75.6391 27.8167 78.065 21.1051C80.5717 14.3935 84.1297 9.21833 88.7388 5.57952C93.348 1.85984 98.8466 0 105.235 0C111.704 0 117.202 1.85984 121.731 5.57952C126.34 9.21833 129.857 14.3935 132.283 21.1051C134.79 27.8167 136.043 35.8221 136.043 45.1213C136.043 54.3396 134.79 62.3046 132.283 69.0162C129.857 75.7278 126.34 80.903 121.731 84.5418C117.202 88.1806 111.704 90 105.235 90ZM105.235 76.4151C108.469 76.4151 111.219 75.2426 113.483 72.8976C115.747 70.4717 117.485 66.9137 118.698 62.2237C119.911 57.5337 120.518 51.8329 120.518 45.1213C120.518 38.248 119.911 32.5067 118.698 27.8976C117.485 23.2075 115.747 19.6496 113.483 17.2237C111.219 14.7978 108.469 13.5849 105.235 13.5849C102.081 13.5849 99.3318 14.7978 96.9868 17.2237C94.7227 19.6496 92.9841 23.2075 91.7712 27.8976C90.5582 32.5067 89.9518 38.248 89.9518 45.1213C89.9518 51.8329 90.5582 57.5337 91.7712 62.2237C92.9841 66.9137 94.7227 70.4717 96.9868 72.8976C99.3318 75.2426 102.081 76.4151 105.235 76.4151Z" fill="#D1D5DB"/>
          <path d="M39.7844 88.059V70.3501H0V57.7356L37.8437 1.94043H54.7035V57.0078H65.0135V70.3501H54.7035V88.059H39.7844ZM14.3127 57.0078H39.7844V21.2262L14.3127 57.0078Z" fill="#D1D5DB"/>
        </svg>
        <h1 class="error-page__title">Page not found</h1>
        <span class="error-page__text paragraph-medium">The page you’re looking for doesn’t exist or you don’t have access to it.</span>
        <div class="error-page__btns">
          <UIButton color_type="outline" size="large" @cta="goBack">
            Back
          </UIButton>
          <UIButton color_type="primary" size="large" @cta="goDashboard">
            Go to Home
          </UIButton>
        </div>
      </div>
    </div>
    <Modals />
  </div>
</template>
<script setup>
import DashboardHeader from "../../components/Base/DashboardHeader.vue";
import UIButton from "../../components/UI/UIButton.vue";
import {useRouter} from "vue-router";
import Modals from "../../components/UI/Modals.vue";

const router = useRouter();

const goBack = () => {
  if (window.history.length > 1) {
    router.back();
  } else {
    router.push({ name: "DashboardProfile" });
  }
};

const goDashboard = () => {
  router.push({ name: "Home" });
};
</script>
<style scoped lang="scss">
.error-page {
  height: 100%;

  &__inner {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  &__content {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding-top: 157px;
  }

  &__title {
    margin-top: 53px;
    color: #030712;

    font-size: 48px;
    font-style: normal;
    font-weight: 600;
    line-height: 100%;
    letter-spacing: -0.48px;
  }

  &__text {
    margin-top: 16px;
    display: block;
    font-weight: 400;
    color: #6B7280;
  }

  &__btns {
    margin-top: 40px;
    display: flex;
    align-items: center;
    gap: 12px;

    & .ui-button {
      min-width: 165px;
    }
  }
}

@media (max-width: 1200px) {
  .error-page {
    &__header {
      display: none;
    }

    &__text {
      text-align: center;
    }

    &__title {
      font-size: 30px;
      font-style: normal;
      font-weight: 600;
      line-height: 100%;
      letter-spacing: -0.3px;
    }

    &__content {
      padding: 0 16px;
      justify-content: center;
      position: relative;
      transform: translateY(-56px);

      & svg {
        width: 177px;
        height: 76px;
      }
    }

    &__btns {
      width: 100%;

      & .ui-button {
        width: 100%;
        min-width: auto;
      }
    }
  }
}
</style>
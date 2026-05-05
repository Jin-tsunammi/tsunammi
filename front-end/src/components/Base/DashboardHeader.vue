<template>
  <div class="dashboard-header">
    <div :class="['dashboard-header__inner']">
      <div v-if="isBackButtonVisible" class="dashboard-header__back paragraph-small medium" @click="router.back()">
        <SVGArrowPrevious/>
        Back
      </div>
      <div class="dashboard-header__right">
        <UIButton v-if="isRefreshVisible" color_type="ghost" size="small" @cta="handleHeaderRefresh">
          <template #left-icon>
            <SVGRefresh />
          </template>
          Refresh
        </UIButton>
        <div v-if="isRefreshVisible" class="divider"></div>
        <UISelect
          v-if="userStore.isUserAuth"
          :selected="loginData.email || ''"
          v-model="isDropDownOpen"
          size="large"
          class="dashboard-header__select"
        >
          <template #left-icon>
            <SVGProfileIcon />
          </template>
          <template #dropdown>
            <UIDropdown
              :is-open="isDropDownOpen"
              :options="options"
              label="name"
              @handle-option-select="handleValidatorSelect"
              class="dashboard-header__dropdown"
            />
          </template>
        </UISelect>
        <UIButton v-if="!userStore.isUserAuth" class="dashboard-header__sign-in" color_type="outline" size="large" @cta="userStore.isOpenLoginModal">
          <template #right-icon>
            <SVGLogIn />
          </template>
          Sign in/Sign up
        </UIButton>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../UI/UIButton.vue";
import SVGRefresh from "../SVG/SVGRefresh.vue";
import UISelect from "../UI/UISelect.vue";
import {computed, ref} from "vue";
import {useRoute, useRouter} from "vue-router";
import UIDropdown from "../UI/UIDropdown.vue";
import {useUserStore} from "../../store/userStore.js";
import SVGProfileIcon from "../SVG/SVGProfileIcon.vue";
import SVGLogOut from "../SVG/SVGLogOut.vue";
import SVGLogIn from "../SVG/SVGLogIn.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import {useHeaderRefreshStore} from "../../store/headerRefreshStore.js";
import ProfileImage from "../../../public/images/default-avatar.webp";
import CookieManager from "../../helpers/cookieManager.js";
import SVGArrowPrevious from "../SVG/SVGArrowPrevious.vue";

const emits = defineEmits(['handlePageDataRefresh'])
const userStore = useUserStore();
const modalsStore = useModalsStore();
const headerRefreshStore = useHeaderRefreshStore();
const route = useRoute();
const router = useRouter();
const isRPCOptionsOpen = ref(false);
const isDropDownOpen = ref(false);
const loginData = computed(() => {
  if (userStore.userData) {
    return {
      email: userStore.userData.email || '',
    }
  } else {
    return {}
  }
})
const options = [
  {name: 'Profile', val: 'profile', image: ProfileImage},
  {name: 'Logout', val: 'logout', svg: SVGLogOut},
];
const rpcOptions = [
  {label: 'Helius', val: 'helius'},
  {label: 'Helius', val: 'helius1'},
  {label: 'Helius', val: 'helius2'},
]

const isRefreshVisible = computed(() => {
  const pages = ['Home', 'TokenCreate', 'TokenVolumeMaker', 'TokenHistory', 'DashboardNotFound', 'LiquidityPool', 'LiquidityBurn'];

  if (!userStore.isUserAuth) {
    return false;
  }

  return !pages.includes(route.name);
})
const isBackButtonVisible = computed(() => {
  const pages = [''];

  if (route.name === 'MarketTargetPullUpCreate' && route.params?.campaign_id !== 'create') {
    return true;
  }

  return pages.includes(route.name);
})
const clearUserData = () => {
  CookieManager.removeItem("access_token");
  CookieManager.removeItem("refresh_token");
  userStore.setUserData(null);
  userStore.isUserAuth = false;
  router.push({name: 'Home'})
}

const handleValidatorSelect = (option) => {
  if (option.val === 'logout') {
    clearUserData();
  }

  isDropDownOpen.value = false;
}

const handleHeaderRefresh = async () => {
  await headerRefreshStore.runRefreshHandler();
  emits('handlePageDataRefresh');
}
</script>
<style scoped lang="scss">
.dashboard-header {
  width: 100%;
  display: flex;
  min-height: 64px;
  background: #F9FAFB;
  border-bottom: 1px solid #D1D5DB;
  position: relative;
  z-index: 100;

  &__inner {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    padding: 0 24px;
  }

  &__back {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #374151;
    cursor: pointer;
  }

  &__dropdown {
    ::v-deep(.icon) {
      background: transparent;
      & img {
        width: 100%;
        height: 100%;
      }
    }

    ::v-deep(path) {
      fill: #0F1729;
    }
  }

  &__sign-in {
    margin-left: 10px;
    background: #FFF !important;
  }

  &__right {
    margin-left: auto;
    display: flex;
    align-items: center;

    & .divider {
      height: 25px;
      width: 1px;
      background: #D1D5DB;
      margin: 0 8px;
    }
  }

  &__profile {
    height: 36px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding-left: 12px;
    padding-right: 8px;
    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #FFF;
    transition: 0.3s ease;
    overflow: hidden;
    width: 133px;
    max-width: 133px;
    margin-left: 12px;

    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

    &:hover {
      box-shadow: 0 0 0 3px var(--focus-ring, #D1D5DB);
    }

    & img {
      width: 20px;
      height: 20px;
    }

    & span {
      color: #030712;
      font-family: "Geist Mono", sans-serif;
      font-size: 16px;
      font-style: normal;
      font-weight: 600;
      line-height: 150%;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  &__rpc {
    display: flex;
    align-items: center;
    gap: 8px;

    & .label {
      font-weight: 500;
      width: 53px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    &_dropdown {
      position: relative;
      width: 140px;
    }

    &_options {
      position: absolute;
      z-index: 10;
      top: calc(100% + 8px);
      right: 0;
    }
  }
}

.rpc-options {
  border-radius: 8px;
  border: 1px solid #E5E7EB;
  background: var(--general-input, #FFF);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.10), 0 8px 10px -6px rgba(0, 0, 0, 0.10);
  width: 235px;

  &__top {
    height: 35px;
    padding: 0 12px;
    display: flex;
    align-items: center;
    border-bottom: 1px solid #E5E7EB;
    color: #374151;
    font-weight: 500;
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px 0;
  }

  &__item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
    transition: .3s ease;

    &.selected {
      background: #F3F4F6;
      height: 45px;
    }

    &_info {
      display: flex;
      flex-direction: column;

      & .text {
        color: #6B7280;
        font-size: 11px;
        font-style: normal;
        font-weight: 400;
        line-height: 150%;
        letter-spacing: 0.165px;
      }
    }
  }
}

@media (max-width: 1200px) {
  .dashboard-header {
    &__right {
      display: none;
    }
  }
}
</style>
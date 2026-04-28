<template>
  <div :class="['token-create']">
    <PageLoading v-if="isPageLoading"/>
    <div v-show="!isPageLoading" class="token-create__inner">
      <div class="token-create__desktop">
        <div class="token-create__desktop_left">
          <CreateTokenTop
            :data="tokenData"
            @handle-image-save="handleLogoChoose"
          />
          <UIAlert
            class="token-create__alert"
            text="You can view and manage your token inside your wallet."
            status="blue"
            :icon="SVGAlertInfo"
          />
          <div class="token-create__btns">
            <UIButton
              class="token-create__start"
              size="large"
              color_type="primary"
              @cta="handleTokenCreate"
              :is_disabled="(isChangesSaving)"
            >
              {{ startButtonText }}
            </UIButton>
          </div>
        </div>
        <div class="token-create__desktop_right">
          <UISectionTitleWithBorder>Token Custom Setting</UISectionTitleWithBorder>
          <div class="token-create__desktop_settings">
            <div
              v-for="item in settings"
              :key="item.key"
              class="setting"
            >
              <div class="setting__top">
                <div class="setting__left">
                  <span class="paragraph-small medium">{{ item.label }}</span>
                  <p class="paragraph-mini regular grey">{{ item.text }}</p>
                </div>
                <UIRoundToggle
                  v-model:is-active="tokenData[item['key']]"
                />
              </div>

              <transition name="pumpfun-socials">
                <div v-if="item.key === 'social_links_toggle' && tokenData.social_links_toggle"
                     class="setting__social-links">
                  <UIBaseInput
                    v-for="key in Object.keys(socialLinksInfo)"
                    :key="key"
                    size="large"
                    v-model="tokenData.social_links[key]"
                    :label="socialLinksInfo[key].label"
                    :placeholder="socialLinksInfo[key].placeholder"
                    class="social-links__item"
                  >
                    <template #icon-left>
                      <component class="icon" v-if="socialLinksInfo[key].icon" :is="socialLinksInfo[key].icon"/>
                    </template>
                  </UIBaseInput>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </div>
      <MobileAdaptsNotification class="token-create__mobile"/>
    </div>
  </div>
</template>
<script setup>
import {computed, markRaw, onMounted, ref, watch} from "vue";
import UIButton from "../../components/UI/UIButton.vue";
import UIAlert from "../../components/UI/UIAlert.vue";
import SVGAlertInfo from "../../components/SVG/SVGAlertInfo.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import {useRoute, useRouter} from "vue-router";
import {errorToast} from "../../helpers/index.js";
import PageLoading from "../../components/UI/PageLoading.vue";
import {useToastStore} from "../../store/toastStore.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import {useSmartCampaignsStore} from "../../store/smartCampaignsStore.js";
import CreateTokenTop from "../../components/Token/CreateToken/CreateTokenTop.vue";
import UIRoundToggle from "../../components/UI/UIRoundToggle.vue";
import UIBaseInput from "../../components/UI/UIBaseInput.vue";
import SVGTwitter from "../../components/SVG/SVGTwitter.vue";
import SVGDiscord from "../../components/SVG/SVGDiscord.vue";
import SVGWebsite from "../../components/SVG/SVGWebsite.vue";
import SVGTelegram from "../../components/SVG/SVGTelegram.vue";

const route = useRoute();
const router = useRouter();
const toastStore = useToastStore();
const smartCampaignStore = useSmartCampaignsStore();
const userStore = useUserStore();
const tokenData = ref({
  name: '',
  ticker: '',
  decimals: 9,
  supply: 0,
  logo: '',
  description: '',
  ownership: '',
  social_links: {
    twitter: '',
    discord: '',
    website: '',
    telegram: '',
  },
  fixed_supple: false,
  revoke_freeze: false,
  immutable: false,
  social_links_toggle: false,
})
const isChangesSaving = ref(false);
const isPageLoading = ref(true)
const startButtonText = computed(() => {
  return isChangesSaving.value ? 'Creating...' : 'Create token';
})
const socialLinksInfo = {
  twitter: {label: "X (Twitter)", placeholder: "Add X", icon: markRaw(SVGTwitter)},
  discord: {label: "Discord", placeholder: "Add Discord", icon: markRaw(SVGDiscord)},
  website: {label: "Website", placeholder: "Add Website", icon: markRaw(SVGWebsite)},
  telegram: {label: "Telegram", placeholder: "Add Telegram", icon: markRaw(SVGTelegram)},
}
const settings = [
  {
    key: 'fixed_supple',
    label: 'Fixed Supply', 
    text: 'Permanently disables minting of new tokens. Total supply becomes fixed and cannot be increased by anyone, including the creator.'
  },
  {
    key:  'revoke_freeze',
    label: 'Revoke Freeze', 
    text: 'Removes the ability to freeze token accounts. Ensures holders can always transfer or sell their tokens without restrictions.'
  },
  {
    key:  'immutable',
    label: 'Immutable',
    text: 'Locks token metadata forever. Name, symbol, logo, and description cannot be changed after this action is executed.'
  },
  {
    key: 'social_links_toggle',
    label: 'Social Links',
    text: 'Add Twitter, Telegram, and website URLs to token metadata for community discovery and verification.'
  },
];

const handlePageRefresh = async (isRefreshing = false, isAuth = false) => {
  isPageLoading.value = true;

  if (!isAuth) {
    smartCampaignStore.clearStore();
  }

  if (!userStore.isUserAuth) {
    try {
    } catch (e) {
      errorToast(e.response.data)
    } finally {
      isPageLoading.value = false;
    }

    return
  }

  try {

    if (isRefreshing) {
      toastStore.success({text: "Page is refreshed"})
    }
  } catch (e) {
    errorToast(e.response.data)
  } finally {
    isPageLoading.value = false;
  }
}

const handleLogoChoose = (data) => {
  if (!data) return;

  tokenData.value.logo = data.url || '';
  tokenData.value.logo_file = data.file || null;
}

const handleTokenCreate = async () => {

}

useHeaderRefresh(() => handlePageRefresh(true));

watch(() => [route.name, route.params.campaign_id], async () => {
  smartCampaignStore.clearStore();

  await handlePageRefresh();
})

watch(() => userStore.isUserAuth, async (newVal) => {
  if (newVal) {
    await handlePageRefresh(false, true);
  }
})

onMounted(async () => {
  if (route.params?.campaign_id !== 'create' && !userStore.isUserAuth) {
    await router.push({params: {campaign_id: 'create'}});
  }
  await handlePageRefresh();
});
</script>
<style scoped lang="scss">
.token-create {
  &__block {
    margin-top: 32px;
  }

  &__btns {
    display: flex;
    gap: 12px;

  }

  &__back {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #374151;
    cursor: pointer;
  }

  &__start {
    margin: 32px 0;
  }

  &__estimate {
    margin-top: 32px;
  }

  &__alert {
    margin-top: 32px;
  }

  &__desktop {
    width: 100%;
    display: flex;
    gap: 10px;
    justify-content: space-between;

    &_left {
      max-width: 669px;
      width: 100%;
    }

    &_right {
      width: 100%;
      max-width: 371px;
    }

    &_settings {
      margin-top: 20px;
      display: flex;
      flex-direction: column;
      gap: 12px;


      & .setting {
        border-radius: 8px;
        border: 1px solid #D1D5DB;
        padding: 16px;
        display: flex;
        gap: 16px;
        flex-direction: column;
        justify-content: space-between;

        &__top {
          display: flex;
          justify-content: space-between;
        }

        &__left {
          display: flex;
          flex-direction: column;
          gap: 8px;
          max-width: 227px;
        }

        &__social-links {
          width: 100%;
          display: flex;
          flex-direction: column;
          gap: 12px;

          ::v-deep(.base-input__input) {
            gap: 6px;
          }

          & .icon {
            min-width: 16px;
            max-width: 16px;
            min-height: 16px;
            max-height: 16px;

            ::v-deep(path) {
              fill: #9CA3AF;
            }
          }
        }
      }
    }
  }

  &__mobile {
    display: none;
  }

  &.pull-down {
    & .token-create {
      &__block {
        &.transaction-setup, &.target {
          opacity: .5;
          pointer-events: none;
        }
      }
    }
  }
}

.modal-stop-campaign {
  max-width: 480px;
}

@media (max-width: 1200px) {
  .token-create {
    &__desktop {
      display: none;
    }

    &__mobile {
      margin: 16px auto;
      padding: 0 16px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>
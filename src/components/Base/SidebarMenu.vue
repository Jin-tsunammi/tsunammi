<template>
  <div class="sidebar-menu">
    <div class="sidebar-menu__inner">

      <div class="sidebar-menu__list">
        <div
          v-for="category in menu"
          :key="category.label"
          class="sidebar-menu__block"
        >
          <div class="sidebar-menu__block_category" @click="toggleCategory(category)">
            <component v-if="category.icon" :is="category.icon" class="icon"/>
            <span class="paragraph-small">{{category.label}}</span>
            <button :class="{open: !category.is_open}">
              <SVGSmallArrowDown />
            </button>
          </div>
          <div :class="['sidebar-menu__block_pages-container', {hidden: !category.is_open}]">
            <div :class="['sidebar-menu__block_pages']">
              <div class="line"></div>
              <router-link
                v-for="page in category.pages"
                :key="page.label"
                :to="{name: page.name, params: page.params}"
                :class="['sidebar-menu__block_page', {active: page.children && page.children.includes(route.name)}]"
                @click="closeMobileSideBar"
              >
                {{page.label}}
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <div class="sidebar-menu__bottom">
        <ul class="sidebar-menu__bottom_socials">
          <li>
            <router-link to=""><SVGDiscord/></router-link>
          </li>
          <li>
            <router-link to=""><SVGTwitter/></router-link>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
<script setup>
import {ref} from "vue";
import SVGHand from "../SVG/SVGHand.vue";
import SVGMarket from "../SVG/SVGMarket.vue";
import SVGWallet from "../SVG/SVGWallet.vue";
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";
import SVGDiscord from "../SVG/SVGDiscord.vue";
import SVGTwitter from "../SVG/SVGTwitter.vue";
import SVGLogOut from "../SVG/SVGLogOut.vue";
import CookieManager from "../../helpers/cookieManager.js";
import {useSidebarStore} from "../../store/sidebarStore.js";
import {useRoute} from "vue-router";
import SVGTokenIcon from "../SVG/SVGTokenIcon.vue";
import SVGRocket from "../SVG/SVGRocket.vue";

const sidebarStore = useSidebarStore();
const route = useRoute();

const menu = ref([
  {
    label: 'Token Management',
    icon: SVGTokenIcon,
    pages: [
      {
        label: 'Create token',
        name: 'TokenCreate'
      },
      {
        label: 'Liquidity Pool',
        name: 'LiquidityPool'
      },
      {
        label: 'Liquidity Burn',
        name: 'LiquidityBurn'
      },
      {
        label: 'Token activity history',
        name: 'TokenHistory'
      }
    ],
    is_open: true,
  },
  {
    label: 'Wallet Management',
    icon: SVGWallet,
    pages: [
      {
        label: 'Wallet Pools',
        name: 'WalletsProjects',
        children: ['WalletsSelectedProject']
      },
      {
        label: 'Connect CEX API',
        name: 'WalletsConnectCexApi'
      },
      {
        label: 'Distribute Funds from CEX',
        name: 'WalletsTopUpCex'
      },
      {
        label: 'History',
        name: 'WalletsHistory'
      }
    ],
    is_open: true,
  },
  {
    label: 'Market Operations',
    icon: SVGRocket,
    pages: [
      {
        label: 'Price Boost',
        name: 'MarketTargetPullUpCreate',
        params: {campaign_id: 'create'}
      },
      {
        label: 'Price Drop',
        name: 'MarketTargetDrop',
        params: {campaign_id: 'create'}
      },
      {
        label: 'Smart Buyback',
        name: 'MarketSmartBuyback'
      },
      {
        label: 'Ops history',
        name: 'MarketHistory'
      },
    ],
    is_open: true,
  },
])

const toggleCategory = (category) => {
  category.is_open = !category.is_open
}

const closeMobileSideBar = () => {
  if (sidebarStore.isMobileMenuOpen) {
    sidebarStore.isMobileMenuOpen = false;
  }
}
</script>
<style scoped lang="scss">
.sidebar-menu {
  height: 100%;
  &__inner {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 60px;
  }

  &__block {
    display: flex;
    flex-direction: column;

    &_category {
      display: flex;
      align-items: center;
      cursor: pointer;
      height: 32px;
      padding: 0 12px;

      & .icon {
        ::v-deep(path) {
          fill: #9CA3AF;
        }
      }

      & span {
        color: #FFF;
        font-weight: 400;
        display: block;
        margin-left: 8px;
      }

      & button {
        background: transparent;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-left: auto;

        & svg {
          transition: .3s ease;
          transform: rotate(180deg);
        }

        &.open {
          & svg {
            transform: rotate(0);
          }
        }
      }
    }

    &_pages {
      display: flex;
      flex-direction: column;
      gap: 3px;
      margin-left: 18px;
      padding-left: 9px;
      margin-top: 12px;
      position: relative;

      &-container {
        max-height: 170px;
        overflow: hidden;
        transition: .3s ease;

        &.hidden {
          max-height: 0;
        }
      }

      & .line {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        width: 1px;
        background: #9CA3AF;
      }
    }

    &_page {
      height: 29px;
      padding: 4px;
      display: flex;
      align-items: center;
      color: #9CA3AF;
      transition: .3s ease;
      border-radius: 6px;

      &:hover, &.router-link-active, &.active {
        background: #EA580C;
        color: #FFF;
      }
    }
  }

  &__bottom {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: 48px;
    padding: 0 12px;

    &_socials {
      display: flex;
      align-items: center;
      gap: 16px;

      & a {
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
  }
}
</style>
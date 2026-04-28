<template>
  <div class="home">
    <div class="home__inner">
      <div class="home__main-block">
        <img class="background" src="../../public/images/main-background.webp" alt="Solana icon">

        <div class="home__main-block_content">
          <div class="home__main-block_left">
            <h1 class="">Tsunammi Tools</h1>
            <p>
              The simplest way to create, launch and manage tokens on
              <img src="../../public/images/solana-sol-logo.webp" alt="Solana icon">
              <span>Solana</span>
            </p>
            <div class="paragraph-medium bold grey">no code, no expensive <span class="regular">team needed.</span></div>
          </div>
          <SVGLogoTsunammi class="logo" />
        </div>
      </div>

      <div class="home__top">
        <div
          v-for="block in topContent"
          :key="block.block"
          :class="['home__top-block', block.block]"
        >
          <div class="home__top-block_left">
            <div class="gradient"></div>
            <img v-if="block.image" :src="block.image" alt="Image">
            <h4 class="heading-4">{{block.title}}</h4>
            <p class="paragraph-small">{{block.text}}</p>
          </div>

          <div class="home__top-block_right">
            <router-link
              v-for="link in block.links"
              :key="link.page"
              :to="{name: link.page, params: getPageParams(link)}"
              class="home__top-block_link"
            >
              <SVGArrowLink class="arrow"/>
              <span class="paragraph-small">{{link.title}}</span>
              <p class="paragraph-mini regular">{{link.text}}</p>
            </router-link>
          </div>
        </div>
      </div>

      <div class="home__mm">
        <h2 class="heading-2">Market Making for Everyone</h2>
        <div class="home__mm_links">
          <div
            class="home__mm_item"
            v-for="page in mmLinks"
            :key="page.label"
          >
            <router-link class="link" v-if="!page.is_empty" :to="{name: page.page, params: getPageParams(page)}">
              <SVGArrowLink class="arrow" />
              <div class="icon">
                <component v-if="page.svg" :is="page.svg" />
              </div>

              <div class="right">
                <span class="paragraph-medium">{{page.label}}</span>
                <span class="paragraph-small regular grey">{{page.text}}</span>
              </div>
            </router-link>

            <div v-else class="empty">
              <img src="../../public/images/Container.svg" alt="Grid">
              <h4 class="heading-4">More Control, Coming Soon</h4>
              <span class="paragraph-mini regular">Building new tools to expand your market control.</span>
            </div>
          </div>
        </div>
      </div>

      <div class="home__bottom">
        <img class="left-img" src="../../public/images/login-left-abstract.webp" alt="Image">
        <img class="right-img" src="../../public/images/login-right-abstract.webp" alt="Image">
        <div class="home__bottom_content">
          <div class="left">
            <h2 class="heading-2">Building in Public</h2>
            <p class="paragraph-medium regular">Tsunammi is building the easiest way to run market making — from launch to liquidity control<br>Get early access to new tools as we ship.</p>
          </div>
          <a href="https://x.com/tsunammitools" target="_blank">Follow on X</a>
        </div>
      </div>
    </div>
    <Modals/>

  </div>
</template>
<script setup>
import SVGWalletPool from "../components/SVG/SVGWalletPool.vue";
import SVGConnectApi from "../components/SVG/SVGConnectApi.vue";
import SVGFundsFromCEX from "../components/SVG/SVGFundsFromCEX.vue";
import SVGPriceBoost from "../components/SVG/SVGPriceBoost.vue";
import SVGPriceDrop from "../components/SVG/SVGPriceDrop.vue";
import SVGBuyback from "../components/SVG/SVGBuyback.vue";
import SVGOpsHistory from "../components/SVG/SVGOpsHistory.vue";
import SVGWalletsHistory from "../components/SVG/SVGWalletsHistory.vue";
import SVGClockBack from "../components/SVG/SVGClockBack.vue";
import SVGCreateToken from "../components/SVG/SVGCreateToken.vue";
import SVGArrowLink from "../components/SVG/SVGArrowLink.vue";
import SVGLogoTsunammi from "../components/SVG/SVGLogoTsunammi.vue";
import Modals from "../components/UI/Modals.vue";
import MMTokenImage from "../../public/images/main-token-image.webp";
import MMImage from "../../public/images/main-mm-image.webp";
import SVGPumpFun from "../components/SVG/SVGPumpFun.vue";

const topContent = [
  {
    block: 'token',
    title: 'Manage Token',
    text: 'Create tokens in seconds with a simple step-by-step flow.',
    image: MMTokenImage,
    links: [
      {
        title: 'Create Token',
        text: 'Launch your SPL token  — in just a few clicks.',
        page: 'TokenCreate',
      },
      {
        title: 'Multiwallet Management',
        text: 'Operate dozens or hundreds of wallets from a single dashboard.',
        page: 'WalletsProjects',
      },
      {
        title: 'Check Token Activity History',
        text: 'Track all token operations, transactions, metrics in one place.',
        page: 'TokenHistory',
      },
    ]
  },
  {
    block: 'mm',
    title: 'Market Making',
    text: 'Powerful tools to control and stabilize your token price after launch.',
    image: MMImage,
    links: [
      {
        title: 'Price Boost',
        text: 'Trigger targeted buys to drive momentum when it matters most.',
        page: 'MarketTargetPullUpCreate',
        params: {campaign_id: 'create'},
      },
      {
        title: 'Price Drop',
        text: 'Gradually reduce price with controlled sell pressure.',
        page: 'MarketTargetDrop',
        params: {campaign_id: 'create'},
      },
      {
        title: 'Smart Buyback',
        text: 'Automatically support your token with intelligent buybacks.',
        page: 'MarketSmartBuyback',
        params: {campaign_id: 'create'},
      },
    ]
  }
];

const mmLinks = [
  {label: 'Create Wallet Pools', text: 'Generate hot wallets for MM ops.', page: 'WalletsProjects', svg: SVGWalletPool, is_empty: false},
  {label: 'Connect CEX API', text: 'Link CEX accounts for funding.', page: 'WalletsConnectCexApi', svg: SVGConnectApi, is_empty: false},
  {label: 'Distribute Funds from CEX', text: 'Auto-split liquidity across wallets.', page: 'WalletsTopUpCex', svg: SVGFundsFromCEX, is_empty: false},
  {label: 'Price Boost', text: 'Strategic buys to pump the price.', page: 'MarketTargetPullUpCreate', params: {campaign_id: 'create'}, svg: SVGPriceBoost, is_empty: false},
  {label: 'Price Drop', text: 'Controlled sells for soft landing.', page: 'MarketTargetDrop', params: {campaign_id: 'create'}, svg: SVGPriceDrop, is_empty: false},
  {label: 'Smart Buyback', text: 'Auto buyback on price triggers.', page: 'MarketSmartBuyback', params: {campaign_id: 'create'}, svg: SVGBuyback, is_empty: false},
  {label: 'MM History', text: 'Full log of all market operations.', page: 'MarketHistory', svg: SVGOpsHistory, is_empty: false},
  {label: 'Wallets Ops History', text: 'Track wallet funding and transfers.', page: 'WalletsHistory', svg: SVGWalletsHistory, is_empty: false},
  {label: 'Create Token', text: 'Deploy SPL tokens in seconds.', page: 'TokenCreate', svg: SVGCreateToken, is_empty: false},
  {label: 'Launch on PumpFun', text: 'Deploy token to PF launchpad.', page: 'LaunchPumpFun', svg: SVGPumpFun, is_empty: false},
  // {label: 'Liquidity Pool', text: 'Create DEX pools for trading.', page: 'LiquidityPool', svg: SVGWaves, is_empty: false},
  // {label: 'Liquidity Burn', text: 'Lock LP tokens permanently.', page: 'LiquidityBurn', svg: SVGBurn, is_empty: false},
  {label: 'Token Activity History', text: 'Complete token lifecycle data.', page: 'TokenHistory', svg: SVGClockBack, is_empty: false},
  {label: 'empty', text: '', page: '', svg: null, is_empty: true},
]

const getPageParams = (page) => {
  if (!page || !page?.params) return null;

  return page.params;
}
</script>
<style scoped lang="scss">
.home {
  &__inner {
    padding-bottom: 67px;
    display: flex;
    flex-direction: column;

    & .home {
      &__main-block_content, &__top, &__mm, &__bottom {
        align-self: center;
        margin-inline: auto;
        max-width: 1083px;
        width: calc(100% - 128px);
      }
    }
  }
  &__main-block {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 359px;
    padding: 0 64px;

    & .background {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
    }

    &_content {
      display: flex;
      align-items: center;
      justify-content: space-between;
      width: 100%;
    }

    & .logo {
      position: relative;
      z-index: 3;
    }

    &_left {
      position: relative;
      z-index: 2;
      & h1 {
        color: #030712;
        font-family: "Geist Mono", sans-serif;
        font-size: 56px;
        font-style: normal;
        font-weight: 700;
        line-height: 120%;
        letter-spacing: -1.12px;
        margin-bottom: 12px;
      }

      & p {
        color: #030712;
        font-family: Geist, sans-serif;
        font-size: 32px;
        font-style: normal;
        font-weight: 400;
        line-height: 120%;
        letter-spacing: -0.64px;
        max-width: 582px;
        margin-bottom: 28px;

        & span {
          font-weight: 600;
        }

        & img {
          position: relative;
          top: 5px;
          width: 30px;
          height: 30px;
          margin-right: 6px;
        }
      }
    }
  }

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    margin-top: 48px;

    &-block {
      display: flex;
      gap: 12px;
      max-width: 482px;

      &.token {
        & .gradient {
          background: linear-gradient(180deg, rgba(217, 217, 217, 0.00) 3.93%, #030712 38.47%);
        }

        &:has(.home__top-block_left:hover), &:has(.home__top-block_link:hover) {
          & .gradient {
            height: 120%;
          }
        }
      }

      &.mm {
        & .gradient {
          background: linear-gradient(180deg, rgba(217, 217, 217, 0.00) 15.21%, #F97316 37.37%);
        }

        &:has(.home__top-block_left:hover), &:has(.home__top-block_link:hover) {
          & .gradient {
            height: 120%;
          }
        }
      }

      &_left {
        position: relative;
        display: flex;
        flex-direction: column;
        gap: 8px;
        width: 220px;
        border-radius: 8px;
        overflow: hidden;
        min-height: 100%;
        padding: 0 25px 25px;
        justify-content: flex-end;

        & .gradient {
          position: absolute;
          z-index: 2;
          top: 0;
          left: 0;
          width: 100%;
          height: 200%;
          transition: .3s;
        }

        & img {
          position: absolute;
          left: 0;
          top: 0;
        }

        & h4, & p {
          position: relative;
          z-index: 3;
        }

        & h4 {
          color: #FFF;
        }

        & p {
          color: #E2E8F0;
        }
      }

      &_right {
        display: flex;
        flex-direction: column;
        gap: 12px;
        max-width: 250px;
      }

      &_link {
        position: relative;
        padding: 12px;
        background: #FFF;
        border-radius: 6px;
        border: 1px solid transparent;

        & .arrow {
          position: absolute;
          top: 12px;
          right: 12px;
          display: none;
        }

        & span {
          display: block;
          padding-right: 20px;
        }

        & p {
          color: #64748B;
        }

        &:hover {
          border: 1px solid #F97316;
          box-shadow: 0 4px 4px 0 rgba(0, 0, 0, 0.25);

          & .arrow {
            display: block;
          }
        }
      }
    }
  }

  &__mm {
    margin-top: 64px;
    display: flex;
    flex-direction: column;
    gap: 28px;

    &_links {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 24px;
    }

    &_item {
      height: 88px;


      & .link {
        position: relative;
        border-radius: 8px;
        display: flex;
        align-items: center;
        gap: 16px;
        padding: 16px;
        background: #FFF;
        transition: .3s;
        border: 1px solid transparent;

        & .arrow {
          position: absolute;
          top: 16px;
          right: 16px;
          display: none;
        }

        &:hover {
          border: 1px solid #F97316;
          box-shadow: 0 4px 4px 0 rgba(0, 0, 0, 0.25);

          & .arrow {
            display: block;
          }
        }
      }

      & .right {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      & .icon {
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 50%;
        min-width: 56px;
        max-width: 56px;
        min-height: 56px;
        max-height: 56px;
        background: #F3F4F6;
      }

      & .empty {
        border-radius: 8px;
        position: relative;
        background: #030712;
        padding: 19px 24px;
        overflow: hidden;

        & h4 {
          position: relative;
          z-index: 2;
          color: #FFF;
          margin-bottom: 8px;
        }

        & span {
          color: rgba(255, 255, 255, 0.8);
        }

        & img {
          position: absolute;
          bottom: 0;
          right: 0;
        }
      }
    }
  }

  &__bottom {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid #D1D5DB;
    background: transparent;
    padding: 39px 0;
    margin-top: 64px;

    & img {
      height: 282px;
      width: auto;
      position: absolute;
      top: 50%;
      transform: translateY(-50%);
    }

    & .left-img {
      left: 0;
    }

    & .right-img {
      right: 0;
    }

    &_content {
      position: relative;
      z-index: 2;
      display: flex;
      align-items: center;
      gap: 32px;
      padding: 64px;
      background: #FFF;
      border-radius: 8px;

      & .left {
        display: flex;
        flex-direction: column;
        gap: 16px;
        max-width: 584px;

        & p {
          color: #475569;
        }
      }

      & a {
        min-width: max-content;
        color: #FFF;
        font-family: Inter, sans-serif;
        font-size: 14px;
        font-style: normal;
        font-weight: 500;
        line-height: 24px;
        border-radius: 6px;
        background: #0F172A;
        padding: 8px;

        &:hover, &:active {
          background: #374151;
        }

        &:focus {
          background: #111827;
          box-shadow: 0 0 0 3px #D1D5DB;
        }
      }
    }
  }
}

@media (max-width: 1500px) {
  .home {
    &__mm {
      &_links {
        grid-template-columns: repeat(2, 1fr);
      }
    }
  }
}
</style>
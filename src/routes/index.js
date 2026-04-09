import {createRouter, createWebHistory} from "vue-router";
import CookieManager from "../helpers/cookieManager.js";
import {useToken} from "../composable/useToken.js";
import {useUserStore} from "../store/userStore.js";

export const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'Home',
            component: () => import(/* webpackChunkName: "Home" */ "../pages/Index.vue"),
        },
        {
            path: '/profile',
            name: 'DashboardProfile',
            component: () => import(/* webpackChunkName: "Home" */ "../pages/profile/Index.vue"),
        },
        // Wallets
        {
            path: '/wallets/projects',
            name: 'WalletsProjects',
            component: () => import("../pages/wallets/CreateProject.vue"),
        },
        {
            path: '/wallets/projects/:project_id',
            name: 'WalletsSelectedProject',
            component: () => import("../pages/wallets/Manage.vue"),
        },
        {
            path: '/wallets/top-up-cex',
            name: 'WalletsTopUpCex',
            component: () => import("../pages/wallets/TopUpCEX.vue"),
        },
        {
            path: '/wallets/connect-cex-api',
            name: 'WalletsConnectCexApi',
            component: () => import("../pages/wallets/ConnectCexApi.vue"),
        },
        {
            path: '/wallets/history',
            name: 'WalletsHistory',
            component: () => import("../pages/wallets/History.vue"),
        },

        // Market Making
        {
            path: '/market-making/target-pull-up/:campaign_id',
            name: 'MarketTargetPullUpCreate',
            component: () => import("../pages/market-making/TargetPullUp.vue"),
        },
        {
            path: '/market-making/target-drop/:campaign_id',
            name: 'MarketTargetDrop',
            component: () => import("../pages/market-making/TargetPullUp.vue"),
        },
        {
            path: '/market-making/smart-buyback',
            name: 'MarketSmartBuyback',
            component: () => import("../pages/market-making/SmartBuyback.vue"),
        },
        {
            path: '/market-making/history',
            name: 'MarketHistory',
            component: () => import("../pages/market-making/History.vue"),
        },
        {
            path: '/market-making/transactions/:campaign_id',
            name: 'MarketTransactions',
            component: () => import("../pages/market-making/Transactions.vue"),
        },

        // Token
        {
            path: '/token/create-token',
            name: 'TokenCreate',
            component: () => import("../pages/token/CreateToken.vue"),
        },
        {
            path: '/token/liquidity-pool',
            name: 'LiquidityPool',
            component: () => import("../pages/token/LiquidityPool.vue"),
        },
        {
            path: '/token/liquidity-burn',
            name: 'LiquidityBurn',
            component: () => import("../pages/token/LiquidityBurn.vue"),
        },
        {
            path: '/token/history',
            name: 'TokenHistory',
            component: () => import("../pages/token/History.vue"),
        },

        //Not found
        {
            path: '/:pathMatch(.*)*',
            name: 'DashboardNotFound',
            component: () => import('../pages/not-found/Index.vue'),
        },
    ],
})

router.beforeEach(async (to, from, next) => {
    const userStore = useUserStore();
    const refreshTokenData = CookieManager.getItem('access_token');
    const accessToken = CookieManager.getItem('refresh_token');
    const {refreshToken} = useToken();

    if (!accessToken) {
        if (refreshTokenData) {
            try {
                await refreshToken();
                userStore.isUserAuth = true;

                if (!userStore.userData) {
                    await userStore.getUserData();
                }
            } catch (err) {
                console.error('Refresh token failed:', err);
                userStore.isUserAuth = false;
            }
        } else {
            userStore.isUserAuth = false;
        }
    } else {
        userStore.isUserAuth = true;

        if (!userStore.userData) {
            await userStore.getUserData();
        }
    }

    next();
});


import {defineStore} from "pinia";
import {computed, ref} from "vue";
import {errorToast} from "../helpers/index.js";
import {cloneDeep} from "lodash";
import {DeleteCampaign, GetAllActiveCampaigns, GetAllCampaigns, GetCampaignByID} from "../api/api.js";
import {useTokensStore} from "./tokensStore.js";
import {useRoute} from "vue-router";

export const useCampaignsStore = defineStore('campaigns', () => {
    const route = useRoute();
    const tokensStore = useTokensStore();
    const DEFAULT_CAMPAIGN_DATA = {
        source_token_mint: "",
        dest_token_mint: "",
        project_id: null,
        budget: 0,
        budget_percent: 0,
        slippage: 5,
        goal_percentage_change: 0,
        parallel_transactions_amount: 1,
        min_transactions_budget: 0,
        max_transactions_budget: 0,
        priority_fee: 0,
        provider_id: 1,
        min_time_between_transactions: 1000000000, // nanosec
        max_time_between_transactions: 3000000000, // nanosec
        transaction_speed: "fast",
        using_jito: false,
    };

    const API_TO_CAMPAIGN = {
        token_mint_from: "source_token_mint",
        token_mint_to: "dest_token_mint",
        goal_percent_change: "goal_percentage_change",
        slippage_bps: (value) => ({ slippage: value != null ? value / 100 : 0.25 })
    };

    const campaign = ref(cloneDeep(DEFAULT_CAMPAIGN_DATA));
    const campaignDataBeforeChange = ref(null);
    const allCampaigns = ref([]);
    const activeCampaigns = ref([]);
    const selectedToken = ref(null);
    const isCampaignDataChanged = computed(() => {
        if (!campaignDataBeforeChange.value) return false;

        const before = campaignDataBeforeChange.value;
        const current = campaign.value;

        for (const key of Object.keys(before)) {
            if (current[key] !== before[key]) {
                return true;
            }
        }
        return false;
    });

    const setSelectedToken = (data) => {
        if (!data) return;

        selectedToken.value = data;
    }

    const getAllCampaigns = async (params = null) => {
        try {
            const resp = await GetAllCampaigns(params);
            allCampaigns.value = resp.data;
        } catch (e) {
            errorToast(e?.response?.data || '')
        }
    }

    const getAllActiveCampaigns = async (params = null) => {
        try {
            const resp = await GetAllActiveCampaigns(params);
            activeCampaigns.value = resp.data?.campaign_summary || [];

            const address = route.name === 'MarketTargetPullUpCreate' ? 'token_mint_to' : 'token_mint_from';

            await tokensStore.updateSolTokensData( resp.data?.campaign_summary || [], address);
        } catch (e) {
            errorToast(e?.response?.data || '')
        }
    }

    const getCampaign = async (id) => {
        if (!id) return;

        try {
            const resp = await GetCampaignByID(id);
            if (resp.data) {
                const data = resp.data;

                for (const [apiKey, value] of Object.entries(data)) {
                    const mapper = API_TO_CAMPAIGN[apiKey];
                    if (mapper) {
                        if (typeof mapper === "function") {
                            const mapped = mapper(value);
                            for (const [k, v] of Object.entries(mapped)) {
                                if (Object.hasOwn(DEFAULT_CAMPAIGN_DATA, k)) {
                                    campaign.value[k] = v;
                                }
                            }
                        } else if (Object.hasOwn(DEFAULT_CAMPAIGN_DATA, mapper)) {
                            campaign.value[mapper] = value;
                        }
                    } else if (Object.hasOwn(DEFAULT_CAMPAIGN_DATA, apiKey)) {
                        campaign.value[apiKey] = value;
                    }
                }

                campaignDataBeforeChange.value = cloneDeep(campaign.value);

                return resp.data;
            }
        } catch (e) {
            errorToast(e.response.data);
        }
    }

    const handleStopCampaign = async (id, params) => {
        if (!id) return;

        try {
            await DeleteCampaign(id);
            await getAllActiveCampaigns(params);
        } catch (e) {
            errorToast(e?.response?.data || '')
        }
    }

    const returnCampaignChanged = () => {
        campaign.value = cloneDeep(campaignDataBeforeChange.value);
    }

    const clearStore = () => {
        campaign.value = cloneDeep(DEFAULT_CAMPAIGN_DATA);
        allCampaigns.value = [];
        activeCampaigns.value = [];
        campaignDataBeforeChange.value = null;
        selectedToken.value = null;
    }

    return {
        campaign,
        allCampaigns,
        activeCampaigns,
        isCampaignDataChanged,
        campaignDataBeforeChange,
        selectedToken,

        clearStore,
        getAllActiveCampaigns,
        getAllCampaigns,
        handleStopCampaign,
        getCampaign,
        returnCampaignChanged,
        setSelectedToken,
    }
})
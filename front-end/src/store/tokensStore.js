import {defineStore} from "pinia";
import {ref} from "vue";
import {fetchSolTokenMetadata} from "../helpers/index.js";
import {SOLANA_MINT} from "../constants/const.js";
import axios from "axios";

export const useTokensStore = defineStore('tokens', () => {
    const solTokensData = ref({});
    const solTokensMints = ref();

    const updateSolTokensData = async(tokens, token_name='') => {
        if (!tokens || !tokens.length || !token_name) return;

        const newMints = new Set();
        newMints.add(SOLANA_MINT);

        tokens.forEach(token => {
            if (token[token_name]) {
                newMints.add(token[token_name]);
            }
        });

        const oldMints = new Set(solTokensMints.value);

        const wasEmptyBefore = solTokensMints.value?.size === 0;

        const addedTokens = [...newMints].filter(mint => !oldMints.has(mint));

        solTokensMints.value = newMints;

        if (wasEmptyBefore && newMints.size > 0) {
            await fetchSolTokenMetadata(newMints, solTokensData)
        }

        if (addedTokens.length > 0) {
            await fetchSolTokenMetadata(newMints, solTokensData)
        }
    }

    const getTokenPrice = async(tokenMint) => {
        if (!tokenMint) return;
        try {
            const resp = await axios.get('https://api.jup.ag/price/v3', {
                headers: {'x-api-key': import.meta.env.VITE_JUPITER_API_KEY},
                params: {ids: tokenMint}
            });

            if (resp.data) return resp.data;
            else return null;

        } catch (e) {
            console.error(e.response);
        }
    }

    return {
        solTokensData,
        updateSolTokensData,
        getTokenPrice,
    }
})
import {defineStore} from "pinia";
import {ref} from "vue";
import {
    CreateNewCEXApi,
    DeleteCEXApi,
    GetAllCEXApi,
    GetCEXApiByID,
} from "../api/api.js";
import {useToastStore} from "./toastStore.js";
import {useModalsStore} from "./modalsStore.js";
import {errorToast, formatText} from "../helpers/index.js";

export const useCEXApiStore = defineStore('cex', () => {
    const toastStore = useToastStore();
    const modalStore = useModalsStore();

    const allCEXApi = ref([]);

    const setAllCEXApi = (data) => allCEXApi.value = data;

    const getAllCEXApi = async() => {
        try {
            const resp = await GetAllCEXApi();

            allCEXApi.value = resp.data?.accounts || [];
        } catch (error) {
            console.error(error)
            errorToast(error.response.data)
            throw error;
        }
    }

    const updateCEX = async (cex_id) => {
        if (!cex_id) return;

        try {
            const resp = await GetCEXApiByID(cex_id);

            if (resp.data) {
                const originalIndex = allCEXApi.value.findIndex(elem => elem.id === resp.data.id);

                if (originalIndex > -1) {
                    allCEXApi.value[originalIndex] = resp.data;
                }
            }

            toastStore.success({text: 'Data refreshed'});
        } catch (error) {
            console.error(error)
            toastStore.error({text: formatText(error.response?.data)});
            throw error;
        }
    }

    const deleteCEXApi = async(cex_id) => {
        if (!cex_id) return;

        try {
            await DeleteCEXApi(cex_id);

            await getAllCEXApi();
            modalStore.closeModal();
            toastStore.success({text: 'CEX API removed'});
        } catch (error) {
            console.error(error)
            toastStore.error({text: formatText(error.response?.data)});
            throw error;
        }
    }

    const createNewCEXApi = async(data) => {
        if (!data) return;

        try {
            await CreateNewCEXApi(data);

            await getAllCEXApi();
            modalStore.modalData.title = `API Connection Successful`
            modalStore.modalData.type = 'create-confirmation'
            modalStore.modalData.action = 'confirmation';
        } catch (error) {
            console.error(error)
            toastStore.error({text: formatText(error.response?.data)});

            throw error;
        }
    }

    const clearStore = () => {
        allCEXApi.value = [];
    }

    return {
        allCEXApi,

        getAllCEXApi,
        setAllCEXApi,
        updateCEX,
        deleteCEXApi,
        createNewCEXApi,
        clearStore,
    }
})
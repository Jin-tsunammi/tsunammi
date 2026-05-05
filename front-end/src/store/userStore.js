import {defineStore} from "pinia";
import {ref} from "vue";
import {GetUser} from "../api/api.js";
import {useToastStore} from "./toastStore.js";
import {useModalsStore} from "./modalsStore.js";
import {useRouter} from "vue-router";

export const useUserStore = defineStore('user', () => {
    const toastStore = useToastStore();
    const modalsStore = useModalsStore();
    const router = useRouter();

    const isUserAuth = ref(false);
    const userData = ref(null);

    const setUserData = (data) => userData.value = data;

    const getUserData = async() => {
        if (!isUserAuth.value) return;

        try {
            const resp = await GetUser();

            userData.value = resp.data;
        } catch (error) {
            console.error(error)
            toastStore.error({text: 'Something went wrong'});
            throw error;
        }
    }

    const refreshProfileRequests = async(isPressedRefreshBtn=false) => {
        try {
            await getUserData();


            if (isPressedRefreshBtn) {
                toastStore.success({text: 'Page data refreshed'});
            }
        } catch (error) {
        }
    }

    const isOpenLoginModal = () => {
        if (!isUserAuth.value) {
            modalsStore.modalData.is_open = true;
            modalsStore.modalData.action = 'login';

            return true
        }

        return false;
    }

    return {
        userData,
        isUserAuth,

        setUserData,
        getUserData,
        refreshProfileRequests,
        isOpenLoginModal,
    }
})
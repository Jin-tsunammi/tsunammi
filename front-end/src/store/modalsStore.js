import {defineStore} from "pinia";
import {ref} from "vue";
import {cloneDeep} from "lodash";

export const useModalsStore = defineStore('modals', () => {
    const DEFAULT_MODAL_DATA = {
        is_open: false,
        title: '',
        mainText: '',
        additionalText: '',
        type: '',
        is_custom: false,
        action: 'configure', // configure | confirmation | login
        is_close_icon: true,
        icon: null,
    }

    const modalData = ref(cloneDeep(DEFAULT_MODAL_DATA));

    const openModal = () => modalData.value.is_open = true;
    const closeModal = () => modalData.value = cloneDeep(DEFAULT_MODAL_DATA);

    return {
        modalData,
        openModal,
        closeModal,
    }
})
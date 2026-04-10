import {defineStore} from "pinia";
import {ref} from "vue";

const toastStatus = {success: "success", error: "error", info: "info"};

const defaultTimeout = 3000;

const createToast = (text, second_text, status, timeout) => ({
    timeout,
    text,
    second_text,
    status,
    id: Math.random() * 1000,
});

export const useToastStore = defineStore('toast', () => {
    const toasts = ref([]);

    const updateState = (payload, status) => {
        const { text, second_text, timeout } = payload;

        const toast = createToast(text, second_text, status, timeout);

        toasts.value.push(toast);

        setTimeout(() => {
            toasts.value = toasts.value.filter((t) => t.id !== toast.id);
        }, timeout ?? defaultTimeout);
    };

    function success(payload) {
        updateState(payload, toastStatus.success);
    }

    function info(payload) {
        updateState(payload, toastStatus.info);
    }

    function error(payload) {
        updateState(payload, toastStatus.error);
    }

    return {
        toasts,
        success,
        error,
        info,
    }
})

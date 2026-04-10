import {defineStore} from "pinia";
import {ref} from "vue";

export const useHeaderRefreshStore = defineStore("headerRefresh", () => {
    const refreshHandler = ref(null);

    const setRefreshHandler = (handler) => {
        refreshHandler.value = typeof handler === "function" ? handler : null;
    };

    const clearRefreshHandler = () => {
        refreshHandler.value = null;
    };

    const runRefreshHandler = async () => {
        if (typeof refreshHandler.value !== "function") {
            return;
        }

        await Promise.resolve(refreshHandler.value());
    };

    return {
        refreshHandler,
        setRefreshHandler,
        clearRefreshHandler,
        runRefreshHandler,
    };
});

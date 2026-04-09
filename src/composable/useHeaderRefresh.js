import {onBeforeUnmount, onMounted} from "vue";
import {useHeaderRefreshStore} from "../store/headerRefreshStore.js";

export const useHeaderRefresh = (handler) => {
    const headerRefreshStore = useHeaderRefreshStore();

    onMounted(() => {
        headerRefreshStore.setRefreshHandler(handler);
    });

    onBeforeUnmount(() => {
        if (headerRefreshStore.refreshHandler === handler) {
            headerRefreshStore.clearRefreshHandler();
        }
    });
};

import {defineStore} from "pinia";
import {ref} from "vue";

export const useSidebarStore = defineStore('sidebar', () => {
    const isMobileMenuOpen = ref(false);

    const toggleMobileMenu = () => isMobileMenuOpen.value = !isMobileMenuOpen.value;

    return {
        isMobileMenuOpen,
        toggleMobileMenu,
    }
})
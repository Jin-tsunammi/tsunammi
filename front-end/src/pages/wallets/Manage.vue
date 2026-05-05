<template>
 <div class="manage">
   <div class="manage__content">
     <DesktopManage
       :columns="columns"
       :rows="projectsStore.selectedProject?.wallets || []"
       :is-page-loading="isPageLoading"
       @get-wallet-private-key="getProjectWalletPrivetKey"
       @open-wallet-modal="openWalletModal"
     />
     <MobileAdaptsNotification class="manage__mobile" />
     <Modals>
       <ModalImportWallets v-if="modalsStore.modalData.type === 'wallet-category-import-wallets'" />
       <ModalCreateEditProject v-if="modalsStore.modalData.type === 'edit-project'" />
       <ConfirmationModal
         class="create-confirmation"
         v-if="modalsStore.modalData.type === 'delete-project' || modalsStore.modalData.type === 'delete-wallet'"
         :main-text="modalsStore.modalData.mainText"
         :additional-text="modalsStore.modalData.additionalText"
         :confirmation-btn-style="'destructive'"
         :confirmation-btn-text="isDeleting ? 'Deleting...' : 'Delete'"
         @handle-confirmation="handleDelete"
         :is-loading="isDeleting"
       />
       <template #custom-content>
         <ModalCreateWallets v-if="modalsStore.modalData.type === 'wallet-category-create-wallets'" />
         <ModalGetPrivateKey v-if="modalsStore.modalData.type === 'wallet-category-get-private-key'" v-model="walletPrivetKey" />
       </template>
     </Modals>
   </div>
 </div>
</template>
<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue';
import {useModalsStore} from "../../store/modalsStore.js";
import ModalCreateWallets from "../../components/Wallets/Modals/ModalCreateWallets.vue";
import ModalImportWallets from "../../components/Wallets/Modals/ModalImportWallets.vue";
import Modals from "../../components/UI/Modals.vue";
import DesktopManage from "../../components/Wallets/Manage/DesktopManage.vue";
import {useProjectsStore} from "../../store/projectsStore.js";
import {DeleteSolWallet, GetWalletPrivetKeyByID} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import ModalGetPrivateKey from "../../components/Wallets/Modals/ModalGetPrivateKey.vue";
import {useRoute, useRouter} from "vue-router";
import ModalCreateEditProject from "../../components/Wallets/Modals/ModalCreateEditProject.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useUserStore} from "../../store/userStore.js";
import {errorToast, formatWalletAddress} from "../../helpers/index.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const route = useRoute();
const router = useRouter();
const modalsStore = useModalsStore();
const projectsStore = useProjectsStore();
const toastStore = useToastStore();
const userStore = useUserStore();
const currentPage = ref(1);
const totalOrders = ref(300);
const projects = ref([]);
const walletPrivetKey = ref('');
const isDeleting = ref(false);
const isPageLoading = ref(false);
const itemsOnPage = 20;
const columns = [
  { label: 'Wallet', field: 'wallets' },
  { label: 'Age', field: 'lifetime' },
  { label: 'Balance', field: 'balance' },
  { label: 'Frozen Money', field: 'frozen_money' },
  { label: '', field: 'actions' },
];

const totalPages = computed(() => {
  return Math.ceil(totalOrders.value / itemsOnPage);
})

const openWalletModal = ({type, item=null}) => {
  if (type === 'create') {
    modalsStore.modalData.title = 'Create wallets';
    modalsStore.modalData.type = 'wallet-category-create-wallets';
    modalsStore.modalData.is_custom = true;
  } else if (type === 'import') {
    modalsStore.modalData.title = 'Import wallets';
    modalsStore.modalData.type = 'wallet-category-import-wallets';
  } else if (type === 'private-key') {
    modalsStore.modalData.title = 'Private key';
    modalsStore.modalData.type = 'wallet-category-get-private-key';
    modalsStore.modalData.is_custom = true;
  } else if (type === 'edit-project') {
    modalsStore.modalData.title = 'Edit project';
    modalsStore.modalData.type = type;
    modalsStore.modalData.item = projectsStore.selectedProject;
  } else if (type === 'delete' || type === 'delete-wallet') {
    if (type === 'delete') {
      modalsStore.modalData.type = 'delete-project';
      modalsStore.modalData.title = `Delete ${projectsStore.selectedProject?.name || ''}?`;
      modalsStore.modalData.mainText = `This action will permanently delete the pool and all associated wallets.`;
      modalsStore.modalData.additionalText = `This action cannot be undone.`;
    } else if (type === 'delete-wallet') {
      modalsStore.modalData.type = 'delete-wallet';
      modalsStore.modalData.title = `Delete ${formatWalletAddress(item?.public_key) || ''}?`;
      modalsStore.modalData.mainText = `This action will permanently delete the wallet.`;
    }
    modalsStore.modalData.action = 'confirmation';
  }

  if (item) {
    modalsStore.modalData.item = item;
  }

  modalsStore.openModal();
}

const handlePageChange = async (page) => {
  currentPage.value = page;

  // await getUsers();
}

const getProjectWalletPrivetKey = async(wallet_id) => {
  if (!wallet_id) return;

  try {
    const resp = await GetWalletPrivetKeyByID(wallet_id);

    walletPrivetKey.value = resp?.data?.private_key || '';

    openWalletModal({type: 'private-key'})
  } catch (error) {
    console.error(error)
    toastStore.error({text: 'Something went wrong'});
    throw error;
  }
}

const handleProjectDelete = async() => {
  isDeleting.value = true;
  await projectsStore.handleProjectDelete(projectsStore.selectedProject);
  isDeleting.value = false;

  await router.push({name: 'WalletsProjects'});
}

const handleWalletDelete = async() => {
  const wallet = modalsStore.modalData.item;

  if (!wallet) {
    modalsStore.closeModal();
    toastStore.error({text: 'Failed to delete wallet.'});
    return;
  }

  try {
    isDeleting.value = true;
    await DeleteSolWallet(wallet.id);
    await getProjectData();
    toastStore.success({text: `Wallet ${formatWalletAddress(wallet.public_key)} has been deleted.`});
  } catch (e) {
    errorToast(e?.response?.data || '');
  } finally {
    modalsStore.closeModal();
    isDeleting.value = false;
  }
}

const handleDelete = async() => {
  if (modalsStore.modalData.type === 'delete-project') {
    await handleProjectDelete();
  } else if (modalsStore.modalData.type === 'delete-wallet') {
    await handleWalletDelete();
  }
}
const getProjectData = async(isRefresh=false) => {
  try {
    isPageLoading.value = true;

    if (isRefresh) {
      await projectsStore.getProjectById(route.params.project_id);
      toastStore.success({text: "Project data refreshed"})
    } else {
      await projectsStore.getProjectById(route.params.project_id, true);
    }

  } catch (e) {

  } finally {
    isPageLoading.value = false;
  }
}

watch(() => projectsStore.allProjects, (newVal) => {
  projects.value = newVal;
}, {immediate: true, deep: true})
useHeaderRefresh(() => getProjectData(true));
onMounted(async() => {
  if (route.params.project_id && userStore.isUserAuth) {
    await getProjectData();
  } else {
    router.back();
  }
})

onBeforeUnmount(() => {
  projectsStore.setSelectedProject(null);
})
</script>
<style scoped lang="scss">
.manage {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: fit-content;
  }

  &__pagination {
    margin-inline: auto;
    margin-top: auto;
  }

  &__mobile {
    display: none;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.create-confirmation {
  max-width: 360px;
}

@media (max-width: 1200px) {
  .manage {
    &__mobile {
      margin: 16px auto;
      padding: 0 16px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>
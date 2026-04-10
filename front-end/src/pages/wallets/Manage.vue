<template>
 <div class="manage">
   <div class="manage__content">
     <PageLoading v-if="isPageLoading" />
     <template v-else>
       <DesktopManage
         :columns="columns"
         :rows="projectsStore.selectedProject?.wallets || []"
         @get-wallet-private-key="getProjectWalletPrivetKey"
         @open-wallet-modal="openWalletModal"
       />
       <MobileAdaptsNotification class="manage__mobile" />
       <!--        <div class="manage__pagination">-->
       <!--          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>-->
       <!--        </div>-->
     </template>
     <Modals>
       <ModalImportWallets v-if="modalsStore.modalData.type === 'wallet-category-import-wallets'" />
       <ModalCreateEditProject v-if="modalsStore.modalData.type === 'edit-project'" />
       <ConfirmationModal
         class="create-confirmation"
         v-if="modalsStore.modalData.type === 'delete-project'"
         main-text="This action will permanently delete the project and all associated wallets."
         additional-text="This action cannot be undone."
         :confirmation-btn-style="'destructive'"
         :confirmation-btn-text="isDeleting ? 'Deleting...' : 'Delete'"
         @handle-confirmation="handleProjectDelete"
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
import UIButton from "../../components/UI/UIButton.vue";
import SVGImport from "../../components/SVG/SVGImport.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import ModalCreateWallets from "../../components/Wallets/Modals/ModalCreateWallets.vue";
import ModalImportWallets from "../../components/Wallets/Modals/ModalImportWallets.vue";
import Modals from "../../components/UI/Modals.vue";
import Pagination from "../../components/UI/Pagination.vue";
import DesktopManage from "../../components/Wallets/Manage/DesktopManage.vue";
import MobileManage from "../../components/Wallets/Manage/MobileManage.vue";
import {useProjectsStore} from "../../store/projectsStore.js";
import {GetWalletPrivetKeyByID} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import ModalGetPrivateKey from "../../components/Wallets/Modals/ModalGetPrivateKey.vue";
import {useRoute, useRouter} from "vue-router";
import ModalCreateEditProject from "../../components/Wallets/Modals/ModalCreateEditProject.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import PageLoading from "../../components/UI/PageLoading.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useUserStore} from "../../store/userStore.js";

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
  } else if (type === 'delete') {
    modalsStore.modalData.type = 'delete-project';
    modalsStore.modalData.title = `Delete ${projectsStore.selectedProject?.name || ''}?`;
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

const getProjectData = async(isRefresh=false) => {
  try {
    isPageLoading.value = true;

    if (isRefresh) {
      await projectsStore.getProjectById(route.params.project_id);
      toastStore.success({text: "Project data refreshed"})
    } else {
      await projectsStore.getProjectById(route.params.project_id, true);
    }

    isPageLoading.value = false;
  } catch (e) {

  }
}

watch(() => projectsStore.allProjects, (newVal) => {
  projects.value = newVal;
}, {immediate: true, deep: true})

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
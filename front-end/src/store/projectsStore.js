import {defineStore} from "pinia";
import {ref, watch} from "vue";
import {
    CreateNewProject,
    DeleteProject,
    GetAllProjects, GetCachedProjectByID,
    GetProjectByID, UpdateProject,
} from "../api/api.js";
import {useToastStore} from "./toastStore.js";
import {useModalsStore} from "./modalsStore.js";
import {useRoute, useRouter} from "vue-router";
import {errorToast, fetchSolTokenMetadata, formatText} from "../helpers/index.js";
import {useTokensStore} from "./tokensStore.js";

export const useProjectsStore = defineStore('projects', () => {
    const tokensStore = useTokensStore();
    const toastStore = useToastStore();
    const modalStore = useModalsStore();
    const route = useRoute();
    const router = useRouter();

    const selectedProject = ref(null);
    const allProjects = ref([]);
    const solTokensData = ref({});

    const setAllProjects = (data) => allProjects.value = data;
    const setSelectedProject = (data) => selectedProject.value = data;

    const getAllProjects = async (params = null) => {
        try {
            const resp = await GetAllProjects(params);

            allProjects.value = resp.data.projects;

            return resp.data;
        } catch (error) {
            console.error(error)
            errorToast(error.response.data)
            throw error;
        }
    }

    const updateProjectName = async(id, newName) => {
        if (!id || !newName.trim().length) return;
        try {
            await UpdateProject(id, {name: newName.trim()});
            selectedProject.value.name = newName;
            toastStore.success({text: 'Project data updated.'});
        } catch(error) {
           errorToast(error.response.data);
        }
    }

    const updateProjectData = async (project_id, isRefreshing=true) => {
        if (!project_id) return;

        try {
            const resp = await GetProjectByID(project_id);

            if (resp.data) {
                if (route.name === 'WalletsSelectedProject') {
                    selectedProject.value = resp.data;
                } else {
                    const originalIndex = allProjects.value.findIndex(elem => elem.id === resp.data.id);

                    if (originalIndex > -1) {
                        allProjects.value[originalIndex] = resp.data;
                    }
                }
            }

            if (isRefreshing) {
                toastStore.success({text: 'Data refreshed'});
            }
        } catch (error) {
            console.error(error)
            errorToast(error.response.data);
            throw error;
        }
    }

    const getProjectById = async (project_id, isCached=false) => {
        if (!project_id) return;

        try {
            let resp = null;
            resp = await GetProjectByID(project_id);

            selectedProject.value = resp.data;
        } catch (error) {
            console.error(error)
            errorToast(error.response.data);
            await router.push({name: 'WalletsProjects'});
            throw error;
        }
    }

    const handleProjectDelete = async (project) => {
        if (!project) return;

        try {
            await DeleteProject(project.id);
            await getAllProjects();
            modalStore.closeModal();
            toastStore.success({text: `Project “${project.name}” has been deleted.`});
        } catch (error) {
            console.error(error);
            toastStore.error({text: 'Something went wrong'});
        }
    }

    const handleProjectCreate = async (data) => {
        if (!data) return;
        let result;

        try {
            const resp = await CreateNewProject(data);
            await getAllProjects();
            result = resp.data;

            return result;
        } catch (error) {
            console.error(error);
            errorToast(error.response.data);
        }
    }

    const refreshManageRequests = async (isPressedRefreshBtn = false) => {
        try {
            await getAllProjects();


            if (isPressedRefreshBtn) {
                toastStore.success({text: 'Page data refreshed'});
            }
        } catch (error) {
        }
    }

    watch(
        () => selectedProject.value,
        async(project) => {
            if (!project) return;

            const tokens = [];

            project.wallets?.forEach(wallet => {
                if (wallet.tokens?.length) {
                    wallet.tokens.forEach((token) => {
                        tokens.push(token);
                    })
                }
            })

            await tokensStore.updateSolTokensData(tokens, 'token_symbol');
        },
        {deep: true, immediate: true}
    );
    return {
        allProjects,
        selectedProject,
        solTokensData,

        setAllProjects,
        setSelectedProject,
        getAllProjects,
        refreshManageRequests,
        updateProjectData,
        handleProjectDelete,
        handleProjectCreate,
        getProjectById,
        updateProjectName,
    }
})
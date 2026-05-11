import {apiInstance} from "./instance.js";

//Auth
export const GetCodeByEmail = async (data) => apiInstance.post('/auth/send-code', data);
export const SignInByEmail = async (data) => apiInstance.post('/auth/sign-in-email', data);
export const SignInByGoogle = async (id_token) => apiInstance.post('/auth/sign-in-google', {}, {
    headers: {
        Authorization: `${id_token}`,
    },
});
export const SignUpByEmail = async (data) => apiInstance.post('/auth/sign-up-email', data);
export const SignUpByGoogle = async (id_token) => {
    return apiInstance.post(
        "/auth/sign-up-google",
        {},
        {
            headers: {
                Authorization: `${id_token}`,
            },
        }
    );
};
export const SignUpByWallet = async (data) => apiInstance.post('/auth/sign-in-wallet', data);
export const CheckEmail = async (data) => apiInstance.post('/auth/is-user-exists', data);

//User
export const GetUser = async () => apiInstance.get('/user');
export const GetUserHistory = async (params) => apiInstance.get('/user/history', {
    params: {
        from: params?.from,
        to: params?.to,
        page: params?.page,
        pageSize: params?.pageSize,
    }
});
export const ChangeUserEmail = async (data) => apiInstance.patch('/user/email', data);

//Projects
export const GetAllProjects = async (params) => apiInstance.get('/projects', {
    params: {
        page: params?.page,
        pageSize: params?.pageSize,
        sortBy: params?.sortBy,
        sortOrder: params?.sortOrder,
    }
});
export const GetAllProjectsNameOnly = async () => apiInstance.get('/without-wallets/projects',);
export const GetAllProjectsWithBalance = async (params) => apiInstance.get('/mint-balance/projects', {
    params: {
        mint: params?.mint,
    }
});
export const GetProjectWithBalance = async (id, params) => apiInstance.get(`/mint-balance/projects/${id}`, {
    params: {
        mint: params?.mint,
    }
});
export const GetProjectByID = async (id) => apiInstance.get(`/projects/${id}`);
export const GetCachedProjectByID = async (id) => apiInstance.get(`cache/projects/${id}`);
export const CreateNewProject = async (data) => apiInstance.post('/projects', data);
export const DeleteProject = async (id) => apiInstance.delete(`/projects/${id}`);
export const UpdateProject = async (id, data) => apiInstance.put(`/projects/${id}`, data);

//Exchanges
export const GetExchanges = async () => apiInstance.get('/exchanges');

//Wallets
export const GenerateSolWallets = async (data) => apiInstance.post('/wallets/solana/generate', data);
export const ImportSolWallets = async (data) => apiInstance.post('/wallets/solana/import', data);
export const ImportSolWalletsFromFile = async (data) => apiInstance.post('/wallets/solana/import-file', data);
export const MonitorSolWallets = async (data) => apiInstance.post('/wallets/solana/monitor', data);
export const GetWalletPrivetKeyByID = async (id) => apiInstance.get(`/wallets/solana/${id}`);
export const DeleteSolWallet = async (id) => apiInstance.delete(`/wallets/solana/${id}`);
export const GetWalletsPrivateKeys = async (id) => apiInstance.get(`/wallets/solana/`, {
    params: {
        projectID: id,
    }
});

//Accounts
export const GetAllCEXApi = async () => apiInstance.get(`/accounts`);
export const CreateNewCEXApi = async (data) => apiInstance.post(`/accounts`, data);
export const GetCEXApiByID = async (id) => apiInstance.get(`/accounts/${id}`);
export const DeleteCEXApi = async (id) => apiInstance.delete(`/accounts/${id}`);

//Jito
export const GetJitoInfo = async () => apiInstance.get(`/jito/tip-floor`);

//Deposit
export const CreateDeposit = async (data) => apiInstance.post(`/wallets/solana/deposit`, data);
export const ProcessDepositOrder = async (id) => apiInstance.post(`/wallets/solana/deposit/process/${id}`);
export const GetDepositHistory = async () => apiInstance.get(`/wallets/solana/deposit/history`);
export const GetDepositHistoryByProjectID = async (id) => apiInstance.get(`/wallets/solana/deposit/history/${id}`);

//PumpFun
export const GetPumpFunEstimate = async (data) => apiInstance.post('/pumpfun/estimate', data);
export const CreatePumpFunPullDown = async (data) => apiInstance.post('/pumpfun/pull-down', data);
export const CreatePumpFunPullUp = async (data) => apiInstance.post('/pumpfun/pull-up', data);

//Raydium
export const GetRaydiumEstimate = async (data) => apiInstance.post('/raydium/estimate', data);
export const CreateRaydiumPullDown = async (data) => apiInstance.post('/raydium/pull-down', data);
export const CreateRaydiumPullUp = async (data) => apiInstance.post('/raydium/pull-up', data);

//Server IP
export const GetServerIP = async () => apiInstance.get('/ip');

//Campaign
export const GetAllCampaigns = async (params) => apiInstance.get('/campaigns', {
    params: {
        page: params?.page,
        pageSize: params?.pageSize,
        type: params?.type,
    }
});
export const GetAllActiveCampaigns = async (params) => apiInstance.get('/campaigns-summary', {
    params: {
        status: params?.status,
        type: params?.type,
    }
});
export const GetCampaignAllTransactions = async (id, params=null) => apiInstance.get(`/campaigns/${id}/transactions`, {
    params: {
        page: params?.page,
        pageSize: params?.pageSize,
    }
});
export const UpdateCampaign = async (id, data) => apiInstance.patch(`/campaigns/${id}`, data);
export const DeleteCampaign = async (id) => apiInstance.delete(`/campaigns/${id}`);
export const GetCampaignByID = async (id) => apiInstance.get(`/campaigns/${id}`);

//SmartBuyBack
export const CreateSmartBuyBack = async(data) => apiInstance.post('/buyback', data);
export const StopSmartBuyBack = async(id) => apiInstance.delete(`/buyback/${id}`);
export const GetSmartBuyBack = async(id) => apiInstance.get(`/buyback/${id}`);
export const GetSmartBuyBackTransactions = async(id, params) => apiInstance.get(`/buyback/${id}/transactions`, {
    params: {
        page: params?.page,
        pageSize: params?.pageSize,
    }
});
export const GetSmartBuyBackHistory = async(params) => apiInstance.get(`/buyback`, {
    params: {
        page: params?.page,
        pageSize: params?.pageSize,
        status: params?.status,
    }
});
export const StopSmartBuyBackTarget = async({id, targetID}) => apiInstance.delete(`/buyback/${id}/targets/${targetID}`);
export const CreateSmartBuyBackTarget = async(id, data) => apiInstance.post(`/buyback/${id}/targets`, data);
export const UpdateSmartBuyBackTarget = async({id, targetID, data}) => apiInstance.patch(`/buyback/${id}/targets/${targetID}`, data);

//Upload via Lighthouse
export const UploadImage = async(formdata) => apiInstance.post(`/upload/image`, formdata);
export const UploadMetadata = async(data) => apiInstance.post(`/upload/metadata`, data);

//Launch Token on PumpFun
export const GetTokenPumpFunEstimate = async(data) => apiInstance.post(`/launch/pumpfun/estimate`, data);
export const PostTokenPumpFunPrepare = async(formdata) => apiInstance.post(`/launch/pumpfun/prepare`, formdata);
export const LaunchTokenPumpFunPrepare = async(data) => apiInstance.post(`/launch/pumpfun/launch`, data);

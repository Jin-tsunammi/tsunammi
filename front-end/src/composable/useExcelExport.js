import * as XLSX from "xlsx";
import { sanitizeFilename, getToday } from "../helpers/index.js";

export function useExcelExport() {
    const pickFirst = (obj, keys) => {
        if (!obj) return "";
        for (const k of keys) {
            const v = obj?.[k];
            if (v !== undefined && v !== null && v !== "") return v;
        }
        return "";
    };

    const buildProjectSheet = (project) => {
        const header = [
            "Project name",
            "ID",
            "Wallets count",
            "Balance (SOL)",
            "Balance (USD)",
        ];

        const data = [
            pickFirst(project, ["name", "title"]),
            pickFirst(project, ["id", "project_id", "projectId"]),
            pickFirst(project, ["wallet_count", "wallets_qty", "walletCount"]),
            pickFirst(project, ["total_balance_sol", "totalBalanceSol"]),
            pickFirst(project, ["total_balance_usd", "totalBalanceUsd"]),
        ];

        const ws = XLSX.utils.aoa_to_sheet([header, data]);
        ws["!cols"] = [
            { wch: 20 },
            { wch: 10 },
            { wch: 18 },
            { wch: 16 },
            { wch: 16 },
        ];

        return ws;
    };

    const buildWalletsSheet = (wallets) => {
        const header = [
            "Wallet ID",
            "Public key",
            "Balance (SOL)",
            "Balance (USD)",
        ];

        const data = (wallets ?? []).map((w) => [
            pickFirst(w, ["id", "wallet_id", "walletId"]),
            pickFirst(w, ["public_key", "publicKey", "address"]),
            pickFirst(w, ["balance_sol", "balanceSol"]),
            pickFirst(w, ["balance_usd", "balanceUsd"]),
        ]);

        const ws = XLSX.utils.aoa_to_sheet([header, ...data]);
        ws["!cols"] = [
            { wch: 9 },
            { wch: 52 },
            { wch: 16 },
            { wch: 16 },
        ];

        return ws;
    };

    const exportProjectExcel = ({ project, wallets }) => {
        if (!project) return;

        const finalWallets = (wallets ?? project.wallets ?? []).filter(Boolean);

        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, buildProjectSheet(project), "Project");
        XLSX.utils.book_append_sheet(wb, buildWalletsSheet(finalWallets), "Wallets");

        const projectName =
            pickFirst(project, ["name", "title"]) ||
            pickFirst(project, ["id", "project_id", "projectId"]) ||
            "project";

        const filename = sanitizeFilename(
            `${projectName}_${getToday()}.xlsx`
        );

        XLSX.writeFile(wb, filename);
    };

    return {
        exportProjectExcel,
    };
}

import * as XLSX from "xlsx";
import { sanitizeFilename, getToday } from "../helpers/index.js";

export function usePrivateKeysExport() {
    const exportPrivateKeysExcel = (keys, projectName = "private-keys") => {
        const list = Array.isArray(keys) ? keys : [];

        const header = ["Public key", "Private key"];
        const rows = list.map((item) => [
            item.public_key ?? "",
            item.private_key ?? "",
        ]);

        const ws = XLSX.utils.aoa_to_sheet([header, ...rows]);
        ws["!cols"] = [
            { wch: 52 },
            { wch: 100 },
        ];

        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, `${projectName} Private keys`);

        const filename = sanitizeFilename(
            `${projectName}_private-keys_${getToday()}.xlsx`
        );

        XLSX.writeFile(wb, filename);
    };

    return {
        exportPrivateKeysExcel,
    };
}

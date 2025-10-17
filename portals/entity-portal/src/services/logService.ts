interface LogEntry {
    id: string;
    timestamp: string;
    status: 'failure' | 'success';
    requestedData: string;
    consumerId: string;
    providerId: string;
}

interface LogResponse {
    logs: LogEntry[];
    count: number;
}

interface LogQueryParams {
    consumerId?: string;
    providerId?: string;
    status?: 'success' | 'failure';
    startDate?: string;
    endDate?: string;
    limit?: number;
    offset?: number;
}

export class LogService {

    static async fetchLogsWithParams(params?: LogQueryParams): Promise<LogEntry[]> {
        try {
            const url = new URL(`${window.configs.logsUrl}/logs`);
            
            if (params) {
                Object.entries(params).forEach(([key, value]) => {
                    if (value !== undefined && value !== null && value !== '') {
                        url.searchParams.append(key, value.toString());
                    }
                });
            }
            
            const response = await fetch(url.toString());
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data:LogResponse = await response.json();
            
            return data.logs;
            
        } catch (error) {
            console.error('Error fetching logs with params:', error);
            throw error;
        }
    }

    
    /**
     * Export logs data (placeholder for future implementation)
     * @param params - Query parameters for filtering the export
     * @param format - Export format (csv, json, etc.)
     * @returns Promise<Blob>
     */
    static async exportLogs(params?: LogQueryParams, format: 'csv' | 'json' = 'csv'): Promise<Blob> {
        try {
            const url = new URL(`${window.configs.logsUrl}/logs/export`);
            url.searchParams.append('format', format);
            
            if (params) {
                Object.entries(params).forEach(([key, value]) => {
                    if (value !== undefined && value !== null && value !== '') {
                        url.searchParams.append(key, value.toString());
                    }
                });
            }
            
            const response = await fetch(url.toString());
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.blob();
        } catch (error) {
            console.error('Error exporting logs:', error);
            throw error;
        }
    }

    /**
     * Get log statistics
     * @param params - Query parameters for filtering
     * @returns Promise with log statistics
     */
    static async getLogStatistics(params?: LogQueryParams): Promise<{
        total: number;
        success: number;
        failure: number;
        successRate: number;
    }> {
        try {
            const logs = await this.fetchLogsWithParams(params);
            const total = logs.length;
            const success = logs.filter(log => log.status === 'success').length;
            const failure = logs.filter(log => log.status === 'failure').length;
            const successRate = total > 0 ? (success / total) * 100 : 0;

            return {
                total,
                success,
                failure,
                successRate
            };
        } catch (error) {
            console.error('Error getting log statistics:', error);
            throw error;
        }
    }
}

export type { LogEntry, LogResponse, LogQueryParams };
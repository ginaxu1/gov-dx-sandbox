import { URLBuilder } from '../utils';

interface LogEntry {
    id: string;
    timestamp: string;
    status: 'failure' | 'success';
    requestedData: string;
    applicationId: string;
    consumerId: string;
    schemaId: string;
    providerId: string;
}

interface LogResponse {
    logs: LogEntry[];
    total: number;
}

interface LogQueryParams {
    consumerId?: string;
    providerId?: string;
    startDate?: string;
    endDate?: string;
    limit?: number;
    offset?: number;
}

export class LogService {
    static async fetchLogsWithParams(params?: LogQueryParams): Promise<LogEntry[]> {
        try {
            const url = URLBuilder.build(
                window.configs.VITE_LOGS_URL, 
                '/logs', 
                params as Record<string, string | number | boolean | undefined | null>
            );
            
            const response = await fetch(url);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: LogResponse = await response.json();
            if (!data || !Array.isArray(data.logs)) {
                throw new Error('Invalid log data received from API');
            }
            if (data.total === 0) {
                return [];
            }
            return data.logs;
        } catch (error) {
            console.error('Error fetching logs with params:', error);
            throw error;
        }
    }

    static async exportLogs(params?: LogQueryParams, format: 'csv' | 'json' = 'csv'): Promise<Blob> {
        try {
            const urlBuilder = URLBuilder.from(window.configs.VITE_LOGS_URL)
                .path('/logs/export')
                .param('format', format);
            
            if (params) {
                urlBuilder.params(params as Record<string, string | number | boolean | undefined | null>);
            }
            
            const response = await fetch(urlBuilder.build());
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.blob();
        } catch (error) {
            console.error('Error exporting logs:', error);
            throw error;
        }
    }

    static async getLogStatistics(params?: LogQueryParams): Promise<{
        total: number;
        success: number;
        failure: number;
        successRate: number;
    }> {
        try {
            const logs = await LogService.fetchLogsWithParams(params);
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
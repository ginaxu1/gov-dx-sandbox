interface LogEntry {
    id: string;
    timestamp: string;
    status: 'failure' | 'success';
    requestedData: string;
    consumerId: string;
    providerId: string;
}

interface LogResponse {
    logs?: LogEntry[];
    items?: LogEntry[];
    count?: number;
}

interface LogQueryParams {
    consumerId?: string;
    providerId?: string;
    status?: 'success' | 'failure';
    startDate?: string;
    endDate?: string;
    search?: string;
    limit?: number;
    offset?: number;
}

class LogService {
    private baseUrl: string;

    constructor() {
        // Default to localhost for now, can be configured based on environment
        this.baseUrl = window.configs.logsUrl;
    }

    /**
     * Set the base URL for API calls
     * @param url - The base URL to set
     */
    setBaseUrl(url: string): void {
        this.baseUrl = url;
    }

    /**
     * Fetch all logs from the API
     * @returns Promise<LogEntry[]>
     */
    async fetchLogs(): Promise<LogEntry[]> {
        try {
            const response = await fetch(`${this.baseUrl}/logs`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            // Handle different response formats
            if (Array.isArray(data)) {
                return data;
            } else if (data.logs && Array.isArray(data.logs)) {
                return data.logs;
            } else if (data.items && Array.isArray(data.items)) {
                return data.items;
            } else {
                throw new Error('Invalid response format');
            }
        } catch (error) {
            console.error('Error fetching logs:', error);
            throw error;
        }
    }

    /**
     * Fetch logs with query parameters for role-based filtering
     * @param role - The role of the user ('consumer' | 'provider')
     * @param entityId - The consumer or provider ID
     * @returns Promise<LogEntry[]>
     */
    async fetchLogsByRole(role: 'consumer' | 'provider', entityId: string): Promise<LogEntry[]> {
        try {
            const params = role === 'consumer' 
                ? { consumerId: entityId }
                : { providerId: entityId };
            
            return await this.fetchLogsWithParams(params);
        } catch (error) {
            console.error(`Error fetching logs for ${role} ${entityId}:`, error);
            throw error;
        }
    }

    /**
     * Fetch logs with optional query parameters
     * @param params - Query parameters for filtering
     * @returns Promise<LogEntry[]>
     */
    async fetchLogsWithParams(params?: LogQueryParams): Promise<LogEntry[]> {
        try {
            const url = new URL(`${this.baseUrl}/logs`);
            
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
            
            const data = await response.json();
            
            // Handle different response formats
            if (Array.isArray(data)) {
                return data;
            } else if (data.logs && Array.isArray(data.logs)) {
                return data.logs;
            } else if (data.items && Array.isArray(data.items)) {
                return data.items;
            } else {
                throw new Error('Invalid response format');
            }
        } catch (error) {
            console.error('Error fetching logs with params:', error);
            throw error;
        }
    }

    /**
     * Fetch a single log entry by ID
     * @param logId - The log ID to fetch
     * @returns Promise<LogEntry>
     */
    async fetchLogById(logId: string): Promise<LogEntry> {
        try {
            const response = await fetch(`${this.baseUrl}/logs/${logId}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: LogEntry = await response.json();
            return data;
        } catch (error) {
            console.error(`Error fetching log ${logId}:`, error);
            throw error;
        }
    }

    /**
     * Export logs data (placeholder for future implementation)
     * @param params - Query parameters for filtering the export
     * @param format - Export format (csv, json, etc.)
     * @returns Promise<Blob>
     */
    async exportLogs(params?: LogQueryParams, format: 'csv' | 'json' = 'csv'): Promise<Blob> {
        try {
            const url = new URL(`${this.baseUrl}/logs/export`);
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
    async getLogStatistics(params?: LogQueryParams): Promise<{
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

// Create and export a singleton instance
export const logService = new LogService();
export type { LogEntry, LogResponse, LogQueryParams };
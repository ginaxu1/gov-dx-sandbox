interface Entity {
    entityId: string;
    name: string;
    entityType: 'gov' | 'admin' | 'private';
    email: string;
    phoneNumber: string;
    consumerId?: string;
    providerId?: string;
    createdAt: string;
    updatedAt: string;
}

interface EntityResponse {
    items: Entity[];
    count: number;
}

export class MemberService {
    /**
     * Fetch all entities from the API
     * @returns Promise<EntityResponse>
     */
    static async fetchEntities(): Promise<Entity[]> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/entities`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: EntityResponse = await response.json();
            return data.items;
        } catch (error) {
            console.error('Error fetching entities:', error);
            throw error;
        }
    }

    /**
     * Fetch entities with optional query parameters
     * @param params - Query parameters for filtering
     * @returns Promise<EntityResponse>
     */
    static async fetchEntitiesWithParams(params?: Record<string, string>): Promise<EntityResponse> {
        try {
            const url = new URL(`${window.configs.apiUrl}/entities`);
            
            if (params) {
                Object.entries(params).forEach(([key, value]) => {
                    if (value) {
                        url.searchParams.append(key, value);
                    }
                });
            }
            
            const response = await fetch(url.toString());
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: EntityResponse = await response.json();
            return data;
        } catch (error) {
            console.error('Error fetching entities with params:', error);
            throw error;
        }
    }

    /**
     * Fetch a single entity by ID
     * @param entityId - The entity ID to fetch
     * @returns Promise<Entity>
     */
    static async fetchEntityById(entityId: string): Promise<Entity> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/entities/${entityId}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: Entity = await response.json();
            return data;
        } catch (error) {
            console.error(`Error fetching entity ${entityId}:`, error);
            throw error;
        }
    }

    /**
     * Export entities data (placeholder for future implementation)
     * @param format - Export format (csv, json, etc.)
     * @returns Promise<Blob>
     */
    static async exportEntities(format: 'csv' | 'json' = 'csv'): Promise<Blob> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/entities/export?format=${format}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.blob();
        } catch (error) {
            console.error('Error exporting entities:', error);
            throw error;
        }
    }
}

export type { Entity, EntityResponse };
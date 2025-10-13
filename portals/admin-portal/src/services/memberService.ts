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

class MemberService {
    private baseUrl: string;

    constructor() {
        // Get the API URL from window.configs or fallback to localhost
        this.baseUrl = (window as any).configs?.apiUrl || 'https://localhost:8080/api/v1';
    }

    /**
     * Fetch all entities from the API
     * @returns Promise<EntityResponse>
     */
    async fetchEntities(): Promise<EntityResponse> {
        try {
            const response = await fetch(`${this.baseUrl}/entities`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: EntityResponse = await response.json();
            return data;
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
    async fetchEntitiesWithParams(params?: Record<string, string>): Promise<EntityResponse> {
        try {
            const url = new URL(`${this.baseUrl}/entities`);
            
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
    async fetchEntityById(entityId: string): Promise<Entity> {
        try {
            const response = await fetch(`${this.baseUrl}/entities/${entityId}`);
            
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
    async exportEntities(format: 'csv' | 'json' = 'csv'): Promise<Blob> {
        try {
            const response = await fetch(`${this.baseUrl}/entities/export?format=${format}`);
            
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

// Create and export a singleton instance
export const memberService = new MemberService();
export type { Entity, EntityResponse };
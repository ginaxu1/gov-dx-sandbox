interface Entity {
    entityId: string;
    idpUserId: string;
    name: string;
    email: string;
    phoneNumber: string;
    createdAt: string;
    updatedAt: string;
}

interface EntityResponse {
    items: Entity[];
    count: number;
}

export class MemberService {
    static async createEntity(memberData: Partial<Entity>): Promise<Entity> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/entities`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(memberData),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data: Entity = await response.json();
            return data;
        } catch (error) {
            console.error('Error creating entity:', error);
            throw error;
        }
    }

    static async updateEntity(entityId: string, memberData: Partial<Entity>): Promise<Entity> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/entities/${entityId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(memberData),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data: Entity = await response.json();
            return data;
        } catch (error) {
            console.error('Error updating entity:', error);
            throw error;
        }
    }

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

    // static async fetchEntitiesWithParams(params?: Record<string, string>): Promise<EntityResponse> {
    //     try {
    //         const url = new URL(`${window.configs.apiUrl}/entities`);         
    //         if (params) {
    //             Object.entries(params).forEach(([key, value]) => {
    //                 if (value) {
    //                     url.searchParams.append(key, value);
    //                 }
    //             });
    //         }        
    //         const response = await fetch(url.toString());      
    //         if (!response.ok) {
    //             throw new Error(`HTTP error! status: ${response.status}`);
    //         }   
    //         const data: EntityResponse = await response.json();
    //         return data;
    //     } catch (error) {
    //         console.error('Error fetching entities with params:', error);
    //         throw error;
    //     }
    // }

}

export type { Entity, EntityResponse };
interface Member {
    memberId: string;
    idpUserId: string;
    name: string;
    email: string;
    phoneNumber: string;
    createdAt: string;
    updatedAt: string;
}

interface MemberResponse {
    items: Member[];
    count: number;
}

export class MemberService {
    static async createMember(memberData: Partial<Member>): Promise<Member> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/members`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(memberData),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data: Member = await response.json();
            return data;
        } catch (error) {
            console.error('Error creating member:', error);
            throw error;
        }
    }

    static async updateMember(memberId: string, memberData: Partial<Member>): Promise<Member> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/members/${memberId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(memberData),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data: Member = await response.json();
            return data;
        } catch (error) {
            console.error('Error updating member:', error);
            throw error;
        }
    }

    static async fetchMembers(): Promise<Member[]> {
        try {
            const response = await fetch(`${window.configs.apiUrl}/members`);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data: MemberResponse = await response.json();
            return data.items;
        } catch (error) {
            console.error('Error fetching members:', error);
            throw error;
        }
    }

    // static async fetchMembersWithParams(params?: Record<string, string>): Promise<MemberResponse> {
    //     try {
    //         const url = new URL(`${window.configs.apiUrl}/members`);
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
    //         const data: MemberResponse = await response.json();
    //         return data;
    //     } catch (error) {
    //         console.error('Error fetching members with params:', error);
    //         throw error;
    //     }
    // }

}

export type { Member, MemberResponse };
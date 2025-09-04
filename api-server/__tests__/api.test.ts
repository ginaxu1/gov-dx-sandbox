import request from 'supertest';
import app from '../src/index';
// Corrected import path for the database file
import { applicationsDB, providerProfilesDB, providerSchemasDB } from '../src/database';

// Helper to reset the in-memory databases before each test
const resetDatabases = () => {
    applicationsDB.length = 0;
    providerProfilesDB.length = 0;
    providerSchemasDB.length = 0;
};

describe('API Endpoints', () => {
    beforeEach(() => {
        resetDatabases();
    });

    // Tests for Consumer Application Endpoints
    describe('Consumer Applications [/applications]', () => {
        const validApplicationPayload = {
            appId: "passport-app-123",
            requiredFields: { "drp.personInfo.address": {} }
        };

        it('POST /applications - should create a new application', async () => {
            const res = await request(app).post('/applications').send(validApplicationPayload);
            expect(res.statusCode).toEqual(201);
            expect(res.body.status).toBe('pending');
        });

        it('POST /applications - should fail with 409 if application already exists', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app).post('/applications').send(validApplicationPayload);
            expect(res.statusCode).toEqual(409);
        });
    });

    // Tests for Provider Profile Endpoints
    describe('Provider Profiles [/providers]', () => {
        const validProviderPayload = {
            providerName: "Department of Registration of Persons",
            contactEmail: "contact@drp.gov",
            phoneNumber: "123456789",
            providerType: "government" as const
        };

        it('POST /providers - should create a new provider profile', async () => {
            const res = await request(app).post('/providers').send(validProviderPayload);
            expect(res.statusCode).toEqual(201);
            expect(res.body).toHaveProperty('providerId');
        });
    });

    // Tests for Provider Schema Endpoints
    describe('Provider Schemas', () => {
        let testProviderId: string;

        beforeEach(async () => {
            const res = await request(app).post('/providers').send({
                providerName: "Test Provider",
                contactEmail: "test@provider.com",
                phoneNumber: "555-5555",
                providerType: "business"
            });
            testProviderId = res.body.providerId;
        });
        
        // UPDATED: This helper now creates the new nested payload structure
        const createValidSchemaPayload = () => ({
            schemaInput: { type: 'sdl', value: 'type Query { person: PersonData } type PersonData { nic: String, fullName: String }' },
            fieldConfigurations: {
                "PersonData": {
                    "nic": { source: 'fallback', isOwner: false, description: 'National ID Card number.' },
                    "fullName": { source: 'authoritative', isOwner: true, description: 'The full legal name.' }
                },
                "Query": {
                    "person": { source: 'authoritative', isOwner: true, description: 'The main person query.' }
                }
            }
        });

        it('POST /providers/:providerId/register_schema - should create a new schema submission', async () => {
            const res = await request(app)
                .post(`/providers/${testProviderId}/register_schema`)
                .send(createValidSchemaPayload());

            expect(res.statusCode).toEqual(201);
            expect(res.body).toHaveProperty('submissionId');
            expect(res.body.providerId).toBe(testProviderId);
            // Verify the nested structure was received correctly
            expect(res.body.fieldConfigurations.PersonData.nic.isOwner).toBe(false);
        });

        it('POST /providers/:providerId/register_schema - should fail with 400 for invalid payload', async () => {
            const invalidPayload = {
                fieldConfigurations: {
                    "PersonData": {
                        "nic": { isOwner: false, description: 'Missing source' } // 'source' is required
                    }
                }
            };
            const res = await request(app)
                .post(`/providers/${testProviderId}/register_schema`)
                .send(invalidPayload);
            expect(res.statusCode).toEqual(400);
        });

        it('GET /providers/:providerId/schema - should return a specific schema', async () => {
            await request(app).post(`/providers/${testProviderId}/register_schema`).send(createValidSchemaPayload());
            const res = await request(app).get(`/providers/${testProviderId}/schema`);
            expect(res.statusCode).toEqual(200);
            expect(res.body.providerId).toEqual(testProviderId);
        });
        
        it('GET /provider-schemas - should return an array of all schema submissions', async () => {
            await request(app).post(`/providers/${testProviderId}/register_schema`).send(createValidSchemaPayload());
            const res = await request(app).get('/provider-schemas');
            expect(res.statusCode).toEqual(200);
            expect(res.body.length).toBe(1);
        });

        it('POST /provider-schemas/:submissionId/review - should approve a pending schema', async () => {
            const creationRes = await request(app).post(`/providers/${testProviderId}/register_schema`).send(createValidSchemaPayload());
            const submissionId = creationRes.body.submissionId;

            const reviewRes = await request(app)
                .post(`/provider-schemas/${submissionId}/review`)
                .send({ decision: 'approve' });
            
            expect(reviewRes.statusCode).toEqual(200);
            expect(reviewRes.body.status).toEqual('approved');
        });
    });
});


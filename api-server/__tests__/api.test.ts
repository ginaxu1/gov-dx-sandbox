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

describe('API Endpoints with Standardized Responses', () => {

    beforeEach(() => {
        // Reset the state of all in-memory DBs before each test
        resetDatabases();
    });

    // --- Tests for Consumer Application Endpoints ---
    describe('Consumer Applications', () => {
        const validApplicationPayload = {
            appId: "passport-app-123",
            requiredFields: { "drp.PersonData.nic": {}, "rgd.BirthCertificate.birthDate": {} }
        };

        it('GET /available-fields - should return the mock structure of available fields', async () => {
            const res = await request(app).get('/available-fields');
            
            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data).toHaveProperty('drp');
            expect(res.body.data).toHaveProperty('rgd');
            expect(res.body.data.drp.types.PersonData).toBeDefined();
        });

        it('POST /applications - should create a new application', async () => {
            const res = await request(app).post('/applications').send(validApplicationPayload);

            expect(res.statusCode).toEqual(201);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('pending');
            expect(res.body.data.appId).toBe(validApplicationPayload.appId);
        });

        it('POST /applications - should return a structured error if application already exists', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app).post('/applications').send(validApplicationPayload);
            
            expect(res.statusCode).toEqual(409);
            expect(res.body.status).toBe('error');
            expect(res.body.data).toBeNull();
            expect(res.body.error).toBeDefined();
        });
    });

    // --- Tests for Provider Profile Endpoints ---
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
            expect(res.body.status).toBe('success');
            expect(res.body.data).toHaveProperty('providerId');
        });
    });

    // --- Tests for Provider Schema Endpoints ---
    describe('Provider Schemas', () => {
        let testProviderId: string;

        beforeEach(async () => {
            const res = await request(app).post('/providers').send({
                providerName: "Test Provider",
                contactEmail: "test@provider.com",
                phoneNumber: "555-5555",
                providerType: "business"
            });
            testProviderId = res.body.data.providerId; // Access providerId from the data object
        });
        
        const createValidSchemaPayload = () => ({
            fieldConfigurations: {
                "PersonData": {
                    "nic": { source: 'fallback', isOwner: false, description: 'National ID Card number.' },
                    "fullName": { source: 'authoritative', isOwner: true, description: 'The full legal name.' }
                }
            }
        });

        it('POST /providers/:providerId/register_schema - should create a new schema submission', async () => {
            const res = await request(app)
                .post(`/providers/${testProviderId}/register_schema`)
                .send(createValidSchemaPayload());

            expect(res.statusCode).toEqual(201);
            expect(res.body.status).toBe('success');
            expect(res.body.data).toHaveProperty('submissionId');
            expect(res.body.data.providerId).toBe(testProviderId);
        });

        it('GET /providers/:providerId/schema - should return a specific schema', async () => {
            await request(app).post(`/providers/${testProviderId}/register_schema`).send(createValidSchemaPayload());
            const res = await request(app).get(`/providers/${testProviderId}/schema`);

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.providerId).toEqual(testProviderId);
        });
        
        it('POST /provider-schemas/:submissionId/review - should approve a pending schema', async () => {
            const creationRes = await request(app).post(`/providers/${testProviderId}/register_schema`).send(createValidSchemaPayload());
            const submissionId = creationRes.body.data.submissionId;

            const reviewRes = await request(app)
                .post(`/provider-schemas/${submissionId}/review`)
                .send({ decision: 'approve' });
            
            expect(reviewRes.statusCode).toEqual(200);
            expect(reviewRes.body.status).toBe('success');
            expect(reviewRes.body.data.status).toEqual('approved');
        });
    });
});


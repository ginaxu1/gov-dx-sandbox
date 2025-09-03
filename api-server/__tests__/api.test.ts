import request from 'supertest';
import app from '../src/index';
import { applicationsDB, providerProfilesDB, providerSchemasDB } from '../src/database';

// Helper to reset the in-memory databases before each test
const resetDatabases = () => {
    applicationsDB.length = 0;
    providerProfilesDB.length = 0;
    providerSchemasDB.length = 0;
};


describe('API Endpoints', () => {

    beforeEach(() => {
        // Reset the state of all in-memory DBs before each test
        resetDatabases();
    });

    // --- Tests for Consumer Application Endpoints ---
    describe('Consumer Applications [/applications]', () => {
        const validApplicationPayload = {
            appId: "passport-app-123",
            requiredFields: { "drp.personInfo.address": {} }
        };

        it('POST /applications - should create a new application', async () => {
            const res = await request(app)
                .post('/applications')
                .send(validApplicationPayload);

            expect(res.statusCode).toEqual(201);
            expect(res.body).toHaveProperty('appId', 'passport-app-123');
            expect(res.body).toHaveProperty('status', 'pending');
        });

        it('POST /applications - should fail with 409 if application already exists', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app).post('/applications').send(validApplicationPayload);
            expect(res.statusCode).toEqual(409);
            expect(res.body).toHaveProperty('error');
        });

        it('GET /applications - should return an array of all applications', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app).get('/applications');
            expect(res.statusCode).toEqual(200);
            expect(res.body.length).toBe(1);
        });

        it('GET /applications?status=pending - should return a filtered array of applications', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app).get('/applications?status=pending');
            expect(res.statusCode).toEqual(200);
            expect(res.body.length).toBe(1);
            expect(res.body[0].status).toBe('pending');
        });

        it('POST /applications/:appId/review - should approve a pending application', async () => {
            await request(app).post('/applications').send(validApplicationPayload);
            const res = await request(app)
                .post('/applications/passport-app-123/review')
                .send({ decision: 'approve' });

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toEqual('approved');
            expect(res.body).toHaveProperty('credentials');
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
            const res = await request(app)
                .post('/providers')
                .send(validProviderPayload);

            expect(res.statusCode).toEqual(201);
            expect(res.body).toHaveProperty('providerId');
            expect(res.body.providerName).toBe(validProviderPayload.providerName);
        });

        it('POST /providers - should fail with 400 for invalid providerType', async () => {
            const res = await request(app)
                .post('/providers')
                .send({ ...validProviderPayload, providerType: 'startup' });
            
            expect(res.statusCode).toEqual(400);
            expect(res.body).toHaveProperty('error');
        });

        it('POST /providers - should fail with 409 if provider name already exists', async () => {
            await request(app).post('/providers').send(validProviderPayload);
            const res = await request(app).post('/providers').send(validProviderPayload);
            expect(res.statusCode).toEqual(409);
        });
    });


    // --- Tests for Provider Schema Endpoints ---
    describe('Provider Schemas [/provider-schemas]', () => {
        let testProviderId: string;

        // Before each test in this suite, create a provider profile to associate schemas with
        beforeEach(async () => {
            const res = await request(app).post('/providers').send({
                providerName: "Test Provider",
                contactEmail: "test@provider.com",
                phoneNumber: "555-5555",
                providerType: "business"
            });
            testProviderId = res.body.providerId;
        });
        
        const createValidSchemaPayload = (providerId: string) => ({
            providerId: providerId,
            apiEndpoint: "https://api.testprovider.com/graphql",
            schema: "type Query { testData: String }",
            fieldMetadata: { fields: { "test.data": {} } }
        });

        it('POST /provider-schemas - should create a new schema submission for an existing provider', async () => {
            const res = await request(app)
                .post('/provider-schemas')
                .send(createValidSchemaPayload(testProviderId));

            expect(res.statusCode).toEqual(201);
            expect(res.body).toHaveProperty('submissionId');
            expect(res.body.providerId).toBe(testProviderId);
        });

        it('POST /provider-schemas - should fail with 404 if providerId does not exist', async () => {
            const res = await request(app)
                .post('/provider-schemas')
                .send(createValidSchemaPayload('prov_nonexistent'));
            
            expect(res.statusCode).toEqual(404);
        });

        it('GET /provider-schemas/:providerId - should return a specific schema', async () => {
            await request(app).post('/provider-schemas').send(createValidSchemaPayload(testProviderId));
            const res = await request(app).get(`/provider-schemas/${testProviderId}`);
            expect(res.statusCode).toEqual(200);
            expect(res.body.providerId).toEqual(testProviderId);
        });
        
        it('GET /provider-schemas - should return an array of all schema submissions', async () => {
            await request(app).post('/provider-schemas').send(createValidSchemaPayload(testProviderId));
            const res = await request(app).get('/provider-schemas');
            expect(res.statusCode).toEqual(200);
            expect(res.body.length).toBe(1);
        });

        it('POST /provider-schemas/:submissionId/review - should approve a pending schema', async () => {
            const creationRes = await request(app).post('/provider-schemas').send(createValidSchemaPayload(testProviderId));
            const submissionId = creationRes.body.submissionId;

            const reviewRes = await request(app)
                .post(`/provider-schemas/${submissionId}/review`)
                .send({ decision: 'approve' });
            
            expect(reviewRes.statusCode).toEqual(200);
            expect(reviewRes.body.status).toEqual('approved');
        });
    });
});


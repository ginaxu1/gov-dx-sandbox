import request from 'supertest';
import app from '../src/index';
// Corrected import path for the database file
import { applicationsDB, providerProfilesDB, providerSchemasDB, providerSubmissionsDB } from '../src/database';

// Helper to reset the in-memory databases before each test
const resetDatabases = () => {
    applicationsDB.length = 0;
    providerProfilesDB.length = 0;
    providerSchemasDB.length = 0;
    providerSubmissionsDB.length = 0;
};

describe('API Endpoints with Standardized Responses', () => {

    beforeEach(() => {
        resetDatabases();
    });

    // Tests for Consumer Application Endpoints
    describe('Consumer Applications', () => {
        it('GET /available-fields - should return the mock structure of available fields', async () => {
            const res = await request(app).get('/available-fields');

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data).toHaveProperty('drp');
        });

        it('POST /applications - should create a new application', async () => {
            const res = await request(app).post('/applications').send({
                appId: "passport-app-123",
                requiredFields: { "drp.PersonData.nic": {} }
            });

            expect(res.statusCode).toEqual(201);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('pending');
        });

        it('POST /applications with invalid payload - should return a 400 error', async () => {
            const res = await request(app).post('/applications').send({});
            expect(res.statusCode).toEqual(400);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('missing');
        });
    });

    // Tests for the Provider Onboarding Workflow
    describe('Provider Onboarding Workflow', () => {
        const validSubmissionPayload = {
            providerName: "New Test Provider",
            contactEmail: "contact@newprovider.com",
            phoneNumber: "555-1234",
            providerType: "business" as const
        };

        it('POST /provider-submissions - should create a new provider submission', async () => {
            const res = await request(app)
                .post('/provider-submissions')
                .send(validSubmissionPayload);

            expect(res.statusCode).toEqual(202); // 202 Accepted
            expect(res.body.status).toBe('success');
            expect(res.body.data).toHaveProperty('submissionId');
        });

        it('POST /provider-submissions/:submissionId/review (approve) - should approve a submission and create a permanent profile', async () => {
            // Step 1: Create the submission
            const submissionRes = await request(app).post('/provider-submissions').send(validSubmissionPayload);
            const submissionId = submissionRes.body.data.submissionId;

            // Step 2: Approve the submission
            const reviewRes = await request(app)
                .post(`/provider-submissions/${submissionId}/review`)
                .send({ decision: 'approve' });

            expect(reviewRes.statusCode).toEqual(200);
            expect(reviewRes.body.status).toBe('success');
            expect(reviewRes.body.data).toHaveProperty('providerId'); // The permanent ID
            expect(reviewRes.body.data.providerName).toBe(validSubmissionPayload.providerName);
        });

        it('POST /provider-submissions/:submissionId/review (reject) - should reject a submission', async () => {
            const submissionRes = await request(app).post('/provider-submissions').send(validSubmissionPayload);
            const submissionId = submissionRes.body.data.submissionId;

            const reviewRes = await request(app)
                .post(`/provider-submissions/${submissionId}/review`)
                .send({ decision: 'reject' });

            expect(reviewRes.statusCode).toEqual(200);
            expect(reviewRes.body.status).toBe('success');
            expect(reviewRes.body.data.status).toBe('rejected');
        });

        it('POST /provider-submissions with invalid payload - should return a 400 error', async () => {
            const res = await request(app)
                .post('/provider-submissions')
                .send({ providerName: 123 }); // Invalid type for providerName

            expect(res.statusCode).toEqual(400);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('missing one or more required fields');
        });

        it('POST /provider-submissions/:submissionId/review with non-existent ID - should return a 404 error', async () => {
            const res = await request(app)
                .post('/provider-submissions/non-existent-id/review')
                .send({ decision: 'approve' });

            expect(res.statusCode).toEqual(404);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('not found');
        });
    });

    // Tests for Schema Endpoints for Approved Providers
    describe('Provider Schemas', () => {
        let approvedProviderId: string;

        // Before each test, run the full onboarding workflow to get an approved provider
        beforeEach(async () => {
            const submissionRes = await request(app).post('/provider-submissions').send({
                providerName: "Approved Test Provider",
                contactEmail: "test@approved.com",
                phoneNumber: "555-5555",
                providerType: "business"
            });
            const submissionId = submissionRes.body.data.submissionId;

            const approvalRes = await request(app)
                .post(`/provider-submissions/${submissionId}/review`)
                .send({ decision: 'approve' });

            approvedProviderId = approvalRes.body.data.providerId;
        });

        const createValidSchemaPayload = () => ({
            fieldConfigurations: {
                "PersonData": {
                    "nic": { source: 'fallback', isOwner: false, description: 'National ID Card number.' }
                }
            }
        });

        it('POST /providers/:providerId/schemas - should create a schema for an approved provider', async () => {
            const res = await request(app)
                .post(`/providers/${approvedProviderId}/schemas`)
                .send(createValidSchemaPayload());

            expect(res.statusCode).toEqual(201);
            expect(res.body.status).toBe('success');
            // The response should now contain providerId, not submissionId
            expect(res.body.data).toHaveProperty('providerId');
            expect(res.body.data.providerId).toBe(approvedProviderId);
        });

        it('POST /providers/:providerId/schemas with invalid provider ID - should return a 404 error', async () => {
            const res = await request(app)
                .post('/providers/non-existent-id/schemas')
                .send(createValidSchemaPayload());

            expect(res.statusCode).toEqual(404);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('not found');
        });

        it('POST /providers/:providerId/schemas with invalid payload - should return a 400 error', async () => {
            const res = await request(app)
                .post(`/providers/${approvedProviderId}/schemas`)
                .send({ fieldConfigurations: {} });

            expect(res.statusCode).toEqual(400);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('required');
        });
    });

    // Tests for GET Endpoints
    describe('GET Endpoints', () => {
        it('GET /provider-schemas - should return a list of schemas', async () => {
            const res = await request(app).get('/provider-schemas');
            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(Array.isArray(res.body.data)).toBe(true);
        });
        
        it('GET /applications - should return a list of all applications', async () => {
            const res = await request(app).get('/applications');
            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(Array.isArray(res.body.data)).toBe(true);
        });

        it('GET /applications?status=pending - should return only pending applications', async () => {
            // Setup: create a pending and a non-pending application
            await request(app).post('/applications').send({ appId: "app-pending", requiredFields: {} });
            await request(app).post('/applications').send({ appId: "app-denied", requiredFields: {} });
            await request(app).post('/applications/app-denied/review').send({ decision: 'deny' });
            
            const res = await request(app).get('/applications?status=pending');
            
            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.length).toBe(1);
            expect(res.body.data[0].status).toBe('pending');
        });
    });

    // Tests for Consumer Application Review
    describe('Consumer Application Review', () => {
        let appId: string;

        beforeEach(async () => {
            const res = await request(app).post('/applications').send({ appId: "test-app-review", requiredFields: {} });
            appId = res.body.data.appId;
        });

        it('POST /applications/:appId/review (approve) - should approve a pending application', async () => {
            const res = await request(app)
                .post(`/applications/${appId}/review`)
                .send({ decision: 'approve' });

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('approved');
        });

        it('POST /applications/:appId/review (deny) - should deny a pending application', async () => {
            const res = await request(app)
                .post(`/applications/${appId}/review`)
                .send({ decision: 'deny' });

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('denied');
        });

        it('POST /applications/:appId/review with non-existent appId - should return a 404 error', async () => {
            const res = await request(app)
                .post('/applications/non-existent-app/review')
                .send({ decision: 'approve' });

            expect(res.statusCode).toEqual(404);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('not found');
        });

        it('POST /applications/:appId/review with invalid decision - should return a 400 error', async () => {
            const res = await request(app)
                .post(`/applications/${appId}/review`)
                .send({ decision: 'invalid' });

            expect(res.statusCode).toEqual(400);
            expect(res.body.status).toBe('error');
            expect(res.body.message).toContain('Invalid decision');
        });
    });

    // Tests for Provider Schema Review
    describe('Provider Schema Review', () => {
        let approvedProviderId: string;

        beforeEach(async () => {
            // Onboard a provider and create a schema
            const submissionRes = await request(app).post('/provider-submissions').send({
                providerName: "Approved Schema Provider",
                contactEmail: "schema@test.com",
                phoneNumber: "111-222-3333",
                providerType: "business"
            });
            const submissionId = submissionRes.body.data.submissionId;
            const approvalRes = await request(app).post(`/provider-submissions/${submissionId}/review`).send({ decision: 'approve' });
            approvedProviderId = approvalRes.body.data.providerId;

            await request(app).post(`/providers/${approvedProviderId}/schemas`).send({
                fieldConfigurations: { "PersonData": { "nic": { source: 'fallback', isOwner: false, description: 'National ID Card number.' } } }
            });
        });

        it('POST /provider-schemas/:providerId/review (approve) - should approve a pending schema', async () => {
            const res = await request(app)
                .post(`/provider-schemas/${approvedProviderId}/review`)
                .send({ decision: 'approve' });
            
            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('approved');
        });

        it('POST /provider-schemas/:providerId/review (changes_required) - should mark a pending schema for changes', async () => {
            const res = await request(app)
                .post(`/provider-schemas/${approvedProviderId}/review`)
                .send({ decision: 'request_changes' });

            expect(res.statusCode).toEqual(200);
            expect(res.body.status).toBe('success');
            expect(res.body.data.status).toBe('changes_required');
        });
    });
});

import { Router } from 'express';
import { 
    createApplication, 
    getAllApplications, 
    reviewApplication,
    getAvailableFields
} from '../controllers/consumerController';

const router = Router();

// Consumer Facing Endpoints

/**
 * @route GET /available-fields
 * @description Provides the data needed for the frontend to build the field selection UI
 */
router.get('/available-fields', getAvailableFields);

/**
 * @route POST /applications
 * @description Endpoint for a data consumer to submit a new application for review
 */
router.post('/applications', createApplication);


// Admin Facing Endpoints

/**
 * @route GET /applications
 * @description Retrieves a list of all consumer applications, with optional status filtering
 */
router.get('/applications', getAllApplications);

/**
 * @route POST /applications/:appId/review
 * @description Endpoint for an admin to approve or deny a pending consumer application
 */
router.post('/applications/:appId/review', reviewApplication);

export default router;


import { Request, Response } from "express";
import { applicationsDB } from "../database";
import { Application, ApplicationStatus } from "../types";
import { notifyAdmin, updatePdpPolicy, generateCredentials } from "../services";
import { sendSuccess, sendError } from '../utils/responseHandler';

// MOCK DATA for available schemas
// TODO: dynamically generate by reading all approved provider schemas
const MOCK_AVAILABLE_FIELDS = {
    "drp": { // providerId
        "displayName": "Dept. of Registration of Persons",
        "types": {
            "PersonData": {
                "fields": ["nic", "fullName", "permanentAddress", "photo"]
            },
            "Query": {
                "fields": ["person"]
            }
        }
    },
    "rgd": { // providerId
        "displayName": "Registrar General's Dept.",
        "types": {
            "BirthCertificate": {
                "fields": ["birthCertificateNo", "birthDate", "birthPlace"]
            },
            "Query": {
                "fields": ["birthInfo"]
            }
        }
    }
};
/**
 * @summary [CONSUMER] Get all available data fields from all approved providers
 */
export const getAvailableFields = (req: Request, res: Response) => {
    // TODO: in v2, query the providerSchemasDB, filter for 'approved' schemas, and transform them into this structure
    return sendSuccess(res, MOCK_AVAILABLE_FIELDS, 'Available fields retrieved successfully.');
};

/**
 * @summary [CONSUMER] Submit a new application for data access
 */
export const createApplication = (req: Request, res: Response) => {
  const { appId, requiredFields } = req.body;
  if (!appId || !requiredFields) {
    return sendError(res, "Invalid request body: missing 'appId' or 'requiredFields'", 400);
  }
  if (applicationsDB.some((app) => app.appId === appId)) {
    return sendError(res, `Application with appId '${appId}' already exists.`, 409);
  }

  const newApplication: Application = { appId, requiredFields, status: "pending" };
  applicationsDB.push(newApplication);
  notifyAdmin(`New application '${appId}' requires review.`);

  return sendSuccess(res, newApplication, 'Application submitted successfully.', 201);
};

/**
 * @summary [ADMIN] Get all consumer applications, optionally filtering by status
 */
export const getAllApplications = (req: Request, res: Response) => {
  const status = req.query.status as ApplicationStatus | undefined;
  if (status) {
    const filteredApps = applicationsDB.filter((app) => app.status === status);
    return sendSuccess(res, filteredApps);
  }
  return sendSuccess(res, applicationsDB);
};

/**
 * @summary [ADMIN] Approve or deny a consumer's application
 */
export const reviewApplication = async (req: Request, res: Response) => {
  const { appId } = req.params;
  const { decision } = req.body;

  if (decision !== "approve" && decision !== "deny") {
    return sendError(res, "Invalid decision. Must be 'approve' or 'deny'.", 400);
  }

  const application = applicationsDB.find((app) => app.appId === appId);
  if (!application) {
    return sendError(res, `Application with appId '${appId}' not found.`, 404);
  }

  if (application.status !== "pending") {
    return sendError(res, `Application is already '${application.status}' and cannot be reviewed again.`, 409);
  }

  if (decision === "approve") {
    application.status = "approved";
    await updatePdpPolicy(application);
    application.credentials = generateCredentials(appId);
    return sendSuccess(res, application, 'Application approved successfully.');
  } else {
    application.status = "denied";
    return sendSuccess(res, application, 'Application denied successfully.');
  }
};
import { Request, Response } from "express";
import { applicationsDB } from "../database";
import { Application, ApplicationStatus } from "../types";
import { notifyAdmin, updatePdpPolicy, generateCredentials } from "../services";

/**
 * @summary [CONSUMER] Submit a new application for data access
 */
export const createApplication = (req: Request, res: Response) => {
  const { appId, requiredFields } = req.body;
  if (!appId || !requiredFields) {
    return res
      .status(400)
      .json({
        error: "Invalid request body: missing 'appId' or 'requiredFields'",
      });
  }
  if (applicationsDB.some((app) => app.appId === appId)) {
    return res
      .status(409)
      .json({ error: `Application with appId '${appId}' already exists.` });
  }
  const newApplication: Application = {
    appId,
    requiredFields,
    status: "pending",
  };
  applicationsDB.push(newApplication);
  notifyAdmin(`New application '${appId}' requires review.`);
  return res.status(201).json(newApplication);
};

/**
 * @summary [ADMIN] Get all consumer applications, optionally filtering by status
 */
export const getAllApplications = (req: Request, res: Response) => {
  const status = req.query.status as ApplicationStatus | undefined;
  if (status) {
    return res
      .status(200)
      .json(applicationsDB.filter((app) => app.status === status));
  }
  return res.status(200).json(applicationsDB);
};

/**
 * @summary [ADMIN] Approve or deny a consumer's application
 */
export const reviewApplication = async (req: Request, res: Response) => {
  const { appId } = req.params;
  const { decision } = req.body;
  if (decision !== "approve" && decision !== "deny") {
    return res
      .status(400)
      .json({ error: "Invalid decision. Must be 'approve' or 'deny'." });
  }
  const application = applicationsDB.find((app) => app.appId === appId);
  if (!application) {
    return res
      .status(404)
      .json({ error: `Application with appId '${appId}' not found.` });
  }
  if (application.status !== "pending") {
    return res
      .status(409)
      .json({
        error: `Application is already '${application.status}' and cannot be reviewed again.`,
      });
  }
  if (decision === "approve") {
    application.status = "approved";
    await updatePdpPolicy(application);
    application.credentials = generateCredentials(appId);
  } else {
    application.status = "denied";
  }
  return res.status(200).json(application);
};

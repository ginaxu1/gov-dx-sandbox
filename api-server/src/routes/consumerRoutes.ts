import { Router } from "express";
import {
  createApplication,
  getAllApplications,
  reviewApplication,
} from "../controllers/consumerController";

const router = Router();

// Routes for the Data Consumer Portal & Admin Portal (Consumer actions)
router.post("/applications", createApplication);
router.get("/applications", getAllApplications);
router.post("/applications/:appId/review", reviewApplication);

export default router;

import { Router } from "express";
import {
  createProviderProfile,
  createProviderSchema,
  getAllProviderSchemas,
  getProviderSchema,
  reviewProviderSchema,
} from "../controllers/providerController";

const router = Router();

// Routes for the Data Provider Portal & Admin Portal (Provider actions)
router.post("/providers", createProviderProfile);
router.post("/provider-schemas", createProviderSchema);
router.get("/provider-schemas", getAllProviderSchemas);
router.get("/provider-schemas/:providerId", getProviderSchema);
router.post("/provider-schemas/:submissionId/review", reviewProviderSchema);

export default router;

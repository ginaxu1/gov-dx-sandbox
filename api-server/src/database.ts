import { Application, ProviderProfile, ProviderSchema, ProviderSubmission } from "./types";

// IN-MEMORY DATABASES
// TODO: in v2, replace with persistent database like PostgreSQL
export const applicationsDB: Application[] = [];
export const providerProfilesDB: ProviderProfile[] = [];
export const providerSchemasDB: ProviderSchema[] = [];
export const providerSubmissionsDB: ProviderSubmission[] = [];
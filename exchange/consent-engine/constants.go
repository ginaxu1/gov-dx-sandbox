package main

// ownerIDToEmailMap stores the mapping from owner_id to owner_email
// This is a fallback mapping for local development and testing.
// In production, this should be removed as owner_email will be resolved
// via SCIM API lookup using the Asgardeo integration.
//
// TODO: Remove this file once SCIM integration is fully tested and deployed
var ownerIDToEmailMap = map[string]string{
	"199512345678": "regina@opensource.lk",
	"199612345678": "regina@opensource.lk",
	"199712345678": "regina@opensource.lk",
	"199812345678": "regina@opensource.lk",
	"199912345678": "regina@opensource.lk",
	"200012345678": "mohamed@opensource.lk",
	"200112345678": "mohamed@opensource.lk",
	"200212345678": "mohamed@opensource.lk",
	"200312345678": "mohamed@opensource.lk",
	"200412345678": "mohamed@opensource.lk",
	"198712345678": "thanikan@opensource.lk",
	"200512345678": "thanikan@opensource.lk",
	"200612345678": "thanikan@opensource.lk",
	"200712345678": "thanikan@opensource.lk",
	"200812345678": "thanikan@opensource.lk",
}

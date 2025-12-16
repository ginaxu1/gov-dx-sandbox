export const PortalAction = {
    approve: "approve",
    reject: "reject",
} as const;

export type PortalAction = (typeof PortalAction)[keyof typeof PortalAction];
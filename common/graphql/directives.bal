// Define the enum for the classification levels
public enum Classification {
    ALLOW,
    ALLOW_PROVIDER_CONSENT,
    ALLOW_CITIZEN_CONSENT,
    ALLOW_CONSENT,
    DENIED
}

public type Permission record {
    Classification classification;
};
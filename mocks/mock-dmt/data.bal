// This is a mock data file for the DMT service

final isolated table<DrivingLicense> key(id) drivingLicenses = table [
    {
        id: "1",
        licenseNumber: "DL-123456",
        issueDate: "2020-01-01",
        expiryDate: "2030-01-01",
        frontImageUrl: "https://example.com/front1.jpg",
        backImageUrl: "https://example.com/back1.jpg",
        ownerInfo: {
            ownerNic: "199512345678",
            name: "John Doe",
            address: "123 Main St, Anytown, USA",
            birthDate: "1995-12-31",
            signatureUrl: "https://example.com/signature1.jpg",
            bloodGroup: A_POSITIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-1",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature1.jpg"
        }
    },
    {
        id: "2",
        licenseNumber: "DL-654321",
        issueDate: "2021-01-01",
        expiryDate: "2031-01-01",
        frontImageUrl: "https://example.com/front2.jpg",
        backImageUrl: "https://example.com/back2.jpg",
        ownerInfo: {
            ownerNic: "199612345678",
            name: "Jane Smith",
            address: "456 Elm St, Othertown, USA",
            birthDate: "1996-01-31",
            signatureUrl: "https://example.com/signature2.jpg",
            bloodGroup: B_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-2",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature2.jpg"
        }
    },
    {
        id: "3",
        licenseNumber: "DL-789012",
        issueDate: "2022-01-01",
        expiryDate: "2032-01-01",
        frontImageUrl: "https://example.com/front3.jpg",
        backImageUrl: "https://example.com/back3.jpg",
        ownerInfo: {
            ownerNic: "199712345678",
            name: "Alice Johnson",
            address: "789 Oak St, Sometown, USA",
            birthDate: "1997-02-28",
            signatureUrl: "https://example.com/signature3.jpg",
            bloodGroup: O_POSITIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-3",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature3.jpg"
        }
    },
    {
        id: "4",
        licenseNumber: "DL-345678",
        issueDate: "2023-01-01",
        expiryDate: "2033-01-01",
        frontImageUrl: "https://example.com/front4.jpg",
        backImageUrl: "https://example.com/back4.jpg",
        ownerInfo: {
            ownerNic: "199812345678",
            name: "Bob Brown",
            address: "321 Pine St, Anycity, USA",
            birthDate: "1998-03-15",
            signatureUrl: "https://example.com/signature4.jpg",
            bloodGroup: AB_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-4",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature4.jpg"
        }
    },
    {
        id: "5",
        licenseNumber: "DL-987654",
        issueDate: "2024-01-01",
        expiryDate: "2034-01-01",
        frontImageUrl: "https://example.com/front5.jpg",
        backImageUrl: "https://example.com/back5.jpg",
        ownerInfo: {
            ownerNic: "199912345678",
            name: "Charlie Davis",
            address: "654 Maple St, Anycity, USA",
            birthDate: "1999-04-20",
            signatureUrl: "https://example.com/signature5.jpg",
            bloodGroup: B_POSITIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-5",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature5.jpg"
        }
    },
    {
        id: "6",
        licenseNumber: "DL-123456",
        issueDate: "2025-01-01",
        expiryDate: "2035-01-01",
        frontImageUrl: "https://example.com/front6.jpg",
        backImageUrl: "https://example.com/back6.jpg",
        ownerInfo: {
            ownerNic: "200012345678",
            name: "David Wilson",
            address: "987 Cedar St, Anycity, USA",
            birthDate: "2000-05-10",
            signatureUrl: "https://example.com/signature6.jpg",
            bloodGroup: A_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-6",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature6.jpg"
        }
    },
    {
        id: "7",
        licenseNumber: "DL-765432",
        issueDate: "2026-01-01",
        expiryDate: "2036-01-01",
        frontImageUrl: "https://example.com/front7.jpg",
        backImageUrl: "https://example.com/back7.jpg",
        ownerInfo: {
            ownerNic: "200112345678",
            name: "Emma Thompson",
            address: "654 Birch St, Anycity, USA",
            birthDate: "2001-06-25",
            signatureUrl: "https://example.com/signature7.jpg",
            bloodGroup: O_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-7",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature7.jpg"
        }
    },
    {
        id: "8",
        licenseNumber: "DL-852963",
        issueDate: "2027-01-01",
        expiryDate: "2037-01-01",
        frontImageUrl: "https://example.com/front8.jpg",
        backImageUrl: "https://example.com/back8.jpg",
        ownerInfo: {
            ownerNic: "200212345678",
            name: "Frank Harris",
            address: "321 Oak St, Anycity, USA",
            birthDate: "2002-07-30",
            signatureUrl: "https://example.com/signature8.jpg",
            bloodGroup: AB_POSITIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-8",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature8.jpg"
        }
    },
    {
        id: "9",
        licenseNumber: "DL-159753",
        issueDate: "2028-01-01",
        expiryDate: "2038-01-01",
        frontImageUrl: "https://example.com/front9.jpg",
        backImageUrl: "https://example.com/back9.jpg",
        ownerInfo: {
            ownerNic: "200312345678",
            name: "Grace Lee",
            address: "654 Pine St, Anycity, USA",
            birthDate: "2003-08-15",
            signatureUrl: "https://example.com/signature9.jpg",
            bloodGroup: B_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-9",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature9.jpg"
        }
    },
    {
        id: "10",
        licenseNumber: "DL-987654",
        issueDate: "2029-01-01",
        expiryDate: "2039-01-01",
        frontImageUrl: "https://example.com/front10.jpg",
        backImageUrl: "https://example.com/back10.jpg",
        ownerInfo: {
            ownerNic: "200412345678",
            name: "Henry Kim",
            address: "321 Maple St, Anycity, USA",
            birthDate: "2004-09-10",
            signatureUrl: "https://example.com/signature10.jpg",
            bloodGroup: AB_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-10",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature10.jpg"
        }
    },
    {
        id: "11",
        licenseNumber: "DL-123456",
        issueDate: "2030-01-01",
        expiryDate: "2040-01-01",
        frontImageUrl: "https://example.com/front11.jpg",
        backImageUrl: "https://example.com/back11.jpg",
        ownerInfo: {
            ownerNic: "200512345678",
            name: "Irene Adler",
            address: "987 Cedar St, Anycity, USA",
            birthDate: "2005-10-20",
            signatureUrl: "https://example.com/signature11.jpg",
            bloodGroup: O_NEGATIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-11",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature11.jpg"
        }
    },
    {
        id: "12",
        licenseNumber: "DL-654321",
        issueDate: "2031-01-01",
        expiryDate: "2041-01-01",
        frontImageUrl: "https://example.com/front12.jpg",
        backImageUrl: "https://example.com/back12.jpg",
        ownerInfo: {
            ownerNic: "200612345678",
            name: "John Doe",
            address: "123 Elm St, Anycity, USA",
            birthDate: "2006-11-30",
            signatureUrl: "https://example.com/signature12.jpg",
            bloodGroup: A_POSITIVE
        },
        permissions: [],
        issuerInfo: {
            id: "issuer-12",
            name: "Department of Motor Vehicles",
            issuingAuthority: "DMV",
            signatureUrl: "https://example.com/issuer-signature12.jpg"
        }
    }
];

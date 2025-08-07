final isolated table<DriverLicense> key(id) licenseData = table [
    {id: "dl-abc", licenseNumber: "D12345678", issueDate: "2020-10-10", expiryDate: "2025-10-09", ownerId: "u-123", photoUrl: "http://example.com/photo1.jpg"},
    {id: "dl-def", licenseNumber: "D87654321", issueDate: "2022-01-01", expiryDate: "2027-12-31", ownerId: "u-456", photoUrl: "http://example.com/photo2.jpg"}
];

final isolated table<VehicleInfo> key(id) vehicleData = table [
    {id: "v-123", make: "Toyota", model: "Camry", yearOfManufacture: 2019, ownerId: "u-123", engineNumber: "EN123456789", conditionAndNotes: "Good condition", vehicleClass: {id: "vc-1", className: "Sedan"}, registrationNumber: "CEO-5678"},
    {id: "v-456", make: "Honda", model: "Civic", yearOfManufacture: 2020, ownerId: "u-456", engineNumber: "EN987654321", conditionAndNotes: "Excellent condition", vehicleClass: {id: "vc-2", className: "Hatchback"}, registrationNumber: "BEO-1234"},
    {id: "v-789", make: "Ford", model: "Focus", yearOfManufacture: 2021, ownerId: "u-123", engineNumber: "EN112233445", conditionAndNotes: "Minor scratches", vehicleClass: {id: "vc-1", className: "Sedan"}, registrationNumber: "BEO-5678"}
];

final isolated table<VehicleClass> key(id) vehicleClassData = table [
    {id: "vc-1", className: "Sedan"},
    {id: "vc-2", className: "Hatchback"}
];
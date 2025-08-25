import React, { useEffect, useState } from "react";
import { gql, useLazyQuery, useQuery } from "@apollo/client";

const GET_PERSON_DATA = gql`
  query Test($nic: ID!) {
    person(nic: $nic) {
      fullName
      nic
      permanentAddress
      photo
    }
  }
`;

function PassportApplicationForm({ onClose, nic, userInfo }) {
  const { loading, error, data } = useQuery(GET_PERSON_DATA, {
    variables: { nic },
    fetchPolicy: "network-only",
  });

  const [uploadedPhoto, setUploadedPhoto] = useState(null);
  const [uploadStatus, setUploadStatus] = useState(
    "Please upload a passport-sized ICAO compliant ",
  );

  useEffect(() => {
    if (data) {
      console.log("Fetched person data:", data);
      if (data.person?.photo) {
        setUploadStatus("Photo from DRP loaded");
      }
    }
  }, [data]);

  const handlePhotoUpload = (event) => {
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setUploadedPhoto(e.target.result);
        setUploadStatus("Photo uploaded successfully.");
      };
      reader.readAsDataURL(file);
    }
  };

  const isDataAvailable = data && data.person;
  let photoSource = null;
  if (uploadedPhoto) {
    photoSource = uploadedPhoto;
  } else if (isDataAvailable && data.person.photo) {
    photoSource = `data:image/jpeg;base64,${data.person.photo}`;
  }

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h3 className="modal-title">Passport Application Form</h3>
          <button
            onClick={onClose}
            className="modal-close-button"
            aria-label="Close"
          >
            &times;
          </button>
        </div>

        {loading ? (
          <p className="modal-status loading">
            Loading your data from the Exchange...
          </p>
        ) : error ? (
          <p className="modal-status error">
            Error fetching data: {error.message}
          </p>
        ) : isDataAvailable ? (
          <div className="form-container">

            <div className="photo-section">
              <label className="form-label">Photo: </label>
              {photoSource ? (
                <img
                  src={photoSource}
                  alt="Applicant Photo"
                  className="applicant-photo"
                  onError={(e) => {
                    e.currentTarget.onerror = null;
                    e.currentTarget.src =
                      "https://placehold.co/128x128/334155/E2E8F0?text=No+Photo";
                  }}
                />
              ) : (
                <div className="no-photo-placeholder">No Photo</div>
              )}
              <p className="upload-status">{uploadStatus}</p>
              <input
                type="file"
                id="photo-upload"
                onChange={handlePhotoUpload}
                accept="image/jpeg, image/png"
              />
            </div>
            <p className="form-intro-text">
              Some fields below have been pre-filled with your data from the
              OpenDIF Exchange
            </p>
            <div className="form-grid">
              {/* Pre-filled Fields */}
              <div className="form-field">
                <label className="form-label">Full Name: </label>
                <span className="form-value">{data.person.fullName}</span>
              </div>
              <div className="form-field">
                <label className="form-label">NIC: </label>
                <span className="form-value">{data.person.nic}</span>
              </div>
              <div className="form-field">
                <label className="form-label">Date of Birth: </label>
                <span className="form-value">{data.person.dateOfBirth}</span>
              </div>
              <div className="form-field">
                <label className="form-label">Permanent Address: </label>
                <span className="form-value">
                  {data.person.permanentAddress}
                </span>
              </div>
              <div className="form-field">
                <label className="form-label">Father's Name: </label>
                <span className="form-value">
                  {data.person.parentInfo?.fatherName}
                </span>
              </div>
              <div className="form-field">
                <label className="form-label">Mother's Name: </label>
                <span className="form-value">
                  {data.person.parentInfo?.motherName}
                </span>
              </div>
            </div>
            <h4 className="manual-fields-heading">
              Please manually fill in these fields:
            </h4>
            <div className="form-grid">
              <div className="form-field">
                <label htmlFor="email" className="form-label">
                  Email Address:{" "}
                </label>
                <input
                  type="email"
                  id="email"
                  defaultValue={userInfo?.email || ""}
                  className="form-input"
                />
              </div>
              <div className="form-field">
                <label htmlFor="emergency-contact" className="form-label">
                  Emergency Contact:{" "}
                </label>
                <input
                  type="tel"
                  id="emergency-contact"
                  className="form-input"
                />
              </div>
            </div>
            <div className="form-submit-section">
              <button className="submit-button">
                Submit Application and Pay with GovPay
              </button>
            </div>
          </div>
        ) : (
          <div className="form-container">
            <div className="no-data-info">
              <p>No data found for the provided NIC ({nic}).</p>
              <p>Please proceed to manually fill out the entire form.</p>
            </div>
            {/* Manual Form */}
            <div className="form-grid">
              <div className="form-field">
                <label htmlFor="manual-fullName" className="form-label">
                  Full Name:
                </label>
                <input
                  type="text"
                  id="manual-fullName"
                  className="form-input"
                />
              </div>
              <div className="form-field">
                <label htmlFor="manual-nic" className="form-label">
                  NIC:
                </label>
                <input type="text" id="manual-nic" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-dob" className="form-label">
                  Date of Birth:
                </label>
                <input type="date" id="manual-dob" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-address" className="form-label">
                  Permanent Address:
                </label>
                <input type="text" id="manual-address" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-father" className="form-label">
                  Father's Name:
                </label>
                <input type="text" id="manual-father" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-mother" className="form-label">
                  Mother's Name:
                </label>
                <input type="text" id="manual-mother" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-email" className="form-label">
                  Email Address:
                </label>
                <input type="email" id="manual-email" className="form-input" />
              </div>
              <div className="form-field">
                <label
                  htmlFor="manual-emergency-contact"
                  className="form-label"
                >
                  Emergency Contact:
                </label>
                <input
                  type="tel"
                  id="manual-emergency-contact"
                  className="form-input"
                />
              </div>
            </div>
            <div className="form-submit-section">
              <button className="submit-button">
                Pay and Submit Application
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default PassportApplicationForm;

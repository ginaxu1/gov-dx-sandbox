import React, { useEffect } from 'react';
import { gql, useLazyQuery, useQuery } from '@apollo/client';

// GraphQL query
const GET_PERSON_DATA = gql`
  query MyQuery($nic: String!) {
    person(nic: $nic) {
        fullName
        nic
        dateOfBirth
        permanentAddress
        photo
        parentInfo {
          fatherName
          motherName
        }
      }
  }
`;

function PassportApplicationForm({ onClose, nic, userInfo }) {
  const { loading, error, data } = useQuery(GET_PERSON_DATA, {
    variables: { nic },
    fetchPolicy: 'network-only',
  });

  // print when the data arrives
  useEffect(() => {
    if (data) {
      console.log(">>>Fetched person data:", data);
    }
  }, [data]);

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h3 className="modal-title">Passport Application Form</h3>
          <button onClick={onClose} className="modal-close-button" aria-label="Close">
            &times;
          </button>
        </div>

        {loading ? (
          <p className="modal-status loading">Loading your data from the Exchange...</p>
        ) : error ? (
          <p className="modal-status error">Error fetching data: {error.message}</p>
        ) : data?.person ? (
          <div className="form-container">
            <p className="form-intro-text">Most fields below have been pre-filled from your data via the OpenDIF Exchange.</p>
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
                <span className="form-value">{data.person.permanentAddress}</span>
              </div>
              <div className="form-field">
                <label className="form-label">Father's Name: </label>
                <span className="form-value">{data.person.parentInfo?.fatherName}</span>
              </div>
              <div className="form-field">
                <label className="form-label">Mother's Name: </label>
                <span className="form-value">{data.person.parentInfo?.motherName}</span>
              </div>
            </div>

            <div className="photo-section">
              <label className="form-label">Photo: </label>
              {data.person.photo ? (
                <img
                  src={`data:image/jpeg;base64,${data.person.photo}`}
                  alt="Applicant Photo"
                  className="applicant-photo"
                  onError={(e) => { e.currentTarget.onerror = null; e.currentTarget.src = 'https://placehold.co/128x128/334155/E2E8F0?text=No+Photo'; }}
                />
              ) : (
                <div className="no-photo-placeholder">No Photo Available</div>
              )}
            </div>

            <hr className="form-divider" />

            <h4 className="manual-fields-heading">Please manually fill in these fields:</h4>
            <div className="form-grid">
              <div className="form-field">
                <label htmlFor="email" className="form-label">Email Address: </label>
                <input
                  type="email"
                  id="email"
                  defaultValue={userInfo?.email || ''}
                  className="form-input"
                />
              </div>
              <div className="form-field">
                <label htmlFor="emergency-contact" className="form-label">Emergency Contact: </label>
                <input
                  type="tel"
                  id="emergency-contact"
                  className="form-input"
                />
              </div>
            </div>
            <div className="form-submit-section">
              <button className="submit-button">Pay and Submit Application</button>
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
                <label htmlFor="manual-fullName" className="form-label">Full Name:</label>
                <input type="text" id="manual-fullName" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-nic" className="form-label">NIC:</label>
                <input type="text" id="manual-nic" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-dob" className="form-label">Date of Birth:</label>
                <input type="date" id="manual-dob" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-address" className="form-label">Permanent Address:</label>
                <input type="text" id="manual-address" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-father" className="form-label">Father's Name:</label>
                <input type="text" id="manual-father" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-mother" className="form-label">Mother's Name:</label>
                <input type="text" id="manual-mother" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-email" className="form-label">Email Address:</label>
                <input type="email" id="manual-email" className="form-input" />
              </div>
              <div className="form-field">
                <label htmlFor="manual-emergency-contact" className="form-label">Emergency Contact:</label>
                <input type="tel" id="manual-emergency-contact" className="form-input" />
              </div>
            </div>
            <div className="form-submit-section">
              <button className="submit-button">Pay and Submit Application</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default PassportApplicationForm;
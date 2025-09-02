package serializer

import (
	"strings"
	"testing"
)

func TestSerializeGraphQLQuery(t *testing.T) {
	rawQuery := `
        query GetUser($nic: String!) {
			drp {
            	person(nic: $nic) {
                	id
                	name
            	}
			}
			dmt {
				vehicle {
					getVehicleInfos(ownerNic: $nic) {
						data {
							model
						}
					}
				}
			}
        }
    `

	expectedSubstr := "query GetUser" // we just check substring, since printer may format slightly

	result, err := serializeGraphQLQuery(rawQuery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, expectedSubstr) {
		t.Errorf("expected query to contain %q, got %q", expectedSubstr, result)
	}

	// Ensure it preserved field selections
	if !strings.Contains(result, "user") || !strings.Contains(result, "name") {
		t.Errorf("expected fields not found in result: %q", result)
	}
}

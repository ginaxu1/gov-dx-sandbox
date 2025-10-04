package tests

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

func TestContractTester_NewContractTester(t *testing.T) {
	// Test creating a new contract tester
	tester := services.NewContractTester(nil)
	assert.NotNil(t, tester)
}

func TestContractTester_ExecuteContractTests(t *testing.T) {
	// Create test schema
	queryType := &ast.ObjectDefinition{
		Name: &ast.Name{Value: "Query"},
		Fields: []*ast.FieldDefinition{
			{
				Name: &ast.Name{Value: "hello"},
				Type: &ast.Named{Name: &ast.Name{Value: "String"}},
			},
		},
	}

	testSchema := &ast.Document{
		Definitions: []ast.Node{queryType},
	}

	// Create contract tester
	tester := services.NewContractTester(nil)

	// Execute test - should fail with nil database
	results, err := tester.ExecuteContractTests(testSchema)

	// Assertions - should fail without database
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestContractTester_LoadContractTests(t *testing.T) {
	// Create contract tester
	tester := services.NewContractTester(nil)

	// Execute test
	tests, err := tester.LoadContractTests()

	// Assertions - should fail without database
	assert.Error(t, err)
	assert.Nil(t, tests)
}

func TestContractTest_Structure(t *testing.T) {
	// Test ContractTest structure
	test := services.ContractTest{
		Name:        "test1",
		Query:       "query { hello }",
		Description: "Test query",
		IsActive:    true,
		Priority:    1,
	}

	assert.Equal(t, "test1", test.Name)
	assert.Equal(t, "query { hello }", test.Query)
	assert.Equal(t, "Test query", test.Description)
	assert.True(t, test.IsActive)
	assert.Equal(t, 1, test.Priority)
}

func TestContractTestResults_Structure(t *testing.T) {
	// Test ContractTestResults structure
	results := &services.ContractTestResults{
		TotalTests: 5,
		Passed:     4,
		Failed:     1,
		Results:    []services.TestResult{},
	}

	assert.Equal(t, 5, results.TotalTests)
	assert.Equal(t, 4, results.Passed)
	assert.Equal(t, 1, results.Failed)
	assert.NotNil(t, results.Results)
}

func TestTestResult_Structure(t *testing.T) {
	// Test TestResult structure
	result := services.TestResult{
		TestName: "test1",
		Passed:   true,
		Error:    "",
		Duration: 100,
	}

	assert.Equal(t, "test1", result.TestName)
	assert.True(t, result.Passed)
	assert.Empty(t, result.Error)
	assert.Equal(t, int64(100), result.Duration)
}

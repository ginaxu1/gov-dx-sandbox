package federator

import (
	"reflect"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
)

func Test_splitQuery(t *testing.T) {
	type args struct {
		rawQuery string
	}
	tests := []struct {
		name string
		args args
		want []*federationServiceRequest
	}{
		{
			name: "Test Case 1: Multiple Services",
			args: args{
				rawQuery: `query MyQuery { drp { person(nic: "199512345678") { nic photo } } dmt { vehicle { getVehicleInfos { data { model } } } } }`,
			},
			want: []*federationServiceRequest{
				{
					ServiceKey: "drp",
					GraphQLRequest: graphql.Request{
						Query: `query MyQuery { person(nic: "199512345678") { nic photo } }`,
					},
				},
				{
					ServiceKey: "dmt",
					GraphQLRequest: graphql.Request{
						Query: `query MyQuery { vehicle { getVehicleInfos { data { model } } } }`,
					},
				},
			},
		},
		{
			name: "Test Case 2 - Single Service",
			args: args{
				rawQuery: `query MyQuery { drp { person(nic: "199512345678") { nic photo } } }`,
			},
			want: []*federationServiceRequest{
				{
					ServiceKey: "drp",
					GraphQLRequest: graphql.Request{
						Query: `query MyQuery { person(nic: "199512345678") { nic photo } }`,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitQuery(tt.args.rawQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

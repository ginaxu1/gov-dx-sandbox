package federator

import (
	"reflect"
	"testing"
)

func Test_splitQuery(t *testing.T) {
	type args struct {
		rawQuery string
	}
	tests := []struct {
		name string
		args args
		want []*FederationServiceRequest
	}{
		{
			name: "Test Case 1",
			args: args{
				rawQuery: `query MyQuery { drp { person(nic: "199512345678") { nic photo } } dmt { vehicle { getVehicleInfos { data { model } } } } }`,
			},
			want: []*FederationServiceRequest{
				{
					ServiceKey: "drp",
					GraphqlQuery: GraphQLRequest{
						Query: `query MyQuery { person(nic: "199512345678") { nic photo } }`,
					},
				},
				{
					ServiceKey: "dmt",
					GraphqlQuery: GraphQLRequest{
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
			want: []*FederationServiceRequest{
				{
					ServiceKey: "drp",
					GraphqlQuery: GraphQLRequest{
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

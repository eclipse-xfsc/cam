// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

// package workload contains service specific code for the Workload Configuration Collection Module.
package workload

import (
	"context"
	"os"
	"testing"

	. "clouditor.io/clouditor/api"
	"clouditor.io/clouditor/voc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/client-go/kubernetes"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil/testevaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

func TestMain(m *testing.M) {
	Server, _, _ := testevaluation.StartBufConnServerToEvaluation()

	code := m.Run()

	Server.Stop()

	os.Exit(code)
}

func Test_Server_StartCollecting(t *testing.T) {
	type envVariable struct {
		hasEnvVariable   bool
		envVariableKey   string
		envVariableValue string
	}

	type fields struct {
		streams      *StreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence]
		grpcOpts     []grpc.DialOption
		envVariables []envVariable
	}

	type args struct {
		in0 context.Context
		req *collection.StartCollectingRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Invalid request (wrong service id)",
			fields: fields{
				streams:  NewStreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](),
				grpcOpts: []grpc.DialOption{grpc.WithContextDialer(testevaluation.BufConnDialer)},
			},
			args: args{
				in0: context.Background(),
				req: &collection.StartCollectingRequest{
					EvalManager: "bufnet",
					ServiceId:   "serviceID",
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				assert.Equal(t, status.Code(err), codes.InvalidArgument)
				return assert.ErrorContains(t, err, collection.ErrRequestServiceID.Error())
			},
		},
		{
			name: "Invalid request (empty rawConfiguration)",
			fields: fields{
				streams:  NewStreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](),
				grpcOpts: []grpc.DialOption{grpc.WithContextDialer(testevaluation.BufConnDialer)},
			},
			args: args{
				in0: context.Background(),
				req: &collection.StartCollectingRequest{
					EvalManager: "bufnet",
					ServiceId:   "00000000-0000-0000-0000-000000000000",
					Configuration: &collection.ServiceConfiguration{
						ServiceId:        "00000000-0000-0000-0000-000000000000",
						RawConfiguration: nil,
					},
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				assert.Equal(t, status.Code(err), codes.InvalidArgument)
				return assert.ErrorContains(t, err, collection.ErrMissingRawConfiguration.Error())
			},
		},
		{
			// TODO(anatheka)
			name: "Error in rawConfiguration",
			fields: fields{
				streams:  NewStreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](),
				grpcOpts: []grpc.DialOption{grpc.WithContextDialer(testevaluation.BufConnDialer)},
			},
			args: args{
				in0: context.Background(),
				req: &collection.StartCollectingRequest{
					EvalManager: "bufnet",
					ServiceId:   "00000000-0000-0000-0000-000000000000",

					Configuration: &collection.ServiceConfiguration{
						ServiceId: "00000000-0000-0000-0000-000000000000",
						RawConfiguration: testproto.NewAny(t, &collection.WorkloadSecurityConfig{
							Kubernetes: &structpb.Value{
								Kind: &structpb.Value_StructValue{
									StructValue: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"Kubernetes": {Kind: &structpb.Value_StringValue{
												StringValue: "testKubernetes"}},
										},
									},
								},
							},
						}),
					},
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				assert.Equal(t, status.Code(err), codes.Internal)
				return assert.ErrorContains(t, err, "could not add provider configuration:")
			},
		},
		// {
		// TODO(anatheka: Add test when I know how the service config must look like

		// 	name: "Error in GetStream()",
		// 	fields: fields{
		// 		streams: NewStreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](),
		// 	},
		// 	args: args{
		// 		in0: context.Background(),
		// 		req: &collection.StartCollectingRequest{
		// 			ServiceId:   "00000000-0000-0000-0000-000000000000",
		// 			EvalManager: "wrongURL",
		// 			Configuration: &collection.ServiceConfiguration{
		// 				ServiceId:        "00000000-0000-0000-0000-000000000000",
		// 				CollectionModule: collection.ServiceConfiguration_WORKLOAD_CONFIGURATION,
		// 				RawConfiguration: &structpb.Value{
		// 					Kind: &structpb.Value_StructValue{
		// 						StructValue: &structpb.Struct{
		// 							Fields: map[string]*structpb.Value{
		// 								"Openstack":  {Kind: &structpb.Value_StringValue{StringValue: "testOpenstack"}},
		// 								"Kubernetes": {Kind: &structpb.Value_StringValue{StringValue: "testKubernetes"}},
		// 							},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
		// 		errResponse := fmt.Sprintf("could not add stream for %s with target '%s':", "Evaluation Manager", "wrongURL")
		// 		return assert.ErrorContains(t, err, errResponse)
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				stream:          tt.fields.streams,
				grpcOpts:        tt.fields.grpcOpts,
				providerConfigs: make(map[string]providerConfiguration),
			}

			// Set env variables
			for _, env := range tt.fields.envVariables {
				if env.hasEnvVariable {
					t.Setenv(env.envVariableKey, env.envVariableValue)
				}
			}

			gotResp, err := s.StartCollecting(tt.args.in0, tt.args.req)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, gotResp)
			}
		})
	}
}

func Test_getWorkloadConfigurations(t *testing.T) {
	type args struct {
		req *collection.StartCollectingRequest
	}
	type fields struct {
		providerConfigs map[string]providerConfiguration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []voc.IsCloudResource
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Invalid serviceID",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:   "",
					EvalManager: "bufnet",
					Configuration: &collection.ServiceConfiguration{
						RawConfiguration: testproto.NewAny(t, &collection.WorkloadSecurityConfig{}),
					},
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, "no discoverer available")
			},
		},
		{
			name: "Empty serviceConfiguration",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:     "00000000-0000-0000-0000-000000000000",
					EvalManager:   "bufnet",
					Configuration: &collection.ServiceConfiguration{},
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, "no discoverer available")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				providerConfigs: tt.fields.providerConfigs,
			}
			got, err := srv.getWorkloadConfigurations(tt.args.req)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func Test_Server_addProviderConfig(t *testing.T) {
	type fields struct {
		providerConfigs map[string]providerConfiguration
	}
	type args struct {
		req  *collection.StartCollectingRequest
		conf *collection.WorkloadSecurityConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ServiceID already available in provider configs",
			fields: fields{
				providerConfigs: map[string]providerConfiguration{
					"00000000-0000-0000-0000-000000000000": {
						kubernetes: &kubernetes.Clientset{},
					},
				},
			},
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:     "00000000-0000-0000-0000-000000000000",
					EvalManager:   "bufnet",
					Configuration: &collection.ServiceConfiguration{},
				},
			},
		},
		{
			name: "Empty serviceConfiguration",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:     "00000000-0000-0000-0000-000000000000",
					EvalManager:   "bufnet",
					Configuration: nil,
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrMissingServiceConfiguration.Error())
			},
		},
		{
			name: "Empty Kubernetes configuration",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId: "00000000-0000-0000-0000-000000000000",
				},
				conf: &collection.WorkloadSecurityConfig{
					Openstack: nil,
					Kubernetes: &structpb.Value{
						Kind: nil,
					},
				},
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrInvalidKubernetesServiceConfiguration.Error())
			},
		},
		// {
		// TODO(anatheka): Add test when I know how the kube config look like.
		// 	name: "Empty OpenStack configuration",
		// 	fields: fields{
		// 		providerConfigs: make(map[string]providerConfiguration),
		// 	},
		// 	args: args{
		// 		req: &collection.StartCollectingRequest{
		// 			ServiceId: "00000000-0000-0000-0000-000000000000",
		// 		},
		// 		conf: Config{
		// TODO(anatheka): How does a correct kube config look like?
		// 			Kubernetes: &structpb.Value{
		// 				Kind: &structpb.Value_StringValue{
		// 					StringValue: "testKubernetes",
		// 				},
		// 			},
		// 		},
		// 	},
		// 	wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
		// 		return assert.ErrorContains(t, err, collection.ErrInvalidOpenstackServiceConfiguration.Error())
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				providerConfigs: tt.fields.providerConfigs,
			}

			err := srv.addProviderConfig(tt.args.req, tt.args.conf)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Server_kubeConfig(t *testing.T) {
	type fields struct {
		providerConfigs map[string]providerConfiguration
	}
	type args struct {
		value     *structpb.Value
		serviceId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO(anatheka: Add test for correct service configuration when I know how it looks like
		{
			name: "Empty Kubernetes service configuration protobuf value",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				value:     &structpb.Value{},
				serviceId: "00000000-0000-0000-0000-000000000000",
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrConversionProtobufToByteArray.Error())
			},
		},
		{
			name: "Invalid Kubernetes service configuration protobuf value",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				value: &structpb.Value{
					Kind: &structpb.Value_StringValue{
						StringValue: "testKubernetes",
					},
				},
				serviceId: "00000000-0000-0000-0000-000000000000",
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrKubernetesClientset.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				providerConfigs: tt.fields.providerConfigs,
			}

			err := srv.kubeConfig(tt.args.value, tt.args.serviceId)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_Server_openstackConfig(t *testing.T) {
	type fields struct {
		providerConfigs map[string]providerConfiguration
	}
	type args struct {
		value     *structpb.Value
		serviceId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO(anatheka: Add test for correct service configuration when I know how it looks like
		{
			name: "Empty Openstack service configuration protobuf value",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				value:     &structpb.Value{},
				serviceId: "00000000-0000-0000-0000-000000000000",
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrConversionProtobufToAuthOptions.Error())
			},
		},
		{
			name: "Invalid Openstack service configuration protobuf value",
			fields: fields{
				providerConfigs: make(map[string]providerConfiguration),
			},
			args: args{
				value: &structpb.Value{
					Kind: &structpb.Value_StringValue{
						StringValue: "testOpenstack",
					},
				},
				serviceId: "00000000-0000-0000-0000-000000000000",
			},
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorContains(t, err, collection.ErrConversionProtobufToAuthOptions.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				providerConfigs: tt.fields.providerConfigs,
			}

			err := srv.openstackConfig(tt.args.value, tt.args.serviceId)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

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

package integrity

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	clouditor_api "clouditor.io/clouditor/api"
	"github.com/Fraunhofer-AISEC/cmc/attestationreport"
	ci "github.com/Fraunhofer-AISEC/cmc/cmcinterface"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil/testevaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

// TestCA used to verify the integrity information
var (
	tcpAddr *net.TCPAddr
	signer  *mockSigner
)

type cmcMockServer struct {
	ci.UnimplementedCMCServiceServer
}

type mockSigner struct {
	certChain attestationreport.CertChain
	priv      crypto.PrivateKey
}

func NewMockSigner() (*mockSigner, error) {
	ms := &mockSigner{}

	// Generate private key and public key for test CA
	caPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	caTmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Mock Server Test CA Cert",
			Country:      []string{"DE"},
			Province:     []string{"BY"},
			Locality:     []string{"Munich"},
			Organization: []string{"Test Company"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caPriv.PublicKey, caPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	tmp := &bytes.Buffer{}
	pem.Encode(tmp, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	ms.certChain.Ca = tmp.Bytes()

	// Generate private key and certificate for test prover, signed by test CA
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ms.priv = priv

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Mock Server Test Key Cert",
			Country:      []string{"DE"},
			Province:     []string{"BY"},
			Locality:     []string{"Munich"},
			Organization: []string{"Test Company"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	der, err = x509.CreateCertificate(rand.Reader, &tmpl, &caTmpl, &priv.PublicKey, caPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	tmp = &bytes.Buffer{}
	pem.Encode(tmp, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	ms.certChain.Leaf = tmp.Bytes()

	return ms, nil
}

func (s *mockSigner) Lock() {}

func (s *mockSigner) Unlock() {}

func (s *mockSigner) GetSigningKeys() (crypto.PrivateKey, crypto.PublicKey, error) {
	return s.priv, &s.priv.(*ecdsa.PrivateKey).PublicKey, nil
}

func (s *mockSigner) GetCertChain() attestationreport.CertChain {
	return s.certChain
}

func TestMain(m *testing.M) {

	// Setup logger
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.TraceLevel)
	log := logrus.WithField("testing", "collection-integrity")

	ms, err := NewMockSigner()
	if err != nil {
		log.Error("Failed to generate signer")
		return
	}
	signer = ms

	// Setup up dummy gRPC server to simulate the prover side, i.e. the remote service
	// that is queried by the integrity module
	addr := "127.0.0.1:0"
	log.Infof("Starting on %v", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("Failed to start server on %v: %v", addr, err)
		return
	}
	s := grpc.NewServer()
	ci.RegisterCMCServiceServer(s, &cmcMockServer{})
	log.Infof("Waiting for requests on %v", listener.Addr())

	// Get the tcp Address with dynamically chosen port
	tcpAddr = listener.Addr().(*net.TCPAddr)

	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	// Mock the Evaluation Manager to stream evidences to
	evaluationManagerServer, _, _ := testevaluation.StartBufConnServerToEvaluation()

	// Run the tests
	code := m.Run()

	evaluationManagerServer.Stop()

	os.Exit(code)
}

func (s *cmcMockServer) Attest(_ context.Context, in *ci.AttestationRequest) (*ci.AttestationResponse, error) {

	var status ci.Status

	log := logrus.WithField("testing", "collection-integrity")
	log.Info("Received Attest Request")

	var metadata [][]byte
	a := attestationreport.Generate(in.Nonce, metadata, []attestationreport.Measurement{})

	ok, data := attestationreport.Sign(a, signer)
	if !ok {
		log.Error("Failed to sign Attestion Report")
		status = ci.Status_FAIL
	} else {
		status = ci.Status_OK
	}

	response := &ci.AttestationResponse{
		Status:            status,
		AttestationReport: data,
	}

	return response, nil
}

func TestStartCollecting(t *testing.T) {
	type fields struct {
		streams  *clouditor_api.StreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence]
		grpcOpts []grpc.DialOption
	}
	type args struct {
		ctx context.Context
		req *collection.StartCollectingRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp assert.ValueAssertionFunc
		wantErr  bool
	}{
		{
			name: "Collect Success",
			fields: fields{
				streams:  clouditor_api.NewStreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](),
				grpcOpts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(testevaluation.BufConnDialer)},
			},
			args: args{
				ctx: context.Background(),
				req: &collection.StartCollectingRequest{
					ServiceId:   tcpAddr.String(),
					EvalManager: "bufnet",
					Configuration: &collection.ServiceConfiguration{
						RawConfiguration: testproto.NewAny(t, mockRawConfig()),
					},
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				resp, _ := i1.(*collection.StartCollectingResponse)
				assert.NotNil(t, resp)

				return assert.NotEmpty(t, resp.Id)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := Server{
				streams:  tt.fields.streams,
				grpcOpts: tt.fields.grpcOpts,
			}
			got, err := srv.StartCollecting(tt.args.ctx, tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			tt.wantResp(t, got)
		})
	}
}

func mockRawConfig() (conf *collection.RemoteIntegrityConfig) {
	conf = &collection.RemoteIntegrityConfig{
		Target:      tcpAddr.String(),
		Certificate: string(signer.certChain.Ca),
	}
	return
}

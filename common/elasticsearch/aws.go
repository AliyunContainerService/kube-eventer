// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package elasticsearch

import (
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	awsauth "github.com/smartystreets/go-aws-auth"
	"k8s.io/klog"
)

// AWSSigningTransport used to sign outgoing requests to AWS ES
type AWSSigningTransport struct {
	HTTPClient  *http.Client
	Credentials awsauth.Credentials
	Session     *session.Session
}

// RoundTrip implementation
func (a AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if a.Session.Config.Credentials.IsExpired() {
		a.newSession()
	}
	return a.HTTPClient.Do(awsauth.Sign4(req, a.Credentials))
}

func createAWSClient() (*http.Client, error) {
	signingTransport := AWSSigningTransport{HTTPClient: http.DefaultClient}
	signingTransport.newSession()

	return &http.Client{Transport: http.RoundTripper(signingTransport)}, nil
}

func useSigV4(opts url.Values) bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_ACCESS_KEY") != "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") != "" || os.Getenv("AWS_SECRET_KEY") != "" ||
		len(opts["sigv4"]) > 0
}

func (a *AWSSigningTransport) newSession() {
	newSession := session.Must(session.NewSession())
	credentials, err := newSession.Config.Credentials.Get()
	if err != nil {
		klog.Fatalf("error getting aws credentials: %v", err)
	}
	a.Session = newSession
	a.Credentials = awsauth.Credentials{
		AccessKeyID:     credentials.AccessKeyID,
		SecretAccessKey: credentials.SecretAccessKey,
		SecurityToken:   credentials.SessionToken,
	}

}

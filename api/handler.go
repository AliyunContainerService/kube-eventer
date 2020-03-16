// Copyright 2017 Google Inc. All Rights Reserved.
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

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AliyunContainerService/kube-eventer/manager"
	"k8s.io/klog"
)

const (
	// MaxEventsScrapeDelay should be larger than `frequency` command line argument.
	MaxEventsScrapeDelay = 3 * time.Minute
)

func healthzChecker(w http.ResponseWriter, r *http.Request) {
	if time.Since(manager.LatestScrapeTime) > MaxEventsScrapeDelay {
		msg := fmt.Sprintf(
			"No event batch within %s (latest: %s)",
			MaxEventsScrapeDelay,
			manager.LatestScrapeTime,
		)
		klog.Warning(msg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error: %s", msg)))

	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

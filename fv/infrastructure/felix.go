// Copyright (c) 2017-2018 Tigera, Inc. All rights reserved.
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

package infrastructure

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/felix/fv/containers"
	"github.com/projectcalico/felix/fv/utils"
)

type Felix struct {
	*containers.Container
}

func (f *Felix) GetFelixPID() int {
	return f.GetSinglePID("calico-felix")
}

func (f *Felix) GetFelixPIDs() []int {
	return f.GetPIDs("calico-felix")
}

func RunFelix(etcdIP string, options TopologyOptions) *Felix {
	log.Info("Starting felix")
	ipv6Enabled := fmt.Sprint(options.EnableIPv6)

	args := []string{
		"--privileged",
		"-e", "CALICO_DATASTORE_TYPE=etcdv3",
		"-e", "CALICO_ETCD_ENDPOINTS=http://" + etcdIP + ":2379",
		"-e", "FELIX_LOGSEVERITYSCREEN=" + options.FelixLogSeverity,
		"-e", "FELIX_DATASTORETYPE=etcdv3",
		"-e", "FELIX_PROMETHEUSMETRICSENABLED=true",
		"-e", "FELIX_USAGEREPORTINGENABLED=false",
		"-e", "FELIX_IPV6SUPPORT=" + ipv6Enabled,
		"-v", "/lib/modules:/lib/modules",
	}

	for k, v := range options.ExtraEnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range options.ExtraVolumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", k, v))
	}

	args = append(args,
		utils.Config.FelixImage,
	)

	c := containers.Run("felix",
		containers.RunOpts{AutoRemove: true},
		args...,
	)

	if options.EnableIPv6 {
		c.Exec("sysctl", "-w", "net.ipv6.conf.all.disable_ipv6=0")
		c.Exec("sysctl", "-w", "net.ipv6.conf.default.disable_ipv6=0")
		c.Exec("sysctl", "-w", "net.ipv6.conf.lo.disable_ipv6=0")
		c.Exec("sysctl", "-w", "net.ipv6.conf.all.forwarding=1")
	} else {
		c.Exec("sysctl", "-w", "net.ipv6.conf.all.disable_ipv6=1")
		c.Exec("sysctl", "-w", "net.ipv6.conf.default.disable_ipv6=1")
		c.Exec("sysctl", "-w", "net.ipv6.conf.lo.disable_ipv6=1")
		c.Exec("sysctl", "-w", "net.ipv6.conf.all.forwarding=0")
	}

	return &Felix{
		Container: c,
	}
}

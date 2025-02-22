/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prepare

import (
	"github.com/pkg/errors"
	"github.com/submariner-io/admiral/pkg/reporter"
	"github.com/submariner-io/cloud-prepare/pkg/api"
	"github.com/submariner-io/subctl/internal/restconfig"
	"github.com/submariner-io/subctl/pkg/cloud"
	"github.com/submariner-io/subctl/pkg/cloud/rhos"
)

func RHOS(restConfigProducer *restconfig.Producer, ports *cloud.Ports, config *rhos.Config, status reporter.Interface) error {
	gwPorts := []api.PortSpec{
		{Port: ports.Natt, Protocol: "udp"},
		{Port: ports.NatDiscovery, Protocol: "udp"},

		// ESP & AH protocols are used for private-ip to private-ip gateway communications
		{Port: 0, Protocol: "esp"},
		{Port: 0, Protocol: "ah"},
	}
	input := api.PrepareForSubmarinerInput{
		InternalPorts: []api.PortSpec{
			{Port: ports.Vxlan, Protocol: "udp"},
			{Port: ports.Metrics, Protocol: "tcp"},
		},
	}

	// nolint:wrapcheck // No need to wrap errors here.
	err := rhos.RunOn(restConfigProducer, config, status,
		func(cloud api.Cloud, gwDeployer api.GatewayDeployer, status reporter.Interface) error {
			if config.Gateways > 0 {
				gwInput := api.GatewayDeployInput{
					PublicPorts: gwPorts,
					Gateways:    config.Gateways,
				}

				err := gwDeployer.Deploy(gwInput, status)
				if err != nil {
					return errors.Wrap(err, "Deployment failed")
				}
			}

			return cloud.PrepareForSubmariner(input, status)
		})

	return status.Error(err, "Failed to prepare RHOS cloud")
}

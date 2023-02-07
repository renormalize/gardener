// Copyright (c) 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package botanist_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	kubernetesmock "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	"github.com/gardener/gardener/pkg/features"
	gardenletfeatures "github.com/gardener/gardener/pkg/gardenlet/features"
	"github.com/gardener/gardener/pkg/operation"
	. "github.com/gardener/gardener/pkg/operation/botanist"
	"github.com/gardener/gardener/pkg/operation/garden"
	shootpkg "github.com/gardener/gardener/pkg/operation/shoot"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/gardener/gardener/pkg/utils/test"
)

var _ = Describe("NginxIngress", func() {
	var (
		ctrl     *gomock.Controller
		botanist *Botanist
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		botanist = &Botanist{Operation: &operation.Operation{}}
		botanist.Shoot = &shootpkg.Shoot{}
		poilicy := corev1.ServiceExternalTrafficPolicyType(corev1.ServiceExternalTrafficPolicyTypeCluster)
		botanist.Shoot.SetInfo(&gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				Kubernetes: gardencorev1beta1.Kubernetes{
					Version: "1.22.1",
				},
				Addons: &gardencorev1beta1.Addons{
					NginxIngress: &gardencorev1beta1.NginxIngress{
						Config:                   map[string]string{},
						LoadBalancerSourceRanges: []string{},
						ExternalTrafficPolicy:    &poilicy,
					},
				},
			},
		})
		botanist.Garden = &garden.Garden{}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DefaultNginxIngress", func() {
		var kubernetesClient *kubernetesmock.MockInterface

		BeforeEach(func() {
			kubernetesClient = kubernetesmock.NewMockInterface(ctrl)

			botanist.SeedClientSet = kubernetesClient
		})

		It("should successfully create a nginxingress interface", func() {
			defer test.WithFeatureGate(gardenletfeatures.FeatureGate, features.APIServerSNI, true)()
			kubernetesClient.EXPECT().Client()
			botanist.ImageVector = imagevector.ImageVector{{Name: "nginx-ingress-controller"}, {Name: "ingress-default-backend"}}

			nginxIngress, err := botanist.DefaultNginxIngress()
			Expect(nginxIngress).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error because the controller image cannot be found", func() {
			botanist.ImageVector = imagevector.ImageVector{}

			nginxIngress, err := botanist.DefaultNginxIngress()
			Expect(nginxIngress).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error because the default backend image cannot be found", func() {
			botanist.ImageVector = imagevector.ImageVector{{Name: "nginx-ingress-controller"}}

			nginxIngress, err := botanist.DefaultNginxIngress()
			Expect(nginxIngress).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
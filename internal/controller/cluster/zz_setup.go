// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	providerconfig "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/providerconfig"
	apitoken "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/apitoken"
	billinghistory "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/billinghistory"
	billingnode "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/billingnode"
	configprofile "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/configprofile"
	externalsquad "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/externalsquad"
	host "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/host"
	hwiddevice "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/hwiddevice"
	infraprovider "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/infraprovider"
	internalsquad "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/internalsquad"
	node "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/node"
	nodemetadata "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/nodemetadata"
	nodeplugin "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/nodeplugin"
	panelsettings "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/panelsettings"
	passkey "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/passkey"
	snippet "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/snippet"
	subpageconfig "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/subpageconfig"
	subscriptionsettings "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/subscriptionsettings"
	subscriptiontemplate "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/subscriptiontemplate"
	user "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/user"
	usermetadata "github.com/STASiAN/crossplane-provider-remnawave/internal/controller/cluster/remnawave/usermetadata"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.Setup,
		apitoken.Setup,
		billinghistory.Setup,
		billingnode.Setup,
		configprofile.Setup,
		externalsquad.Setup,
		host.Setup,
		hwiddevice.Setup,
		infraprovider.Setup,
		internalsquad.Setup,
		node.Setup,
		nodemetadata.Setup,
		nodeplugin.Setup,
		panelsettings.Setup,
		passkey.Setup,
		snippet.Setup,
		subpageconfig.Setup,
		subscriptionsettings.Setup,
		subscriptiontemplate.Setup,
		user.Setup,
		usermetadata.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.SetupGated,
		apitoken.SetupGated,
		billinghistory.SetupGated,
		billingnode.SetupGated,
		configprofile.SetupGated,
		externalsquad.SetupGated,
		host.SetupGated,
		hwiddevice.SetupGated,
		infraprovider.SetupGated,
		internalsquad.SetupGated,
		node.SetupGated,
		nodemetadata.SetupGated,
		nodeplugin.SetupGated,
		panelsettings.SetupGated,
		passkey.SetupGated,
		snippet.SetupGated,
		subpageconfig.SetupGated,
		subscriptionsettings.SetupGated,
		subscriptiontemplate.SetupGated,
		user.SetupGated,
		usermetadata.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		providerconfig.SetupWebhookWithManager,
		apitoken.SetupWebhookWithManager,
		billinghistory.SetupWebhookWithManager,
		billingnode.SetupWebhookWithManager,
		configprofile.SetupWebhookWithManager,
		externalsquad.SetupWebhookWithManager,
		host.SetupWebhookWithManager,
		hwiddevice.SetupWebhookWithManager,
		infraprovider.SetupWebhookWithManager,
		internalsquad.SetupWebhookWithManager,
		node.SetupWebhookWithManager,
		nodemetadata.SetupWebhookWithManager,
		nodeplugin.SetupWebhookWithManager,
		panelsettings.SetupWebhookWithManager,
		passkey.SetupWebhookWithManager,
		snippet.SetupWebhookWithManager,
		subpageconfig.SetupWebhookWithManager,
		subscriptionsettings.SetupWebhookWithManager,
		subscriptiontemplate.SetupWebhookWithManager,
		user.SetupWebhookWithManager,
		usermetadata.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}

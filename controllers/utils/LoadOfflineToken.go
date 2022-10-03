package utils

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadOfflineToken(c client.Client, ctx context.Context, req ctrl.Request) (string, error) {
	secretName := types.NamespacedName{
		Namespace: "default",
		Name:      "sso.redhat.com",
	}
	var ssoRedHatComSecret v1.Secret
	if err := c.Get(ctx, secretName, &ssoRedHatComSecret); err != nil {
		return "", err
	}
	return string(ssoRedHatComSecret.Data["offlineToken"]), nil
}

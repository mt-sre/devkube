// Package devcluster allows performing operations on k8s clusters.
package devcluster

import (
	"context"
	"fmt"

	"github.com/mt-sre/devkube/devcheck"
	"github.com/mt-sre/devkube/devtime"

	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ReadynessChecks = map[string][]devcheck.ObjCheck

func DefaultReadynessCheckSetIfEmpty(cli client.Client, c ReadynessChecks, checker devcheck.Checker) ReadynessChecks {
	if len(c) != 0 {
		return c
	}

	checker = devcheck.RealCheckerIfNil(checker)

	return ReadynessChecks{
		"apps.Deployment": {
			checker.ObjCheckStatusConditionIs("Available", metav1.ConditionTrue),
		},
		"apiextensions.k8s.io.CustomResourceDefinition": {
			checker.ObjCheckStatusConditionIs("Established", metav1.ConditionTrue),
		},
	}
}

// Cluster represents a k8s cluster.
type Cluster struct {
	Cli             client.Client
	Checker         devcheck.Checker
	Poller          devtime.Poller
	ReadynessChecks map[string][]devcheck.ObjCheck
}

func (c Cluster) CreateAndAwaitReadiness(ctx context.Context, objs ...client.Object) error {
	checker := devcheck.RealCheckerIfNil(c.Checker)
	rcs := DefaultReadynessCheckSetIfEmpty(c.Cli, c.ReadynessChecks, c.Checker)

	checks := map[client.Object][]devcheck.ObjCheck{}
	for _, obj := range objs {
		err := c.Cli.Create(ctx, obj)
		switch {
		case err == nil:
		case apimachineryerrors.IsAlreadyExists(err):
		default:
			return err
		}

		gvk, err := apiutil.GVKForObject(obj, c.Cli.Scheme())
		if err != nil {
			return fmt.Errorf("could not determine GVK for object: %w", err)
		}

		if check, checkOK := rcs[gvk.GroupKind().String()]; checkOK {
			checks[obj] = check
		}
	}

	for obj, checks := range checks {
		if err := c.Poller.Wait(ctx, checker.CheckObj(c.Cli, obj, checks...)); err != nil {
			return err
		}
	}

	return nil
}

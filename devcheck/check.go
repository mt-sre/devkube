package devcheck

import (
	"context"
	"encoding/json"
	"fmt"

	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type (
	Check    = func(ctx context.Context) (success bool, err error)
	ObjCheck = func(obj client.Object) (success bool, err error)

	// Checker bundles a common set of Checks.
	Checker interface {
		CheckObj(cli client.Client, obj client.Object, checks ...ObjCheck) Check
		CheckThere(cli client.Client, obj client.Object) Check
		CheckGone(cli client.Client, obj client.Object) Check
		ObjCheckStatusConditionIs(conditionType string, conditionStatus metav1.ConditionStatus) ObjCheck
	}

	// RealChecker implements [Checker] with actual checks.
	RealChecker struct{}
)

// Waits for an object to report the given condition with given status.
// Takes observedGeneration into account when present on the object.
// observedGeneration may be reported on the condition or under .status.observedGeneration.
// Check if a object condition is in a certain state.
// Will respect .status.observedGeneration and .status.conditions[].observedGeneration.
func (r RealChecker) ObjCheckStatusConditionIs(conditionType string, conditionStatus metav1.ConditionStatus) ObjCheck {
	return func(obj client.Object) (bool, error) {
		var err error
		var unstrObj map[string]any

		switch v := any(obj).(type) {
		case *unstructured.Unstructured:
			unstrObj = v.Object
		default:
			unstrObj, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				return false, err
			}
		}

		observedGen, ok, err := unstructured.NestedInt64(unstrObj, "status", "observedGeneration")
		switch {
		case err != nil:
			return false, fmt.Errorf("could not access .status.observedGeneration: %w", err)
		case ok && observedGen != obj.GetGeneration():
			// Object status outdated
			return false, nil
		}

		conditionsRaw, ok, err := unstructured.NestedFieldNoCopy(unstrObj, "status", "conditions")
		switch {
		case err != nil:
			return false, fmt.Errorf("could not access .status.conditions: %w", err)
		case !ok:
			// no conditions reported
			return false, nil
		}

		// Press into metav1.Condition scheme to be able to work typed.
		conditionsJSON, err := json.Marshal(conditionsRaw)
		if err != nil {
			return false, fmt.Errorf("could not marshal conditions into JSON: %w", err)
		}
		var conditions []metav1.Condition
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			return false, fmt.Errorf("could not unmarshal conditions: %w", err)
		}

		// Check conditions
		condition := meta.FindStatusCondition(conditions, conditionType)
		switch {
		case condition == nil:
			// no such condition
			return false, nil
		case condition.ObservedGeneration != 0 && condition.ObservedGeneration != obj.GetGeneration():
			// Condition outdated
			return false, nil
		default:
			return condition.Status == conditionStatus, nil
		}
	}
}

func RealCheckerIfNil(checker Checker) Checker {
	if checker != nil {
		return checker
	}

	return RealChecker{}
}

func (r RealChecker) CheckGone(cli client.Client, obj client.Object) Check {
	return func(ctx context.Context) (bool, error) {
		err := cli.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		switch {
		case err == nil:
			return false, nil
		case apimachineryerrors.ReasonForError(err) == metav1.StatusReasonNotFound:
			return true, nil
		default:
			return false, err
		}
	}
}

func (r RealChecker) CheckThere(cli client.Client, obj client.Object) Check {
	return func(ctx context.Context) (bool, error) {
		err := cli.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		switch {
		case err == nil:
			return true, nil
		case apimachineryerrors.ReasonForError(err) == metav1.StatusReasonNotFound:
			return false, nil
		default:
			return false, err
		}
	}
}

func (r RealChecker) CheckObj(cli client.Client, obj client.Object, checks ...ObjCheck) Check {
	return func(ctx context.Context) (bool, error) {
		err := cli.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		switch {
		case err == nil:
			return true, nil
		case apimachineryerrors.ReasonForError(err) == metav1.StatusReasonNotFound:
			return false, nil
		}

		for _, check := range checks {
			success, err := check(obj)
			switch {
			case err != nil:
				return false, err
			case !success:
				return false, nil
			}
		}

		return true, nil
	}
}

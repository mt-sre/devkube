//nolint:revive
package devcheck

import (
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockChecker implements [Checker] with mock checks.
type MockChecker struct{ mock.Mock }

func (m *MockChecker) CheckObj(cli client.Client, obj client.Object, checks ...ObjCheck) Check {
	return m.Called(cli, obj, checks).Get(0).(Check)
}

func (m *MockChecker) CheckThere(cli client.Client, obj client.Object) Check {
	return m.Called(cli, obj).Get(0).(Check)
}

func (m *MockChecker) CheckGone(cli client.Client, obj client.Object) Check {
	return m.Called(cli, obj).Get(0).(Check)
}

func (m *MockChecker) ObjCheckStatusConditionIs(conditionType string, conditionStatus metav1.ConditionStatus) ObjCheck {
	return m.Called(conditionType, conditionStatus).Get(0).(ObjCheck)
}

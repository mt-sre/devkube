//nolint:revive
package devcluster

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CtrlCliMock struct{ mock.Mock }

func (c *CtrlCliMock) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	args := c.Called(obj)
	return args.Get(0).(schema.GroupVersionKind), args.Error(1)
}

func (c *CtrlCliMock) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	args := c.Called(obj)
	return args.Bool(0), args.Error(1)
}

func (c *CtrlCliMock) SubResource(name string) client.SubResourceClient {
	return c.Called(name).Get(0).(client.SubResourceClient)
}

func (c *CtrlCliMock) Status() client.StatusWriter {
	return c.Called().Get(0).(client.StatusWriter)
}

func (c *CtrlCliMock) Get(
	ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption,
) error {
	return c.Called(ctx, key, obj, opts).Error(0)
}

func (c *CtrlCliMock) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.Called(ctx, list, opts).Error(0)
}

func (c *CtrlCliMock) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return c.Called(ctx, obj, opts).Error(0)
}

func (c *CtrlCliMock) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *CtrlCliMock) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return c.Called(ctx, obj, opts).Error(0)
}

func (c *CtrlCliMock) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption,
) error {
	return c.Called(ctx, obj, patch, opts).Error(0)
}

func (c *CtrlCliMock) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return c.Called(ctx, obj, opts).Error(0)
}

func (c *CtrlCliMock) Scheme() *runtime.Scheme { return c.Called().Get(0).(*runtime.Scheme) }

func (c *CtrlCliMock) RESTMapper() meta.RESTMapper { return c.Called().Get(0).(meta.RESTMapper) }

type CtrlStatusClientMock struct{ mock.Mock }

func (c *CtrlStatusClientMock) Update(
	ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption,
) error {
	return c.Called(ctx, obj, opts).Error(0)
}

func (c *CtrlStatusClientMock) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption,
) error {
	return c.Called(ctx, obj, patch, opts).Error(0)
}

func (c *CtrlStatusClientMock) Create(
	ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption,
) error {
	return c.Called(ctx, obj, subResource, opts).Error(0)
}

type DynCliMock struct{ mock.Mock }

func (dc *DynCliMock) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return dc.Called(resource).Get(0).(dynamic.NamespaceableResourceInterface)
}

type DynCliResIface struct{ mock.Mock }

func (dc *DynCliResIface) Create(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) Update(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) UpdateStatus(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) Delete(
	ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string,
) error {
	args := dc.Called(ctx, name, options, subresources)
	return args.Error(0)
}

func (dc *DynCliResIface) DeleteCollection(
	ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions,
) error {
	args := dc.Called(ctx, options, listOptions)
	return args.Error(0)
}

func (dc *DynCliResIface) Get(
	ctx context.Context, name string, options metav1.GetOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	args := dc.Called(ctx, opts)
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}

func (dc *DynCliResIface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := dc.Called(ctx, opts)
	return args.Get(0).(watch.Interface), args.Error(1)
}

func (dc *DynCliResIface) Patch(
	ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, pt, data, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) Apply(
	ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynCliResIface) ApplyStatus(
	ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, obj, options)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

type DynClientNsResIface struct {
	mock.Mock
}

func (dc *DynClientNsResIface) Namespace(namespace string) dynamic.ResourceInterface {
	return dc.Called(namespace).Get(0).(dynamic.ResourceInterface)
}

func (dc *DynClientNsResIface) Create(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) Update(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) UpdateStatus(
	ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, obj, options)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) Delete(
	ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string,
) error {
	args := dc.Called(ctx, name, options, subresources)
	return args.Error(0)
}

func (dc *DynClientNsResIface) DeleteCollection(
	ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions,
) error {
	args := dc.Called(ctx, options, listOptions)
	return args.Error(0)
}

func (dc *DynClientNsResIface) Get(
	ctx context.Context, name string, options metav1.GetOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) List(
	ctx context.Context, opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	args := dc.Called(ctx, opts)
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}

func (dc *DynClientNsResIface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := dc.Called(ctx, opts)
	return args.Get(0).(watch.Interface), args.Error(1)
}

func (dc *DynClientNsResIface) Patch(
	ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, pt, data, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) Apply(
	ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (dc *DynClientNsResIface) ApplyStatus(
	ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions,
) (*unstructured.Unstructured, error) {
	args := dc.Called(ctx, name, obj, options)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

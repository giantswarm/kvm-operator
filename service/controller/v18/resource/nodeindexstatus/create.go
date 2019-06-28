package nodeindexstatus

import (
	"context"
	"reflect"
	"sort"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v18/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var nodeIndexes map[string]int
	{
		var idx int
		allocations := key.AllocatedNodeIndexes(cr)
		nodes := key.AllNodes(cr)
		nodeIndexes = copyMap(cr.Status.KVM.NodeIndexes)
		if nodeIndexes == nil {
			nodeIndexes = make(map[string]int)
		}

		// Remove node indexes for non-existent nodes.
		for nodeID := range nodeIndexes {
			found := false

			for _, n := range nodes {
				if n.ID == nodeID {
					found = true
					break
				}
			}

			if !found {
				delete(nodeIndexes, nodeID)
			}
		}

		// Ensure all present nodes have node index allocation.
		for _, n := range nodes {
			_, exists := nodeIndexes[n.ID]
			if !exists {
				allocations, idx = allocateIndex(allocations)
				nodeIndexes[n.ID] = idx
			}
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating status with node indexes")

		if !reflect.DeepEqual(cr.Status.KVM.NodeIndexes, nodeIndexes) {
			newObj, err := r.g8sClient.ProviderV1alpha1().KVMConfigs(cr.GetNamespace()).Get(cr.GetName(), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			newObj.Status.KVM.NodeIndexes = nodeIndexes
			_, err = r.g8sClient.ProviderV1alpha1().KVMConfigs(newObj.GetNamespace()).UpdateStatus(newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "updated status with node indexes")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)

			return nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not update status with node indexes")
		}
	}

	return nil
}

// allocateIndex takes existing index allocations as parameter and returns next
// lowest number available. If there is a "hole" in the indexes, that is used.
//
// Examples:
//	In: []
//	Out: [1], 1
//
//	In: [1,2,3]
//  Out: [1,2,3,4], 4
//
//	In: [1,2,4]
//	Out: [1,2,3,4], 3
//
func allocateIndex(indexes []int) ([]int, int) {
	if len(indexes) == 0 {
		return []int{1}, 1
	}

	sort.Ints(indexes)

	for i, v := range indexes {

		// Empty slot in the middle of allocated indexes?
		if (i + 1) < v {
			idx := i + 1

			// Insert allocated index by first extending the slice...
			indexes = append(indexes, 0)
			// ...then moving rest of the elements by one...
			copy(indexes[i+1:], indexes[i:])
			// ...and finally setting the new index at right place.
			indexes[i] = idx

			return indexes, idx
		}
	}

	idx := len(indexes) + 1

	return append(indexes, idx), idx
}

func copyMap(v map[string]int) map[string]int {
	m := make(map[string]int)
	for k, v := range v {
		m[k] = v
	}
	return m
}

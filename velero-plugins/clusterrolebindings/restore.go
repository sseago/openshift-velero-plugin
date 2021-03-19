package clusterrolebindings

import (
	"encoding/json"

	"github.com/konveyor/openshift-velero-plugin/velero-plugins/rolebindings"
	apiauthorization "github.com/openshift/api/authorization/v1"
	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to PVCs
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"clusterrolebinding.authorization.openshift.io"},
	}, nil
}

// Execute action for the restore plugin for the pvc resource
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Info("[clusterrolebindings-restore] Entering Cluster Role Bindings restore plugin")

	clusterRoleBinding := apiauthorization.ClusterRoleBinding{}
	itemMarshal, _ := json.Marshal(input.Item)
	json.Unmarshal(itemMarshal, &clusterRoleBinding)

	p.Log.Infof("[clusterrolebindings-restore] role binding - %s, API version", clusterRoleBinding.Name, clusterRoleBinding.APIVersion)

	namespaceMapping := input.Restore.Spec.NamespaceMapping
	if len(namespaceMapping) > 0 {
		newRoleRefNamespace := namespaceMapping[clusterRoleBinding.RoleRef.Namespace]
		if newRoleRefNamespace != "" {
			clusterRoleBinding.RoleRef.Namespace = newRoleRefNamespace
		}

		clusterRoleBinding.Subjects = rolebindings.SwapSubjectNamespaces(clusterRoleBinding.Subjects, namespaceMapping)
		clusterRoleBinding.UserNames = rolebindings.SwapUserNamesNamespaces(clusterRoleBinding.UserNames, namespaceMapping)
		clusterRoleBinding.GroupNames = rolebindings.SwapGroupNamesNamespaces(clusterRoleBinding.GroupNames, namespaceMapping)
	}

	var out map[string]interface{}
	objrec, _ := json.Marshal(clusterRoleBinding)
	json.Unmarshal(objrec, &out)

	return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: out}), nil
}

// This plugin doesn't need to wait for items
func (p *RestorePlugin) AreAdditionalItemsReady(restore *v1.Restore, additionalItems []velero.ResourceIdentifier) (bool, error) {
	return true, nil
}

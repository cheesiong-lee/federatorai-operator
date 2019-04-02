package component

import (
	"crypto/x509"
	"fmt"
	"net"
	"strings"

	"github.com/containers-ai/federatorai-operator/pkg/assets"
	"github.com/containers-ai/federatorai-operator/pkg/lib/resourceread"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/util/cert"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_alamedaservice")

type ComponentConfig struct {
	NameSpace string
}

func NewComponentConfig(ns string) *ComponentConfig {
	return &ComponentConfig{
		NameSpace: ns,
	}
}

func (c *ComponentConfig) SetNameSpace(ns string) {
	c.NameSpace = ns
}

func (c ComponentConfig) NewClusterRoleBinding(str string) *rbacv1.ClusterRoleBinding {
	crbByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create clusterrolebinding")

	}
	crb := resourceread.ReadClusterRoleBindingV1(crbByte)
	crb.Name = strings.Replace(crb.Name, strings.Split(crb.Name, "-")[0], c.NameSpace, -1)
	crb.Subjects[0].Namespace = c.NameSpace
	return crb
}
func (c ComponentConfig) NewClusterRole(str string) *rbacv1.ClusterRole {
	crByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create clusterrole")
	}
	cr := resourceread.ReadClusterRoleV1(crByte)
	return cr
}
func (c ComponentConfig) NewServiceAccount(str string) *corev1.ServiceAccount {
	saByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create serviceaccount")

	}
	sa := resourceread.ReadServiceAccountV1(saByte)
	sa.Namespace = c.NameSpace
	return sa
}
func (c ComponentConfig) NewConfigMap(str string) *corev1.ConfigMap {
	cmByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create configmap")

	}
	cm := resourceread.ReadConfigMapV1(cmByte)
	cm.Namespace = c.NameSpace
	return cm
}
func (c ComponentConfig) NewPersistentVolumeClaim(str string) *corev1.PersistentVolumeClaim {
	pvcByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create persistentvolumeclaim")

	}
	pvc := resourceread.ReadPersistentVolumeClaimV1(pvcByte)
	pvc.Namespace = c.NameSpace
	return pvc
}
func (c ComponentConfig) NewService(str string) *corev1.Service {
	svByte, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create service")

	}
	sv := resourceread.ReadServiceV1(svByte)
	sv.Namespace = c.NameSpace
	return sv
}
func (c ComponentConfig) NewDeployment(str string) *appsv1.Deployment {
	deploymentBytes, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create deployment")

	}
	d := resourceread.ReadDeploymentV1(deploymentBytes)
	d.Namespace = c.NameSpace
	return d
}

func (c ComponentConfig) NewAdmissionControllerSecret() (*corev1.Secret, error) {

	secret, err := c.NewSecret("Secret/admission-controller-tls.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to buiild admission-controller secret")
	}
	secret.Namespace = c.NameSpace

	caKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "new ca private key failed")
	}

	caCertCfg := cert.Config{}
	caCert, err := cert.NewSelfSignedCACert(caCertCfg, caKey)
	if err != nil {
		return nil, errors.Wrap(err, "new ca cert failed")
	}

	admctlKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "new admctl private key failed")
	}

	admctlCertCfg := cert.Config{
		CommonName: fmt.Sprintf("admission-controller.%s.svc", c.NameSpace),
		Usages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
	}
	admctlCert, err := cert.NewSignedCert(admctlCertCfg, admctlKey, caCert, caKey)
	if err != nil {
		return nil, errors.Wrap(err, "new admctl cert failed")
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["ca.crt"] = cert.EncodeCertPEM(caCert)
	secret.Data["tls.crt"] = cert.EncodeCertPEM(admctlCert)
	secret.Data["tls.key"] = cert.EncodePrivateKeyPEM(admctlKey)

	return secret, nil
}

func (c ComponentConfig) NewInfluxDBSecret() (*corev1.Secret, error) {

	secret, err := c.NewSecret("Secret/alameda-influxdb.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to buiild influxdb secret")
	}
	secret.Namespace = c.NameSpace

	host := fmt.Sprintf("admission-influxdb.%s.svc", c.NameSpace)
	crt, key, err := cert.GenerateSelfSignedCertKey(host, []net.IP{}, []string{})
	if err != nil {
		return nil, errors.Errorf("failed to buiild influxdb secret: %s", err.Error())
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["tls.crt"] = crt
	secret.Data["tls.key"] = key

	return secret, nil
}

func (c ComponentConfig) NewSecret(str string) (*corev1.Secret, error) {
	secretBytes, err := assets.Asset(str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build secret from assets' bin data")
	}
	s, err := resourceread.ReadSecretV1(secretBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build secret from assets' bin data")
	}
	s.Namespace = c.NameSpace
	return s, nil
}

func (c ComponentConfig) RegistryCustomResourceDefinition(str string) *apiextv1beta1.CustomResourceDefinition {
	crdBytes, err := assets.Asset(str)
	if err != nil {
		log.Error(err, "Failed to Test create testcrd")
	}
	crd := resourceread.ReadCustomResourceDefinitionV1Beta1(crdBytes)
	return crd
}

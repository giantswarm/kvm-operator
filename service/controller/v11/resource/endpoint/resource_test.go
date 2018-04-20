package endpoint

import (
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoint_cutIPs(t *testing.T) {
	testCases := []struct {
		name         string
		base         []string
		cutset       []string
		expectedSet  []string
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: Base has one more address than the cutset",
			base: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet: []string{
				"0.0.0.0",
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: Base and cutset have the same addresses",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
			},
			expectedSet:  []string{},
			errorMatcher: nil,
		},
		{
			name: "case 2: Cutset has one more address than base",
			base: []string{
				"1.1.1.1",
			},
			cutset: []string{
				"1.1.1.1",
				"0.0.0.0",
			},
			expectedSet:  []string{},
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			resultSet := cutIPs(tc.base, tc.cutset)

			if !reflect.DeepEqual(resultSet, tc.expectedSet) {
				t.Fatalf("resultSet == %q, want %q", resultSet, tc.expectedSet)
			}
		})
	}
}

func Test_Resource_Endpoint_newK8sEndpoint(t *testing.T) {
	testCases := []struct {
		name                string
		setupService        *corev1.Service
		endpoint            *Endpoint
		expectedK8sEndpoint *corev1.Endpoints
		errorMatcher        func(error) bool
	}{
		{
			name: "case 0: Without adresses, empty k8sEndpoint desired",
			setupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{},
				},
			},
			endpoint: &Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			expectedK8sEndpoint: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: With addresses, no ports",
			setupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{},
				},
			},
			endpoint: &Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
				IPs: []string{
					"1.2.3.4",
				},
			},
			expectedK8sEndpoint: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: With addresses and ports",
			setupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 22,
						},
					},
				},
			},
			endpoint: &Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
				IPs: []string{
					"1.2.3.4",
				},
			},
			expectedK8sEndpoint: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
						},
						Ports: []corev1.EndpointPort{
							{
								Port: 22,
							},
						},
					},
				},
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: Without addresses but ports",
			setupService: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 22,
						},
					},
				},
			},
			endpoint: &Endpoint{
				ServiceName:      "TestService",
				ServiceNamespace: "TestNamespace",
			},
			expectedK8sEndpoint: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestService",
					Namespace: "TestNamespace",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{},
						Ports: []corev1.EndpointPort{
							{
								Port: 22,
							},
						},
					},
				},
			},
			errorMatcher: nil,
		},
	}
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := New(Config{
				K8sClient: clientgofake.NewSimpleClientset(),
				Logger:    logger,
			})
			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}
			if tc.setupService != nil {
				if _, err := r.k8sClient.CoreV1().Services(tc.setupService.Namespace).Create(tc.setupService); err != nil {
					t.Fatalf(" error returned setting up k8s service: %s\n", err)
				}
			}

			k8sEndpoints, err := r.newK8sEndpoint(tc.endpoint)

			switch {
			case err == nil && tc.errorMatcher == nil: // correct; carry on
			case err != nil && tc.errorMatcher != nil:
				if !tc.errorMatcher(err) {
					t.Fatalf("error == %#v, want matching", err)
				}
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			}

			if !reflect.DeepEqual(k8sEndpoints, tc.expectedK8sEndpoint) {
				t.Fatalf("k8sEndpoint == %q, want %q", k8sEndpoints, tc.expectedK8sEndpoint)
			}
		})
	}
}

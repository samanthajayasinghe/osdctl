package k8s

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	awsv1alpha1 "github.com/openshift/aws-account-operator/pkg/apis/aws/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	awsprovider "github.com/openshift/osd-utils-cli/pkg/provider/aws"
)

func TestGetAWSAccount(t *testing.T) {
	_ = awsv1alpha1.AddToScheme(scheme.Scheme)
	g := NewGomegaWithT(t)
	testCases := []struct {
		title           string
		localObjects    []runtime.Object
		namespace       string
		accountName     string
		expectedAccount awsv1alpha1.Account
		errExpected     bool
		errReason       metav1.StatusReason
	}{
		{
			title:        "not found account",
			localObjects: nil,
			namespace:    "foo",
			accountName:  "bar",
			errExpected:  true,
			errReason:    metav1.StatusReasonNotFound,
		},
		{
			title: "success",
			localObjects: []runtime.Object{
				&awsv1alpha1.Account{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar",
						Namespace: "foo",
					},
				},
			},
			namespace:   "foo",
			accountName: "bar",
			errExpected: false,
			expectedAccount: awsv1alpha1.Account{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Account",
					APIVersion: "aws.managed.openshift.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.TODO()
			client := fake.NewFakeClientWithScheme(scheme.Scheme, tc.localObjects...)
			account, err := GetAWSAccount(ctx, client, tc.namespace, tc.accountName)
			if tc.errExpected {
				g.Expect(err).Should(HaveOccurred())
				if tc.errReason != "" {
					g.Expect(tc.errReason).Should(Equal(apierrors.ReasonForError(err)))
				}
			} else {
				g.Expect(tc.expectedAccount).Should(Equal(*account))
			}
		})
	}
}

func TestGetAWSAccountCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	testCases := []struct {
		title        string
		localObjects []runtime.Object
		namespace    string
		secretName   string
		credentials  awsprovider.AwsClientInput
		errExpected  bool
		errReason    metav1.StatusReason
	}{
		{
			title:        "no secret found",
			localObjects: nil,
			namespace:    "foo",
			secretName:   "bar",
			errExpected:  true,
			errReason:    metav1.StatusReasonNotFound,
		},
		{
			title: "found secret but no AWS credentials",
			localObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "foo",
						Name:      "bar",
					},
				},
			},
			namespace:   "foo",
			secretName:  "bar",
			errExpected: true,
		},
		{
			title: "found secret but invalid credentials",
			localObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "foo",
						Name:      "bar",
					},
					Data: map[string][]byte{
						"aws_access_key_id": []byte("foo"),
					},
				},
			},
			namespace:   "foo",
			secretName:  "bar",
			errExpected: true,
		},
		{
			title: "success",
			localObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "foo",
						Name:      "bar",
					},
					Data: map[string][]byte{
						"aws_access_key_id":     []byte("foo"),
						"aws_secret_access_key": []byte("bar"),
					},
				},
			},
			namespace:   "foo",
			secretName:  "bar",
			errExpected: false,
			credentials: awsprovider.AwsClientInput{
				AwsIDKey:     "foo",
				AwsAccessKey: "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.TODO()
			client := fake.NewFakeClientWithScheme(scheme.Scheme, tc.localObjects...)
			creds, err := GetAWSAccountCredentials(ctx, client, tc.namespace, tc.secretName)
			if tc.errExpected {
				g.Expect(err).Should(HaveOccurred())
				if tc.errReason != "" {
					g.Expect(tc.errReason).Should(Equal(apierrors.ReasonForError(err)))
				}
			} else {
				g.Expect(tc.credentials).Should(Equal(*creds))
			}
		})
	}
}
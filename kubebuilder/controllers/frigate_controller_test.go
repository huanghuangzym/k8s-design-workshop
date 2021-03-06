package controllers

import (
	"context"
	shipv1beta1 "github.com/danielfbm/k8s-design-workshop/controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

/*
In TDD it is generally recommended to not be conserned with implementation
but focus on result, in this controller test case we can define our input (CRD instance)
and focus on the end result.

To simplify the business logic we will just add a Phase "Completed" to the CRD instance
*/
var _ = Describe("Reconcile", func() {

	var (
		// variable used in the test or configuration for tests
		frigate    *shipv1beta1.Frigate
		result     *shipv1beta1.Frigate
		controller *FrigateReconciler
		manager    ctrl.Manager

		opts mgr.Options
		ctx  context.Context

		config    *rest.Config
		k8sclient client.Client
		err       error
		stop      chan struct{}
	)

	// Ginkgo framework is based around a few blocks:
	// Describe, Context, BeforeEach, JustBeforeEach, It, JustAfterEach, AfterEach
	// being that for each Describe/Context every time a function is declared it will be used for each It
	// BeforeEach is generally used for initialization
	// combined with a JustBeforeEach that can be used to run the specific test
	// leaving It to only run specific validations
	BeforeEach(func() {
		// Basic initialization
		// cfg  and k8sClient variables declared on suite_test.go
		config = cfg
		k8sclient = k8sClient
		stop = make(chan struct{})
		ctx = context.TODO()

		// Create and start manager
		manager, err = ctrl.NewManager(config, opts)
		Expect(err).ToNot(HaveOccurred(), "building manager")
		go func() {
			Expect(manager.Start(stop)).ToNot(HaveOccurred(), "starting manager")
		}()

		// Create controller
		controller = &FrigateReconciler{Log: logf.Log}
		err = controller.SetupWithManager(manager)
		Expect(err).ToNot(HaveOccurred(), "building controller")

		// Base data input (can be overwritten, example bellow)
		frigate = &shipv1beta1.Frigate{
			ObjectMeta: metav1.ObjectMeta{Name: "some", Namespace: "default"},
			Spec:       shipv1beta1.FrigateSpec{Foo: "foo"},
		}
	})

	// Here are the steps we take for every test case
	// for this case:
	// 1. create resource (resource data can be overwritten)
	// 2. wait for reconcile loop and keep result in result and err variables
	JustBeforeEach(func() {
		// create resource
		err = k8sclient.Create(ctx, frigate)
		Expect(err).To(BeNil(), "create frigate instance")

		objKey := client.ObjectKey{Namespace: frigate.Namespace, Name: frigate.Name}

		// wait for result
		// for this specific case we can validate the phase but
		// each controller might have a different way to validate
		// when does the reconcile loop finishes
		// For more on Eventually workings: http://onsi.github.io/gomega/
		result = &shipv1beta1.Frigate{}
		Eventually(func() string {
			err = k8sclient.Get(ctx, objKey, result)
			logf.Log.Info("got?", "result", result, "err", err)
			return result.Status.Phase
		}, time.Second).ShouldNot(BeEmpty())
	})

	// Some cleanup tasks between each test case
	AfterEach(func() {
		k8sclient.Delete(ctx, frigate)
		close(stop)
	})

	// This is the specific test case
	// here we will use the default data and variable set in BeforeEach
	// and can validate the result directly
	It("should have a Completed phase", func() {
		Expect(result).ToNot(BeNil(), "should have a result")
		Expect(result.Status.Phase).To(Equal("Completed"))
	})

	// How to reuse all the above code and add a new test case?
	// context can make it happen
	Context("new frigate instance with empty Foo", func() {
		// Adding this method will add a new BeforeEach to be executed
		// right after the one executed on top for this Context
		BeforeEach(func() {
			// lets say for the sake of simplicity that "another" Frigate
			// should have a "Failure" phase
			frigate = &shipv1beta1.Frigate{
				ObjectMeta: metav1.ObjectMeta{Name: "another", Namespace: "default"},
				Spec:       shipv1beta1.FrigateSpec{Foo: ""},
			}
		})

		It("should have a Failure phase", func() {
			Expect(result).ToNot(BeNil(), "should have a result")
			Expect(result.Status.Phase).To(Equal("Failure"))
		})
	})
})

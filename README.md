Create a digital ocean kubernetes cluster using operator.

type name -> Kluster
group -> utsab.dev
version -> v1alpha1

The directory will look like
pkg/apis/utsab.dev/v1alpha1 --> inside with there should be types.gos

create main.go in the outside directory

Since the structs created in types.go are only the go structs
and not necessarily a kubernetes object
--> Hence to convert the go struct to kubernetes object,
    we use register.go file.

Generate
1. Deep copy objects
2. ClientSet
3. Informers
4. Listers
--> We can use codegenerator project for this.
---> doc.go file
--> Types are the ways to control the behaviour of the code generator
    - global tags -- specified in doc.go
    - local tags -- specify tags in types.

Even though the struct is registered in Kubernetes,
CRD should be generated, so that kubernetes can access the struct.
--> use controller-gen
(Registering and creating CRD should go in parallel)
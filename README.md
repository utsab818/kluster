Create a digital ocean kubernetes cluster using operator.

type name -> Kluster
group -> utsab.dev
version -> v1alpha1

The directory will look like
pkg/apis/utsab.dev/v1alpha1 --> inside with there should be types.go

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


// controller
create folder as controller inside pkg folder
create a file kluster.go

call the digitalocean api using Spec
create do->digitalocean folder under pkg and create a file 



We want to add other fields to the file
by itself using controller
like configuring clusterID, progress, kubeconfig on the yaml file
without explicitly added by the user.
For this status subresource is useful.

apiVersion
kind
metadata:
spec: // only should be updated by user
    ...
    ...
status: // only should be updated by controller or operator
    clusterID
    progress
    kubeconfig

To obtain this, we use subresource
making status and spec as the subresource gives an endpoint
to which can be mapped a RBAC.

Every resource has an endpoint
resource
    apis/apps/v1/namespaces/<ns>/deployments
    v1/namespaces/<ns>/pods/<podname>/

Subresources
    pods --> logs, describe, exec

Endpoint for the subresource
    v1/namespaces/<ns>/pods/<podname>/logs


Suppose this is the role by which the operator is going to run
role-test:
    resources: Kluster/status

-->This particular role will only allow operator to update the status part.


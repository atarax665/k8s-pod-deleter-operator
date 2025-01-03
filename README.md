# Pod Lifetime Operator

The **Pod Lifetime Operator** is a Kubernetes operator designed to automatically delete pods after a specified lifetime. This helps in efficiently managing resources and ensuring the timely removal of pods that are no longer needed. The lifetime of a pod is defined using a custom label applied to the pod.

## Features

- **Automated Deletion**: Automatically deletes pods after their defined lifetime expires.
- **Periodic Checks**: Regularly scans for expired pods and removes them.
- **Configurable Recheck Interval**: Allows customization of the recheck interval to suit your requirements.

---

## Prerequisites

To use the Pod Lifetime Operator, ensure you have the following:

- A running Kubernetes cluster.
- `kubectl` configured to interact with your cluster.
- Operator SDK (optional, for development purposes).

---

## Installation

Follow these steps to install the Pod Lifetime Operator:

1. **Clone the Repository**:

   ```sh
   git clone https://github.com/atarax665/pod-lifetime-operator.git
   cd pod-lifetime-operator
   ```

2. **Apply RBAC Configuration**:

   The operator requires specific permissions to manage pods. Apply the necessary Role-Based Access Control (RBAC) configurations:

   ```sh
   kubectl apply -f config/rbac/role.yaml
   kubectl apply -f config/rbac/role_binding.yaml
   ```

3. **Deploy the Operator**:

   Deploy the Pod Lifetime Operator in your cluster:

   ```sh
   kubectl apply -f config/crd/bases/podlifetime.atarax.local_poddeleters.yaml
   ```

---

## Usage

To use the Pod Lifetime Operator, simply add the `pod.kubernetes.io/lifetime` label to your pods. The value of this label should represent the lifetime of the pod in seconds.

### Example

Here is an example of a pod configuration with a lifetime label:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    pod.kubernetes.io/lifetime: "60" # Lifetime in seconds
spec:
  containers:
    - name: busybox
      image: busybox
      command: ["sh", "-c", "sleep 3600"]
```

In this example, the pod will automatically be deleted 60 seconds after it is created.

---

## Development

### Prerequisites

To contribute to the development of the Pod Lifetime Operator, ensure you have:

- **Go**: Version 1.16 or later.
- **Docker**: For building container images.
- **Kubernetes Cluster**: A local cluster such as Minikube is recommended for testing.

### Running Locally

1. **Install Dependencies**:

   Ensure all dependencies are installed by running:

   ```sh
   go mod tidy
   ```

2. **Run the Operator Locally**:

   Use the following command to run the operator locally:

   ```sh
   make run
   ```

## Contributing

We welcome contributions to the Pod Lifetime Operator! If you have ideas, issues, or improvements, please open an issue or submit a pull request.

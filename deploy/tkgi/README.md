https://github.com/wavefrontHQ/wavefront-kubernetes-collector

```
kubectl create ns pks-system
kubectl create secret generic wavefront-secret -n pks-system --from-literal=wavefront-token=XXX

kubectl apply -f .
kubectl get secret

kubectl get all -n pks-system
```

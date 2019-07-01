# Debug tool

We want to test to see if an installation is in a good state.

```bash
kubectl get deployments -ojsonpath='{.items[*].status.conditions[?(@.type=="Available")].status}'
```

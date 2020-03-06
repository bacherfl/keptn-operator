# Updating k8s types:
```
operator-sdk generate k8s
operator-sdk generate crds
```


# DEBUGGING:
```
export WATCH_NAMESPACE=keptn
operator-sdk run --local --namespace=keptn --enable-delve
```

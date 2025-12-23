1. `helm repo add prometheus-community https://prometheus-community.github.io/helm-charts`
2. `helm repo update``
3. `helm install monitoring prometheus-community/kube-prometheus-stack`
Output:
```
NAME: monitoring
LAST DEPLOYED: Tue Dec 23 19:17:12 2025
NAMESPACE: default
STATUS: deployed
REVISION: 1
DESCRIPTION: Install complete
NOTES:
kube-prometheus-stack has been installed. Check its status by running:
  kubectl --namespace default get pods -l "release=monitoring"

Get Grafana 'admin' user password by running:

  kubectl --namespace default get secrets monitoring-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo

Access Grafana local instance:

  export POD_NAME=$(kubectl --namespace default get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=monitoring" -oname)
  kubectl --namespace default port-forward $POD_NAME 3000

Get your grafana admin user password by running:

  kubectl get secret --namespace default -l app.kubernetes.io/component=admin-secret -o jsonpath="{.items[0].data.admin-password}" | base64 --decode ; echo


```
4. Verify using: `kubectl get pods`
5. Verify services : `kubectl get svc`
```
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   16m
kubernetes                                ClusterIP   10.96.0.1        <none>        443/TCP                      5d5h
monitoring-grafana                        ClusterIP   10.96.171.127    <none>        80/TCP                       16m
monitoring-kube-prometheus-alertmanager   ClusterIP   10.101.86.207    <none>        9093/TCP,8080/TCP            16m
monitoring-kube-prometheus-operator       ClusterIP   10.103.151.242   <none>        443/TCP                      16m
monitoring-kube-prometheus-prometheus     ClusterIP   10.97.42.209     <none>        9090/TCP,8080/TCP            16m
monitoring-kube-state-metrics             ClusterIP   10.109.254.37    <none>        8080/TCP                     16m
monitoring-prometheus-node-exporter       ClusterIP   10.103.144.205   <none>        9100/TCP                     16m
prometheus-operated                       ClusterIP   None             <none>        9090/TCP                     16m

```

6. We'll be converting ClusterIP `prometheus-operated` service to NodePort service:

`kubectl expose service prometheus-operated --type=NodePort --target-port=9090 --name=prometheus-operated-ext` 
```
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                     ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   19m
kubernetes                                ClusterIP   10.96.0.1        <none>        443/TCP                      5d5h
monitoring-grafana                        ClusterIP   10.96.171.127    <none>        80/TCP                       19m
monitoring-kube-prometheus-alertmanager   ClusterIP   10.101.86.207    <none>        9093/TCP,8080/TCP            19m
monitoring-kube-prometheus-operator       ClusterIP   10.103.151.242   <none>        443/TCP                      19m
monitoring-kube-prometheus-prometheus     ClusterIP   10.97.42.209     <none>        9090/TCP,8080/TCP            19m
monitoring-kube-state-metrics             ClusterIP   10.109.254.37    <none>        8080/TCP                     19m
monitoring-prometheus-node-exporter       ClusterIP   10.103.144.205   <none>        9100/TCP                     19m
prometheus-operated                       ClusterIP   None             <none>        9090/TCP                     19m
prometheus-operated-ext                   NodePort    10.102.94.169    <none>        9090:31574/TCP               4
```

7. For minikube setup `minikube ip`
Execute on browser and use prometheus-operated-ext port : `192.168.64.3:31574`


---

# Installing Grafana

Doc Reference: https://grafana.com/docs/grafana/latest/setup-grafana/installation/helm/

1. `helm repo add grafana https://grafana.github.io/helm-charts`
2. `helm repo update`
3. `helm install grafana grafana/grafana`
```
NOTES:
1. Get your 'admin' user password by running:

   kubectl get secret --namespace default grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo


2. The Grafana server can be accessed via port 80 on the following DNS name from within your cluster:

   grafana.default.svc.cluster.local

   Get the Grafana URL to visit by running these commands in the same shell:
     export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=grafana" -o jsonpath="{.items[0].metadata.name}")
     kubectl --namespace default port-forward $POD_NAME 3000

3. Login with the password from step 1 and the username: admin
#################################################################################
######   WARNING: Persistence is disabled!!! You will lose your data when   #####
######            the Grafana pod is terminated.                            #####
#################################################################################

```

To get password execute `kubectl get secret --namespace default grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo`

3. Perform `kubectl get svc` and create additional grafana svc exposed using NodePort

```
infinitelearner@Vibhors-MacBook-Air:~/Code-Repos/Go-Sight (master *)$ kubectl get svc
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
grafana                                   ClusterIP   10.104.116.167   <none>        80/TCP                       3m10s
```

Execute: `kubectl expose service monitoring-grafana --type=NodePort --target-port=80 --name=monitoring-grafana-ext`
Then patch to use correct port: `kubectl patch svc monitoring-grafana-ext -p '{"spec":{"ports":[{"port":80,"targetPort":3000,"protocol":"TCP"}]}}'`

We've 2 grafana:

http://192.168.64.3:30935 (standalone Grafana)
http://192.168.64.3:30531 (kube-prometheus-stack Grafana)

```
NAME                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
grafana                                   ClusterIP   10.104.116.167   <none>        80/TCP                       11m
monitoring-grafana                        ClusterIP   10.96.171.127    <none>        80/TCP                       53

```

4. `minikube ip`

**Commands to expose Grafana using NodePort:**

For standalone Grafana:
```bash
kubectl expose service grafana --type=NodePort --target-port=80 --name=grafana-ext
kubectl patch svc grafana-ext -p '{"spec":{"ports":[{"port":80,"targetPort":3000,"protocol":"TCP"}]}}'
```

For kube-prometheus-stack Grafana:
```bash
kubectl expose service monitoring-grafana --type=NodePort --target-port=80 --name=monitoring-grafana-ext
kubectl patch svc monitoring-grafana-ext -p '{"spec":{"ports":[{"port":80,"targetPort":3000,"protocol":"TCP"}]}}'
```

**Note:** The `--target-port=80` creates the service, but Grafana actually listens on port 3000 inside the container. The patch command fixes the target port to 3000. Alternatively, you can directly target port 3000 when exposing, but the ClusterIP service uses port 80, so this approach maintains consistency.

**Access Grafana:**
- Get minikube IP: `minikube ip`
- Access via: `http://<minikube-ip>:<NodePort>` (check NodePort with `kubectl get svc grafana-ext monitoring-grafana-ext`)



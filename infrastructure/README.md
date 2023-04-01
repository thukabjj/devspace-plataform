# Infrastructure

Should apply first the namespaces using the following commands:

```bash	
kubectl apply -f namespaces/
```

After applying is just needed do the same to the root yaml:

```bash	
kubectl apply -f .
```

After applying will be needed, if you want to access the resources of Kibana and Jenkins make a port foward command:

For Jenkins:
```bash
 kubectl port-forward svc/jenkins 8080:8080 -n jenkins
``` 
if it's your first deployment of Jenkins you will need to get the admin password from the pod to access the resources and install the required plugins. To get the admin password run the following commands:

Get the pods from jenkins namespace:

```bash

 kubectl get pods -n jenkins

``` 

Get logs from jenkins pod inside the namespace:
```bash

 kubectl logs <jenkins-pod-name> -n jenkins

``` 
Get the admin password from the pod inside the namespace:
```bash
 kubectl exec <jenkins-pod-name> -n jenkins cat /var/jenkins_home/secrets/initialAdminPassword

``` 

For ELK Stack:
```bash
 kubectl port-forward service/kibana 5601:5601 -n logging
``` 


For delete every pod inside the namespace using the following commands:


ELK Stack:
```bash
kubectl delete pods --all -n logging
``` 

Jenkins Stack:
```bash
kubectl delete pods --all -n jenkins
``` 
